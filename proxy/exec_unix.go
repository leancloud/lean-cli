//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package proxy

import (
	"os"
	"syscall"
)

func forkExec(proxyInfo *ProxyInfo, term chan bool) error {
	cli, err := getCli(proxyInfo)
	if err != nil {
		return err
	}
	args := GetCliArgs(proxyInfo)
	procAttr := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{0, 1, 2},
	}
	pid, er := syscall.ForkExec(cli, args, procAttr)
	if er != nil {
		return er
	}

	child, e := os.FindProcess(pid)
	if e != nil {
		return e
	}
	child.Wait()
	term <- true

	return nil
}
