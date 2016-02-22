package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/robinmonjo/dock/notifier"
	"github.com/robinmonjo/dock/port"
	"github.com/robinmonjo/procfs"

	"github.com/robinmonjo/dock/iowire"
	_ "github.com/robinmonjo/dock/logrotate"
)

var (
	version string //injected by the makefile
)

func main() {
	app := cli.NewApp()
	app.Name = "dock"
	app.Version = fmt.Sprintf("v%s", version)
	app.Author = "Robin Monjo"
	app.Email = "robinmonjo@gmail.com"
	app.Usage = "micro init system for containers"

	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug, d", Usage: "run in debug mode"},
		cli.StringFlag{Name: "web-hook", Usage: "web hook to notify process status changes"},
		cli.StringFlag{Name: "bind-port", Usage: "port the process is expected to bind"},
		cli.StringFlag{Name: "io", Usage: "io of the process"},
	}

	app.Action = func(c *cli.Context) {

		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		exit, err := start(c)
		if err != nil {
			log.Error(err)
		}
		log.Debugf("exit status: %d", exit)
		os.Exit(exit)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func start(c *cli.Context) (int, error) {
	log.Debugf("dock pid: %d", os.Getpid())

	wire, err := iowire.NewWire(c.String("io"), "", iowire.NoColor)
	if err != nil {
		return 1, err
	}
	defer wire.Close()

	process := &process{
		argv: c.Args(),
		wire: wire,
	}
	defer process.cleanup()

	sh := newSignalsHandler()

	wh := c.String("web-hook")
	notifier.WebHook = wh

	processStateChanged(notifier.StatusStarting)
	defer processStateChanged(notifier.StatusCrashed)

	if err := process.start(); err != nil {
		return -1, err
	}

	log.Debugf("process pid: %d", process.pid())

	// // log rotation is specified and if stdout redirecto to a file
	// if c.Int("log-rotate") > 0 && s.URL.Scheme == "file" {
	// 	r := logrotate.NewRotator(s.URL.Host + s.URL.Path)
	// 	r.RotationDelay = time.Duration(c.Int("log-rotate")) * time.Hour
	// 	go r.StartWatching()
	// 	defer r.StopWatching()
	// }

	// watch ports
	go func() {
		bindPort := c.String("bind-port")
		if bindPort != "" {
			//wait for process to bind port
			for {
				p := procfs.Self()
				descendants, err := p.Descendants()
				if err != nil {
					log.Error(err)
					break
				}
				pids := []int{}
				for _, p := range descendants {
					pids = append(pids, p.Pid)
				}
				log.Debug(pids)

				binderPid, err := port.IsPortBound(bindPort, pids)
				if err != nil {
					log.Error(err)
					break
				}
				log.Debug(binderPid)
				if binderPid != -1 {
					log.Debugf("port %s binded by pid %d", bindPort, binderPid)
					processStateChanged(notifier.StatusRunning)
					break
				}
				time.Sleep(200 * time.Millisecond)
			}
		} else {
			processStateChanged(notifier.StatusRunning)
		}
	}()

	exit := sh.forward(process) //blocking call

	if c.Bool("debug") {
		//assert, at this point only 1 process should be running, self
		i, err := procfs.CountRunningProcs()
		if err != nil {
			log.Error(err)
		} else {
			if i != 1 {
				exit = 999
			}
		}
	}

	return exit, nil
}

func processStateChanged(state notifier.PsStatus) {
	log.Debugf("process state: %q", state)
	if notifier.WebHook != "" {
		notifier.NotifyHook(notifier.StatusCrashed)
	}
}
