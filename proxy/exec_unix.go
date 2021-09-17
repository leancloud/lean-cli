//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package proxy

import (
	"os"
	"syscall"
)

func forkExec(proxyInfo *ProxyInfo) {
	cli := getCli(proxyInfo)
	args := GetCliArgs(proxyInfo)
	procAttr := &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{0, 1, 2},
	}
	_, err := syscall.ForkExec(cli, args, procAttr)
	if err != nil {
		panic(err)
	}
}
