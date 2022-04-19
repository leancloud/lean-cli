package proxy

import (
	"errors"
	"fmt"
	"os/exec"
)

var RuntimeClis = map[string][]string{
	"udb":   {"mycli", "mysql"},
	"mysql": {"mycli", "mysql"},
	"redis": {"iredis", "redis-cli", "memurai-cli"},
	"mongo": {"mongo"},
}

func GetCli(p *ProxyInfo) ([]string, error) {
	var cli string
	if p.Runtime == "es" {
		cli = "curl"
	} else {
		clis := RuntimeClis[p.Runtime]
		if clis == nil {
			panic("invalid runtime")
		}

		for _, v := range clis {
			_, err := exec.LookPath(v)
			if err == nil {
				cli = v
				break
			}
		}
		if cli == "" {
			msg := fmt.Sprintf("No cli client for LeanDB runtime %s. Please install cli for runtime first.", p.Runtime)
			return nil, errors.New(msg)
		}
	}

	switch p.Runtime {
	case "redis":
		return []string{cli, "-h", "127.0.0.1", "-a", p.AuthPassword, "-p", p.LocalPort}, nil
	case "mongo":
		return []string{cli, "--host", "127.0.0.1", "-u", p.AuthUser, "-p", p.AuthPassword, "-port", p.LocalPort}, nil
	case "udb", "mysql":
		pass := fmt.Sprintf("-p%s", p.AuthPassword)
		return []string{cli, "-h", "127.0.0.1", "-u", p.AuthUser, pass, "-P", p.LocalPort}, nil
	case "es":
		return []string{cli, p.AuthUser + ":" + p.AuthPassword + "@" + "127.0.0.1" + ":" + p.LocalPort}, nil
	}

	panic("invalid runtime")
}
