package main

import (
	"fmt"

	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/api/regions"
	"github.com/leancloud/lean-cli/lean/apps"
)

func infoAction(c *cli.Context) error {
	callbacks := make([]func(), 0)

	loginedRegions, err := api.GetLoginedRegion()
	if err != nil {
		return newCliError(err)
	}

	if len(loginedRegions) == 0 {
		fmt.Println("未登录")
		return nil
	}

	for _, loginedRegion := range loginedRegions {
		bar := chrysanthemum.New(fmt.Sprintf("获取 %s 节点用户信息", loginedRegion)).Start()
		userInfo, err := api.GetUserInfo(loginedRegion)
		if err != nil {
			bar.Failed()
			callbacks = append(callbacks, func() {
				fmt.Printf("获取 %s 节点用户信息失败: %v\r\n", loginedRegion, err)
			})
		} else {
			bar.Successed()
			func(loginedRegion regions.Region) {
				callbacks = append(callbacks, func() {
					fmt.Printf("当前 %s 节点登录用户: %s (%s)\r\n", loginedRegion, userInfo.UserName, userInfo.Email)
				})
			}(loginedRegion)
		}
	}

	bar := chrysanthemum.New("获取应用信息").Start()
	appID, err := apps.GetCurrentAppID("")
	_ = appID

	if err == apps.ErrNoAppLinked {
		bar.Failed()
		callbacks = append(callbacks, func() {
			fmt.Println("当前目录没有关联任何 LeanCloud 应用")
		})
	} else if err != nil {
		bar.Failed()
		callbacks = append(callbacks, func() {
			fmt.Println("获取当前目录关联应用失败：", err)
		})
	} else {
		appInfo, err := api.GetAppInfo(appID)
		if err != nil {
			bar.Failed()
			callbacks = append(callbacks, func() {
				fmt.Println("获取应用信息失败：", err)
			})
		} else {
			bar.Successed()
			region, _ := api.GetAppRegion(appID)
			callbacks = append(callbacks, func() {
				fmt.Printf("当前目录关联 %s 节点应用：%s (%s)\r\n", region, appInfo.AppName, appInfo.AppID)
			})
		}
	}

	for _, callback := range callbacks {
		callback()
	}

	return nil
}
