package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func infoAction(c *cli.Context) error {
	userInfo, err := api.GetUserInfo()
	if err == api.ErrNotLogined {
		return cli.NewExitError("未登录，请先使用 `lean login` 命令登录 LeanCloud。", 1)
	}
	if err != nil {
		return newCliError(err)
	}
	fmt.Printf("当前登录用户: %s (%s)\r\n", userInfo.UserName, userInfo.Email)

	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		fmt.Println("当前目录没有关联任何 LeanCloud 应用。")
		return nil
	}
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return newCliError(err)
	}
	fmt.Printf("当前目录关联应用：%s (%s)\r\n", appInfo.AppName, appInfo.AppID)

	return nil
}
