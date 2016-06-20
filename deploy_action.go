package main

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean/api"
	"github.com/leancloud/lean/apps"
	"github.com/leancloud/lean/utils"
)

func deployAction(*cli.Context) {
	_apps, err := apps.GetApps(".")
	utils.CheckError(err)
	if len(_apps) == 0 {
		log.Fatalln("没有关联任何 app，请使用 lean app add 来关联应用。")
	}

	// TODO: specific app
	app := _apps[0]

	appInfo, err := apps.GetAppInfo(app.AppID)
	utils.CheckError(err)

	client := api.Client{
		AppID:     app.AppID,
		MasterKey: appInfo.MasterKey,
		Region:    api.RegionCN,
	}
	log.Println(client.EngineInfo())
}
