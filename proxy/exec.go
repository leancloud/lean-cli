package proxy

import (
	"errors"
	"fmt"
	"os/exec"
)

var runtimeClis = map[string][]string{
	"udb":   {"mysql", "mycli"},
	"mysql": {"mysql", "mycli"},
	"redis": {"redis-cli", "iredis"},
	"mongo": {"mongo"},
}

func getCli(proxyInfo *ProxyInfo) (string, error) {
	clis := runtimeClis[proxyInfo.Runtime]
	if clis == nil {
		panic("invalid runtime")
	}

	var cli string
	for _, v := range clis {
		b, err := exec.LookPath(v)
		if err == nil {
			cli = b
			break
		}
	}
	if cli == "" {
		msg := fmt.Sprintf("No cli client for LeanDB runtime %s. Please install cli for runtime first.", proxyInfo.Runtime)
		return "", errors.New(msg)
	}

	return cli, nil
}

func GetCliArgs(p *ProxyInfo) []string {
	switch p.Runtime {
	case "redis":
		user := p.AuthUser
		if user == "" {
			user = "default"
		}
		return []string{"redis-cli", "-h", "127.0.0.1", "--user", user, "-a", p.AuthPassword, "-p", p.LocalPort}
	case "mongo":
		return []string{"mongo", "--host", "127.0.0.1", "-u", p.AuthUser, "-p", p.AuthPassword, "-port", p.LocalPort}
	case "udb":
		pass := fmt.Sprintf("-p%s", p.AuthPassword)
		return []string{"mysql", "-h", "127.0.0.1", "-u", p.AuthUser, pass, "-P", p.LocalPort}
	case "mysql":
		pass := fmt.Sprintf("-p%s", p.AuthPassword)
		return []string{"mysql", "-h", "127.0.0.1", "-u", p.AuthUser, pass, "-P", p.LocalPort}
	case "es":
		return []string{"curl", p.AuthUser, ":", p.AuthPassword, "@", "127.0.0.1", ":", p.LocalPort}
	}

	panic("invalid runtime")
}

func ForkExecCli(proxyInfo *ProxyInfo) error {
	return forkExec(proxyInfo)
}
