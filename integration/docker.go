package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/kr/pty"
)

// struct to play with the docker binary
type docker struct {
	path   string
	ps     *os.Process
	stdout []byte
	stderr []byte
}

func newDocker() *docker {
	return &docker{
		path: "docker", //docker must be in the path
	}
}

func (d *docker) start(usePty bool, args ...string) error {
	cmd := exec.Command(d.path, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	var err error

	if usePty {
		var f *os.File
		f, err = pty.Start(cmd)
		stdout, stderr = f, f
		defer f.Close()
	} else {
		err = cmd.Start()
	}

	if err != nil {
		return err
	}

	d.ps = cmd.Process
	d.stdout, _ = ioutil.ReadAll(stdout)
	d.stderr, _ = ioutil.ReadAll(stderr)

	return cmd.Wait()
}

func (d *docker) debugInfo() string {
	return fmt.Sprintf("stdout: %s\nstderr: %s", string(d.stdout), string(d.stderr))
}
