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
	"github.com/robinmonjo/dock/logrotate"
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
		cli.StringFlag{Name: "io", Usage: "smart stdin / stdout (see README for more info)"},
		cli.StringFlag{Name: "web-hook", Usage: "hook where process status changes should be notified"},
		cli.StringFlag{Name: "bind-port", Usage: "port the process is expected to bind"},
		cli.IntFlag{Name: "log-rotate", Usage: "duration in hour when stdoud should rotate (if `--io` is a file)"},
		cli.StringFlag{Name: "stdout-prefix", Usage: "add a prefix to stdout lines (format: <prefix>:<color>)"},
		cli.BoolFlag{Name: "debug, d", Usage: "run with verbose output (for developpers)"},
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

	wire, err := iowire.NewWire(c.String("io"))
	if err != nil {
		return 1, err
	}
	defer wire.Close()

	//experimental, only work fine over network
	prefix, color := parsePrefixArg(c.String("stdout-prefix"))
	if wire.URL.Scheme != "file" && wire.Output != os.Stdout {
		wire.SetPrefix(prefix, color)
	}

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

	// log rotation is specified and if stdout redirecto to a file
	if c.Int("log-rotate") > 0 && wire.URL.Scheme == "file" {
		r := logrotate.NewRotator(wire.URL.Host + wire.URL.Path)
		r.RotationDelay = time.Duration(c.Int("log-rotate")) * time.Hour
		go r.StartWatching()
		defer r.StopWatching()
	}

	// watch ports
	go func() {
		bindPort := c.String("bind-port")
		if bindPort != "" {
			waitPortBinding(bindPort)
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
	log.Debugf("exit code: %d", exit)
	return exit, nil
}

func processStateChanged(state notifier.PsStatus) {
	log.Debugf("process state: %q", state)
	if notifier.WebHook != "" {
		notifier.NotifyHook(state)
	}
}

func waitPortBinding(watchedPort string) {
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

		binderPid, err := port.IsPortBound(watchedPort, pids)
		if err != nil {
			log.Error(err)
			break
		}
		log.Debug(binderPid)
		if binderPid != -1 {
			log.Debugf("port %s binded by pid %d", watchedPort, binderPid)
			processStateChanged(notifier.StatusRunning)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
}
