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

	loginedRegions := regions.GetLoginedRegions()

	if len(loginedRegions) == 0 {
		return cli.NewExitError("Please log in first", 1)
	}

	for _, loginedRegion := range loginedRegions {
		loginedRegion := loginedRegion
		logp.Infof("Retrieving user info from region: %s\r\n", loginedRegion)
		userInfo, err := api.GetUserInfo(loginedRegion)
		if err != nil {
			callbacks = append(callbacks, func() {
				logp.Errorf("Failed to retrieve user info from region: %s: %v\r\n", loginedRegion, err)
			})
		} else {
			callbacks = append(callbacks, func() {
				logp.Infof("Current region:  %s User: %s (%s)\r\n", loginedRegion, userInfo.UserName, userInfo.Email)
			})
		}
	}

	logp.Info("Retrieving app info ...")
	appID, err := apps.GetCurrentAppID(".")

	if err == apps.ErrNoAppLinked {
		callbacks = append(callbacks, func() {
			logp.Warn("There is no LeanCloud app associated with the current directory")
		})
	} else if err != nil {
		callbacks = append(callbacks, func() {
			logp.Error("Failed to retrieve the app associated with the current directory", err)
		})
	} else {
		appInfo, err := api.GetAppInfo(appID)
		if err != nil {
			callbacks = append(callbacks, func() {
				logp.Error("Failed to retrieve app info: ", err)
			})
		} else {
			region, err := apps.GetAppRegion(appID)
			if err != nil {
				callbacks = append(callbacks, func() {
					logp.Error("Failed to retrieve app region: ", err)
				})
			} else {
				callbacks = append(callbacks, func() {
					logp.Infof("Current region: %s App: %s (%s)\r\n", region, appInfo.AppName, appInfo.AppID)
				})
				group, err := apps.GetCurrentGroup(".")
				if err != nil {
					callbacks = append(callbacks, func() {
						logp.Error("Failed to retrieve group info: ", err)
					})
				} else {
					callbacks = append(callbacks, func() {
						logp.Infof("Current group: %s\r\n", group)
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
