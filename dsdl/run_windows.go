// +build !windows
package dsdl

import (
	"os"
	"syscall"
)

func runSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}

func kill(p *os.Process) error {
	return p.Kill()
}
