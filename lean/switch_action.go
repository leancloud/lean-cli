package main

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func switchAction(c *cli.Context) error {
	appList, err := api.GetAppList()
	log.Println(appList)
	if err != nil {
		return newCliError(err)
	}

	app := selectApp(appList)
	log.Println("切换应用至 " + app.AppName)

	err = apps.LinkApp("", app.AppID)
	if err != nil {
		return newCliError(err)
	}
	return nil
}
