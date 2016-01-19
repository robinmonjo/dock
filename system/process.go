package system

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

const (
	noChildrenExitCode = 1
)

type Process struct {
	Pid int
}

//store information about a process, found in /proc/pid/status
type ProcessStatus struct {
	Name   string
	PPid   int
	State  string
	SigBlk []syscall.Signal
	SigIgn []syscall.Signal
	SigCgt []syscall.Signal
}

//return ProcessStatus of the process
func (p *Process) Status() (*ProcessStatus, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/status", p.Pid))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.parseStatusFile(f)
}

func (p *Process) parseStatusFile(f io.Reader) (*ProcessStatus, error) {
	s := &ProcessStatus{}

	reader := csv.NewReader(f)
	reader.Comma = ':'
	reader.FieldsPerRecord = 2
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch record[0] {
		case "Name":
			s.Name = strings.TrimSpace(record[1])
		case "PPid":
			s.PPid, err = strconv.Atoi(strings.TrimSpace(record[1]))
		case "State":
			s.State = strings.TrimSpace(record[1])
		case "SigBlk":
			s.SigBlk, err = decodeSigMask(record[1])
		case "SigIgn":
			s.SigIgn, err = decodeSigMask(record[1])
		case "SigCgt":
			s.SigCgt, err = decodeSigMask(record[1])
		}
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

//return all process children and grand children (using pgrep), including process pid
func (p *Process) Children() ([]int, error) {

	pids := []int{p.Pid}
	cursor := 0

	for {

		if cursor > len(pids) {
			return pids, nil
		}

		pid := pids[cursor]
		out, err := exec.Command("pgrep", "-P", fmt.Sprintf("%d", pid)).Output()
		if err != nil {
			if exitStatus(err) != noChildrenExitCode {
				return nil, err
			}
		} else {
			cpids := strings.Split(string(out), "\n")

			if cpids[len(cpids)-1] == "" {
				cpids = cpids[:len(cpids)-1] //remove the last empty line
			}

			cpidsInt, err := intifySlice(cpids)
			if err != nil {
				return nil, err
			}
			pids = append(pids, cpidsInt...)
		}

		cursor++
	}

	panic("-- should never been executed --")
}

/*func (p *Process) Sockets() ([]*Socket, error) {
	sockets, err := netstat()
	if err != nil {
		return nil, err
	}
	res := []*Socket{}
	for _, s := range sockets {
		if s.Pid == p.Pid {
			res = append(res, s)
		}
	}
	return res, nil
}*/

//implementation of signal mask decoding
//ref: http://jeff66ruan.github.io/blog/2014/03/31/sigpnd-sigblk-sigign-sigcgt-in-proc-status-file/
func decodeSigMask(maskStr string) ([]syscall.Signal, error) {
	b, err := hex.DecodeString(strings.TrimSpace(maskStr))
	if err != nil {
		return nil, err
	}
	//interested in the 32 right bits of the mask
	mask := int32(b[4])<<24 | int32(b[5])<<16 | int32(b[6])<<8 | int32(b[7])

	var signals []syscall.Signal

	for i := 0; i < 32; i++ {
		submask := int32(1 << uint(i))
		if mask&submask > 0 {
			signals = append(signals, syscall.Signal(i+1))
		}
	}

	return signals, nil
}

func exitStatus(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 555
}

func intifySlice(arr []string) ([]int, error) {
	res := []int{}
	for _, s := range arr {
		i, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		res = append(res, i)
	}
	return res, nil
}
