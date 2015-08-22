package system

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
)

type ProcStatus struct {
	Pid    int              //pid of the process
	SigBlk []syscall.Signal //signals blocked by the process
	SigIgn []syscall.Signal //signals ignored by the process
	SigCgt []syscall.Signal //signals caught by the process
}

func NewProcStatus(pid int) (*ProcStatus, error) {
	//fill the ProcStatus struct from /proc/pid/status
	statusFile, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return nil, err
	}
	defer statusFile.Close()

	ps := &ProcStatus{Pid: pid}

	reader := csv.NewReader(statusFile)
	reader.Comma = ':'
	reader.FieldsPerRecord = 2
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		err = nil

		switch record[0] {
		case "SigBlk":
			ps.SigBlk, err = decodeMask(record[1])
		case "SigIgn":
			ps.SigIgn, err = decodeMask(record[1])
		case "SigCgt":
			ps.SigCgt, err = decodeMask(record[1])
		}

		if err != nil {
			return nil, err
		}
	}

	return ps, nil
}

func (ps *ProcStatus) SignalBlocked(sig syscall.Signal) bool {
	return exists(sig, ps.SigBlk)
}

func (ps *ProcStatus) SignalIgnored(sig syscall.Signal) bool {
	return exists(sig, ps.SigIgn)
}

func (ps *ProcStatus) SignalCaught(sig syscall.Signal) bool {
	return exists(sig, ps.SigCgt)
}

//naive implementation of signal mask decoding
//ref: http://jeff66ruan.github.io/blog/2014/03/31/sigpnd-sigblk-sigign-sigcgt-in-proc-status-file/
func decodeMask(maskStr string) ([]syscall.Signal, error) {
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

func exists(sig syscall.Signal, sigs []syscall.Signal) bool {
	for _, s := range sigs {
		if s == sig {
			return true
		}
	}
	return false
}
