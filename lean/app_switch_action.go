package main

import (
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/utils"
	"log"
)

func appSwitchAction(c *cli.Context) {
	if c.NArg() != 1 {
		log.Fatal("Usage: lean app switch <app-name>")
	}
	appName := c.Args()[0]

	utils.CheckError(apps.SwitchApp("", appName))
}
