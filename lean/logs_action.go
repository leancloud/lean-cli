package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func logsAction(c *cli.Context) error {
	follow := c.Bool("f")
	_ = follow

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

	api.PrintLogs(os.Stdout, info.AppID, info.MasterKey, follow)

	return nil
}
