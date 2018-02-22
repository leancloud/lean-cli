package commands

import (
	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
)

func infoAction(c *cli.Context) error {
	callbacks := make([]func(), 0)

	loginedRegions := apps.GetLoginedRegions()

	if len(loginedRegions) == 0 {
		logp.Error("未登录")
		return nil
	}

	for _, loginedRegion := range loginedRegions {
		logp.Infof("获取 %s 节点用户信息\r\n", loginedRegion)
		userInfo, err := api.GetUserInfo(loginedRegion)
		if err != nil {
			callbacks = append(callbacks, func() {
				logp.Errorf("获取 %s 节点用户信息失败: %v\r\n", loginedRegion, err)
			})
		} else {
			func(loginedRegion regions.Region) {
				callbacks = append(callbacks, func() {
					logp.Infof("当前 %s 节点登录用户: %s (%s)\r\n", loginedRegion, userInfo.UserName, userInfo.Email)
				})
			}(loginedRegion)
		}
	}

	logp.Info("获取应用信息")
	appID, err := apps.GetCurrentAppID(".")

	if err == apps.ErrNoAppLinked {
		callbacks = append(callbacks, func() {
			logp.Warn("当前目录没有关联任何 LeanCloud 应用")
		})
	} else if err != nil {
		callbacks = append(callbacks, func() {
			logp.Error("获取当前目录关联应用失败：", err)
		})
	} else {
		appInfo, err := api.GetAppInfo(appID)
		if err != nil {
			callbacks = append(callbacks, func() {
				logp.Error("获取应用信息失败：", err)
			})
		} else {
			region, err := apps.GetAppRegion(appID)
			if err != nil {
				callbacks = append(callbacks, func() {
					logp.Error("获取应用节点信息失败：", err)
				})
			} else {
				callbacks = append(callbacks, func() {
					logp.Infof("当前目录关联 %s 节点应用：%s (%s)\r\n", region, appInfo.AppName, appInfo.AppID)
				})
				group, err := apps.GetCurrentGroup(".")
				if err != nil {
					callbacks = append(callbacks, func() {
						logp.Error("获取关联分组信息失败：", err)
					})
				} else {
					callbacks = append(callbacks, func() {
						logp.Infof("当前目录关联分组：%s\r\n", group)
					})
				}
			}
		}
	}

	for _, callback := range callbacks {
		callback()
	}

	return nil
}
