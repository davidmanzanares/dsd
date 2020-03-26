// +build linux
package dsdl

import (
	"os"
	"syscall"
)

func runSysProcAttr() syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

func kill(p *os.Process) error {
	pgid, err := syscall.Getpgid(p.Pid)
	if err == nil {
		return syscall.Kill(-pgid, 15) // note the minus sign
	}
	return err
}
