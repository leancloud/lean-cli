package main

import (
	"fmt"

	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func infoAction(c *cli.Context) error {
	bar := chrysanthemum.New("获取用户信息").Start()
	userInfo, err := api.GetUserInfo()
	if err == api.ErrNotLogined {
		return cli.NewExitError("未登录，请先使用 `lean login` 命令登录 LeanCloud。", 1)
	}
	if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	bar.End()
	fmt.Printf("当前登录用户: %s (%s)\r\n", userInfo.UserName, userInfo.Email)

	bar = chrysanthemum.New("获取应用信息").Start()
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		bar.End()
		fmt.Println("当前目录没有关联任何 LeanCloud 应用。")
		return nil
	} else if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	region, err := api.GetAppRegion(appID)
	if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	bar.End()

	fmt.Printf("当前目录关联节点：%s \r\n", region)
	fmt.Printf("当前目录关联应用：%s (%s)\r\n", appInfo.AppName, appInfo.AppID)

	return nil
}
