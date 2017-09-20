package runtimes

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
)

var (
	errNoEngineConfig = errors.New("No engine config")
)

type engineConfig struct {
	CMD string `yaml:"cmd"`
}

func (config *engineConfig) parseCMD() (string, []string) {
	if config.CMD == "" {
		return "", []string{}
	}
	splited := strings.Split(strings.TrimSpace(config.CMD), " ")
	if len(splited) == 1 {
		return splited[0], []string{}
	}
	exec, args := strings.TrimSpace(splited[0]), splited[1:len(splited)]
	trimedArgs := []string{}
	for _, arg := range args {
		trimed := strings.TrimSpace(arg)
		if trimed == "" {
			continue
		}
		trimedArgs = append(trimedArgs, trimed)
	}
	return exec, trimedArgs
}

func parseEngineConfig(body []byte) (*engineConfig, error) {
	var result engineConfig
	yaml.Unmarshal(body, &result)
	return &result, nil
}

func getEngineConfig(projectPath string) (*engineConfig, error) {
	content, err := ioutil.ReadFile(filepath.Join(projectPath, "leanengine.yaml"))
	if os.IsNotExist(err) {
		return nil, errNoEngineConfig
	}
	return parseEngineConfig(content)
}
