// +build linux
package dsdl

import "syscall"

func runSysProcAttr() syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

func kill(p *os.Process) error {
	pgid, err := syscall.Getpgid(r.spawned.Pid)
	if err == nil {
		syscall.Kill(-pgid, 15) // note the minus sign
	}
	return err
}
