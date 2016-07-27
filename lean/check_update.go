package main

import (
	"encoding/json"
	"log"

	"github.com/coreos/go-semver/semver"
	"github.com/leancloud/lean-cli/lean/version"
	"github.com/levigross/grequests"
)

const checkUpdateURL = "https://download.leancloud.cn/sdk/lean_cli.json"

var pkgType = "go"

func updateCommand() string {
	switch pkgType {
	case "go":
		return "go get -u github.com/leancloud/lean-cli/lean"
	case "homebrew":
		return "brew update && brew upgrade lean-cli"
	default:
		panic("invalid pkgType: " + pkgType)
	}
}

func checkUpdate() error {
	resp, err := grequests.Get(checkUpdateURL, nil)
	if err != nil {
		return err
	}

	var result struct {
		Version   string `json:"version"`
		ChangeLog string `json:"changelog"`
	}
	if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
		return err
	}

	current := semver.New(version.Version)
	latest := semver.New(result.Version)

	if current.LessThan(*latest) {
		log.Printf("发现新版本 %s，变更如下：\r\n%s\r\n您可以通过以下命令升级：%s", result.Version, result.ChangeLog, updateCommand())
	}

	return nil
}
