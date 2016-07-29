package main

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func checkOutAction(c *cli.Context) error {
	if c.NArg() > 0 {
		appID := c.Args()[0]
		log.Println("切换至应用：" + appID)
		err := apps.LinkApp("", appID)
		if err != nil {
			return newCliError(err)
		}
		return nil
	}

	op.Write("获取应用列表")
	appList, err := api.GetAppList()
	if err != nil {
		op.Failed()
		return newCliError(err)
	}
	op.Successed()

	app := selectApp(appList)
	log.Println("切换应用至 " + app.AppName)

	err = apps.LinkApp("", app.AppID)
	if err != nil {
		return newCliError(err)
	}
	return nil
}
