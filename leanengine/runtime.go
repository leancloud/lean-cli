package main

import (
	"os"
)

type Runtime struct {
	AppPath string
	Type    string
}

func isFileExist(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		os.Exit(1)
	}
	fileInfo, err := f.Stat()
	if err != nil {
		os.Exit(1)
	}
	return !fileInfo.IsDir()
}

func RuntimeNew(path string) *Runtime {
	runtime := Runtime{AppPath: path}
	runtime.detectType()
	return &runtime
}

func (runtime *Runtime) detectType() {
	if runtime.detectPython() {
		runtime.Type = "python"
	} else if runtime.detectNode() {
		runtime.Type = "node"
	}
}

func (runtime *Runtime) detectPython() bool {
	return isFileExist("/requirements.txt") && isFileExist("wsgi.py")
}

func (runtime *Runtime) detectNode() bool {
	return isFileExist("package.json") && isFileExist("server.js")
}
