package proxy

import (
	"log"
	"os/exec"
)

// TODO test on windows
func forkExec(proxyInfo *ProxyInfo) {
	cli := getCli(proxyInfo)
	// args := GetCliArgs(proxyInfo)
	cmd := exec.Command("cmd.exe", "/C", "start", cli)
	if err := cmd.Run(); err != nil {
		log.Println("Error:", err)
	}
	return
}
