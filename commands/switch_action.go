package commands

import (
	"errors"
	"fmt"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/chrysanthemum"
	"github.com/aisk/wizard"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
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
			fmt.Printf("切换至应用：%s (%s)\r\n", app.AppName, region)
			err = apps.LinkApp(".", app.AppID)
			if err != nil {
				return err
			}
			if groupName == "" {
				groupList, err := api.GetGroups(app.AppID)
				if err != nil {
					return err
				}
				if len(groupList) != 1 {
					return cli.NewExitError("此应用对应多个分组，请使用 --group 参数指定分组", 1)
				}
				groupName = groupList[0].GroupName
			}
			return apps.LinkGroup(".", groupName)
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
		matchedApp := matchedApps[0]
		fmt.Printf("切换至应用：%s (%s)\r\n", matchedApp.AppName, region)
		err = apps.LinkApp(".", matchedApps[0].AppID)
		if err != nil {
			return err
		}
		if groupName == "" {
			groupList, err := api.GetGroups(matchedApp.AppID)
			if err != nil {
				return err
			}
			if len(groupList) != 1 {
				return cli.NewExitError("此应用对应多个分组，请使用 --group 参数指定分组", 1)
			}
			groupName = groupList[0].GroupName
		}
		return apps.LinkGroup(".", groupName)
	} else if len(matchedApps) > 1 {
		return cli.NewExitError("找到多个应用使用此应用名，切换失败。请尝试使用 app ID 取代应用名来进行切换。", 1)
	}

	return cli.NewExitError("找不到对应的应用，切换失败。", 1)
}

func checkOutWithWizard(regionString string, groupName string) error {
	var region regions.Region
	var err error
	switch regionString {
	case "tab", "TAB":
		region = regions.TAB
	case "cn", "CN":
		region = regions.CN
	case "us", "US":
		region = regions.US
	case "":
		loginedRegions, err := api.GetLoginedRegion()
		if err != nil {
			return err
		}
		if len(loginedRegions) == 0 {

		} else if len(loginedRegions) == 1 {
			region = loginedRegions[0]
		} else {
			region, err = selectRegion(loginedRegions)
			if err != nil {
				return err
			}
		}
	default:
		return cli.NewExitError("错误的 region 参数", 1)
	}

	spinner := chrysanthemum.New("获取应用列表").Start()
	appList, err := api.GetAppList(region)
	if err != nil {
		spinner.Failed()
		return err
	}
	spinner.Successed()

	var sortedAppList []*api.GetAppListResult
	linq.From(appList).OrderBy(func(in interface{}) interface{} {
		return in.(*api.GetAppListResult).AppName[0]
	}).ToSlice(&sortedAppList)

	currentAppID, err := apps.GetCurrentAppID(".")
	if err != nil {
		if err != apps.ErrNoAppLinked {
			return err
		}
	}

	app, err := selectCheckOutApp(sortedAppList, currentAppID)
	if err != nil {
		return err
	}

	groupList, err := api.GetGroups(app.AppID)
	if err != nil {
		return err
	}

	var group *api.GetGroupsResult
	if groupName == "" {
		group, err = selectGroup(groupList)
		if err != nil {
			return err
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
		return err
	}
	err = apps.LinkGroup(".", group.GroupName)
	if err != nil {
		return err
	}
	return nil
}

func switchAction(c *cli.Context) error {
	region := c.String("region")
	group := c.String("group")
	if c.NArg() > 0 {
		arg := c.Args()[0]
		err := checkOutWithAppInfo(arg, region, group)
		if err != nil {
			return err
		}
		return nil
	}
	return checkOutWithWizard(region, group)
}

func checkOutAction(c *cli.Context) error {
	fmt.Printf(" %s [WARNNING] lean checkout 被标记为废弃，请使用 lean switch 代替此命令。\r\n", chrysanthemum.Fail)
	return switchAction(c)
}
