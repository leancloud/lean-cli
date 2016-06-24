package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/utils"
)

func appListAction(c *cli.Context) {
	appList, err := apps.LinkedApps("")
	utils.CheckError(err)

	currentAppName, err := apps.CurrentAppName("")
	utils.CheckError(err)

	for _, app := range appList {
		if currentAppName == app.AppName {
			fmt.Printf("* %s - %s \r\n", app.AppName, app.AppID)
		} else {
			fmt.Printf("  %s - %s \r\n", app.AppName, app.AppID)
		}
	}
}
