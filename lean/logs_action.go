package main

import (
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func logsAction(c *cli.Context) error {
	follow := c.Bool("f")
	env := c.String("e")
	limit := c.Int("limit")
	isProd := false

	if env == "staging" || env == "stag" {
		isProd = false
	} else if env == "production" || env == "" || env == "prod" {
		isProd = true
	} else {
		return cli.NewExitError("environment 参数必须为 staging 或者 production", 1)
	}

	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		return cli.NewExitError("没有关联任何 app，请使用 lean checkout 来关联应用。", 1)
	}
	if err != nil {
		return newCliError(err)
	}
	info, err := api.GetAppInfo(appID)
	if err != nil {
		return newCliError(err)
	}

	api.PrintLogs(color.Output, info.AppID, info.MasterKey, follow, isProd, limit)

	return nil
}
