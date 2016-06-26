package main

import (
	"log"

	"github.com/aisk/wizard"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/utils"
)

func appAddAction(c *cli.Context) {
	var appName, appID string
	wizard.Ask([]wizard.Question{
		{
			Content: "请输入应用名：",
			Input: &wizard.Input{
				Hidden: false,
				Result: &appName,
			},
		},
		{
			Content: "请输入应用 appID：",
			Input: &wizard.Input{
				Hidden: false,
				Result: &appID,
			},
		},
	})

	utils.CheckError(apps.AddApp("", appName, appID))
	utils.CheckError(apps.SwitchApp("", appName))
	log.Println("已切换至应用：" + appName)
}
