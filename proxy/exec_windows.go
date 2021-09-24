package proxy

import (
	"os/exec"
	"strings"
	"syscall"
)

// reference https://stackoverflow.com/questions/50531370/start-a-detached-process-on-windows-using-golang
func forkExec(p *ProxyInfo, _ chan bool) error {
	cli, err := getCli(p)
	if err != nil {
		return err
	}

	args := []string{"/C", "start", cli}
	args = append(args, GetCliArgs(p)[1:]...)
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: strings.Join(args, " "),
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
