package main

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/utils"
)

func upAction(c *cli.Context) {
	// TODO: get port from args
	port := "3000"
	// TODO:
	apiServerURL := "https://api.leancloud.cn"

	appInfo, err := apps.CurrentAppInfo(".")
	utils.CheckError(err)
	log.Println(">>", appInfo)

	rtm, err := apps.DetectRuntime(".")
	utils.CheckError(err)

	rtm.Envs["LC_APP_ID"] = appInfo.AppID
	rtm.Envs["LC_APP_KEY"] = appInfo.AppKey
	rtm.Envs["LC_APP_MASTER_KEY"] = appInfo.MasterKey
	rtm.Envs["LC_APP_PORT"] = port
	rtm.Envs["LC_API_SERVER"] = apiServerURL
	rtm.Envs["LEANCLOUD_APP_ID"] = appInfo.AppID
	rtm.Envs["LEANCLOUD_APP_KEY"] = appInfo.AppKey
	rtm.Envs["LEANCLOUD_APP_MASTER_KEY"] = appInfo.MasterKey
	rtm.Envs["LEANCLOUD_APP_PORT"] = port
	rtm.Envs["LEANCLOUD_API_SERVER"] = apiServerURL

	rtm.Run()
}
