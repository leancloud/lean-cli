//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package commands

import (
	"os/exec"
	"runtime"
	"syscall"
)

func StartBackgroundCommand(cmd *exec.Cmd) error {
	if runtime.GOOS == "darwin" {
		syscall.Sync() // workaround for https://github.com/golang/go/issues/33565
	}
	return cmd.Start()
}
