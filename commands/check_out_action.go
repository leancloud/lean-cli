package commands

import (
	"errors"
	"fmt"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/chrysanthemum"
	"github.com/aisk/wizard"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
)

func selectCheckOutApp(appList []*api.GetAppListResult, currentAppID string) (*api.GetAppListResult, error) {
	var selectedApp *api.GetAppListResult
	question := wizard.Question{
		Content: "请选择 APP",
		Answers: []wizard.Answer{},
	}
	for _, app := range appList {
		answer := wizard.Answer{
			Content: app.AppName,
		}
		if app.AppID == currentAppID {
			answer.Content += color.RedString(" (current)")
		}
		// for scope problem
		func(app *api.GetAppListResult) {
			answer.Handler = func() {
				selectedApp = app
			}
		}(app)
		question.Answers = append(question.Answers, answer)
	}
	err := wizard.Ask([]wizard.Question{question})
	return selectedApp, err
}

func checkOutWithAppInfo(arg string, regionString string, groupName string) error {
	var region regions.Region
	switch regionString {
	case "cn", "CN", "":
		region = regions.CN
	case "us", "US":
		region = regions.US
	case "tab", "TAB":
		region = regions.TAB
	}
	currentApps, err := api.GetAppList(region)
	if err != nil {
		return err
	}

	// check if arg is an app id
	for _, app := range currentApps {
		if app.AppID == arg {
			fmt.Printf("切换至应用：%s (%s)", app.AppName, region)
			if err = apps.LinkApp(".", app.AppID); err != nil {
				return apps.LinkGroup(".", groupName)
			}
		}
	}

	// check if arg is a app name, and is the app name is unique
	matchedApps := make([]*api.GetAppListResult, 0)
	for _, app := range currentApps {
		if app.AppName == arg {
			matchedApps = append(matchedApps, app)
		}
	}
	if len(matchedApps) == 1 {
		fmt.Printf("切换至应用：%s (%s)", matchedApps[0].AppName, region)
		if err = apps.LinkApp(".", matchedApps[0].AppID); err != nil {
			return apps.LinkGroup(".", groupName)
		}
	} else if len(matchedApps) > 1 {
		return cli.NewExitError("找到多个应用使用此应用名，切换失败。请尝试使用 app ID 取代应用名来进行切换。", 1)
	}

	return cli.NewExitError("找不到对应的应用，切换失败。", 1)
}

func checkOutWithWizard(regionString string, groupName string) error {
	var region regions.Region
	var err error
	switch regionString {
	case "":
		region, err = selectRegion()
		if err != nil {
			return newCliError(err)
		}
	case "tab", "TAB":
		region = regions.TAB
	case "cn", "CN":
		region = regions.CN
	case "us", "US":
		region = regions.US
	default:
		return cli.NewExitError("错误的 region 参数", 1)
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

	currentAppID, err := apps.GetCurrentAppID(".")
	if err != nil {
		if err != apps.ErrNoAppLinked {
			return newCliError(err)
		}
	}

	app, err := selectCheckOutApp(sortedAppList, currentAppID)
	if err != nil {
		return newCliError(err)
	}

	groupList, err := api.GetGroups(app.AppID)
	if err != nil {
		return newCliError(err)
	}

	var group *api.GetGroupsResult
	if groupName == "" {
		group, err = selectGroup(groupList)
		if err != nil {
			return newCliError(err)
		}
	} else {
		err = func() error {
			for _, group = range groupList {
				if group.GroupName == groupName {
					return nil
				}
			}
			return errors.New("找不到分组 " + groupName)
		}()
		if err != nil {
			return err
		}
	}

	fmt.Printf("切换应用至：%s ，分组：%s\r\n", app.AppName, group.GroupName)

	err = apps.LinkApp(".", app.AppID)
	if err != nil {
		return newCliError(err)
	}
	err = apps.LinkGroup(".", group.GroupName)
	if err != nil {
		return newCliError(err)
	}
	return nil
}

func checkOutAction(c *cli.Context) error {
	region := c.String("region")
	group := c.String("group")
	if c.NArg() > 0 {
		arg := c.Args()[0]
		err := checkOutWithAppInfo(arg, region, group)
		if err != nil {
			return newCliError(err)
		}
		return nil
	}
	return checkOutWithWizard(region, group)
}
