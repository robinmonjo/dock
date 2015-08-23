package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
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
	}

	app.Action = func(c *cli.Context) {
		exit, err := start(c)
		if err != nil {
			log.Error(err)
		}
		log.Infof("exit status: %d", exit)
		os.Exit(exit)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func start(c *cli.Context) (int, error) {
	log.Infof("dock pid: %d", os.Getpid())

	process := &process{
		argv:   c.Args(),
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
	defer process.cleanup()

	signalsListener := newSignalsListener()

	var err error

	if c.Bool("interactive") {
		err = process.startInteractive()
	} else {
		err = process.start()
	}

	if err != nil {
		return -1, err
	}

	log.Infof("process pid: %d", process.pid())

	return signalsListener.forward(process), nil
}

func psCount() int {
	return 0
}
