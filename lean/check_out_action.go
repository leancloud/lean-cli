package main

import (
	"fmt"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func checkOutAction(c *cli.Context) error {
	if c.NArg() > 0 {
		appID := c.Args()[0]
		fmt.Println("切换至应用：" + appID)
		err := apps.LinkApp("", appID)
		if err != nil {
			return newCliError(err)
		}
		return nil
	}

	region, err := selectRegion()
	if err != nil {
		return newCliError(err)
	}

	spinner := chrysanthemum.New("获取应用列表").Start()
	appList, err := api.GetAppList(region)
	if err != nil {
		spinner.Failed()
		return newCliError(err)
	}
	spinner.Successed()

	var sortedAppList []*api.GetAppListResult
	linq.From(appList).OrderBy(func(in interface{}) interface{} {
		return in.(*api.GetAppListResult).AppName[0]
	}).ToSlice(&sortedAppList)

	// disable it because it's buggy
	// sortedAppList, err = apps.MergeWithRecentApps(".", sortedAppList)
	// if err != nil {
	// 	return newCliError(err)
	// }

	// remove current linked app from app list
	curentAppID, err := apps.GetCurrentAppID(".")
	if err != nil {
		if err != apps.ErrNoAppLinked {
			return newCliError(err)
		}
	} else {
		for i, app := range sortedAppList {
			if app.AppID == curentAppID {
				sortedAppList = append(sortedAppList[:i], sortedAppList[i+1:]...)
			}
		}
	}

	app, err := selectApp(sortedAppList)
	if err != nil {
		return newCliError(err)
	}
	fmt.Println("切换应用至 " + app.AppName)

	err = apps.LinkApp(".", app.AppID)
	if err != nil {
		return newCliError(err)
	}
	return nil
}
