package runtimes

import (
	"testing"
)

func TestParseEngineConfig(t *testing.T) {
	config, err := parseEngineConfig([]byte{})
	if err != nil {
		t.Fatal(err)
	}
	if config.CMD != "" {
		t.Error()
	}

	config, err = parseEngineConfig([]byte("     "))
	if err != nil {
		t.Fatal(err)
	}
	if config.CMD != "" {
		t.Error()
	}

	config, err = parseEngineConfig([]byte("cmd: ls -al  "))
	if err != nil {
		t.Fatal(err)
	}
	if config.CMD != "ls -al" {
		t.Error()
	}
}

func TestParseEngineConfigCMD(t *testing.T) {
	config := engineConfig{
		CMD: "",
	}
	exec, args := config.parseCMD()
	if exec != "" || len(args) != 0 {
		t.Error()
	}

	config = engineConfig{
		CMD: "ls",
	}
	exec, args = config.parseCMD()
	if exec != "ls" || len(args) != 0 {
		t.Error()
	}

	config = engineConfig{
		CMD: "ls -a -l",
	}
	exec, args = config.parseCMD()
	if exec != "ls" || len(args) != 2 {
		t.Error()
	}

	config = engineConfig{
		CMD: "ls  -a  -l",
	}
	exec, args = config.parseCMD()
	if exec != "ls" || len(args) != 2 {
		t.Error()
	}
}
