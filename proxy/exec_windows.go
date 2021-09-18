package proxy

import (
	"log"
	"os/exec"
)

// TODO test on windows
func forkExec(proxyInfo *ProxyInfo) error {
	cli, err := getCli(proxyInfo)
	if err != nil {
		return err
	}

	// args := GetCliArgs(proxyInfo)
	cmd := exec.Command("cmd.exe", "/C", "start", cli)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
