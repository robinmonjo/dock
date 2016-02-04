package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/robinmonjo/dock/notifier"
	"github.com/robinmonjo/procfs"
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
		cli.BoolFlag{Name: "interactive, i", Usage: "run the process in a pty"},
		cli.BoolFlag{Name: "debug, d", Usage: "run in debug mode"},
		cli.StringFlag{Name: "web-hook", Usage: "web hook to notify process status changes"},
		cli.StringFlag{Name: "bind-port", Usage: "port the process is expected to bind"},
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

	process := &process{
		argv:   c.Args(),
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	defer process.cleanup()

	sh := newSignalsHandler()

	wh := c.String("web-hook")
	notifier.WebHook = wh

	if wh != "" {
		notifier.NotifyHook(notifier.StatusStarting)
		defer notifier.NotifyHook(notifier.StatusCrashed)
	}

	var err error

	if c.Bool("interactive") {
		err = process.startInteractive()
	} else {
		err = process.start()
	}

	if err != nil {
		return -1, err
	}

	if wh != "" {
		go func() {
			bp := c.String("bind-port")
			if bp != "" {
				//wait for process to bind port

			} else {
				notifier.NotifyHook(notifier.StatusRunning)
			}
		}()
	}

	log.Debugf("process pid: %d", process.pid())

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
