package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/version"
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
	case "binary":
		return "访问 https://github.com/leancloud/lean-cli/releases"
	default:
		fmt.Fprintln(os.Stderr, "invalid pkgType: "+pkgType)
		return ""
	}
}

func checkUpdate() error {
	if pkgType == "homebrew-head" {
		return nil
	}
	resp, err := grequests.Get(checkUpdateURL, &grequests.RequestOptions{
		UserAgent: "LeanCloud-CLI/" + version.Version,
	})
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
		color.Green("发现新版本 %s，变更如下：\r\n%s \r\n您可以通过以下方式升级：%s", result.Version, result.ChangeLog, updateCommand())
	}

	return nil
}
