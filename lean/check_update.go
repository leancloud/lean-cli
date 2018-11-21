package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
)

const checkUpdateURL = "https://releases.leanapp.cn/leancloud/lean-cli/version.json"

var pkgType = "go"

func updateCommand() string {
	switch pkgType {
	case "go":
		return "go get -u github.com/leancloud/lean-cli/lean"
	case "homebrew":
		return "brew update && brew upgrade lean-cli"
	case "binary":
		return "Visit https://github.com/leancloud/lean-cli/releases"
	default:
		fmt.Fprintln(os.Stderr, "invalid pkgType: "+pkgType)
		return ""
	}
}

func checkUpdate() error {
	if pkgType == "homebrew-head" || pkgType == "aur-git" {
		return nil
	}
	resp, err := grequests.Get(checkUpdateURL, &grequests.RequestOptions{
		UserAgent: "LeanCloud-CLI/" + version.Version,
	})
	if err != nil {
		return err
	}

	var result struct {
		Version string `json:"version"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(resp.Bytes(), &result); err != nil {
		return err
	}

	current := semver.New(version.Version)
	latest := semver.New(strings.TrimPrefix(result.Version, "v"))

	if current.LessThan(*latest) {
		color.Green("New version found: %s. Update message:\r\n%s \r\nYou can upgrade by: %s", result.Version, result.Message, updateCommand())
	}

	return nil
}
