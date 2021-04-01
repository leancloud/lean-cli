package commands

import (
	"errors"
	"fmt"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/logp"
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
		Content: "Please select an app: ",
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
	region := regions.Parse(regionString)
	if region == regions.Invalid {
		region = regions.ChinaNorth
	}

	currentApps, err := api.GetAppList(region)
	if err != nil {
		return err
	}

	if len(apps.GetRegionCache()) == 0 {
		return cli.NewExitError("Please create an app first.", 1)
	}

	// check if arg is an app id
	for _, app := range currentApps {
		if app.AppID == arg {
			fmt.Printf("Switching to app: %s (%s)\r\n", app.AppName, region)
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
					return cli.NewExitError("This app has multiple groups, please use --group specify one", 1)
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
		fmt.Printf("Switching to app: %s (%s)\r\n", matchedApp.AppName, region)
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
				return cli.NewExitError("This app has multiple groups, please use --group specify one.", 1)
			}
			groupName = groupList[0].GroupName
		}
		return apps.LinkGroup(".", groupName)
	} else if len(matchedApps) > 1 {
		return cli.NewExitError("Multiple apps are using this name. Please use app ID to identify the app instead.", 1)
	}

	return cli.NewExitError("Failed to find the designated app.", 1)
}

func checkOutWithWizard(regionString string, groupName string) error {
	var region regions.Region
	var err error
	if regionString == "" {
		loginedRegions := regions.GetLoginedRegions()
		if len(loginedRegions) == 0 {
			return cli.NewExitError("Please login first.", 1)
		} else if len(loginedRegions) == 1 {
			region = loginedRegions[0]
		} else {
			region, err = selectRegion(loginedRegions)
			if err != nil {
				return err
			}
		}
	} else {
		region = regions.Parse(regionString)
	}

	if region == regions.Invalid {
		return cli.NewExitError("Wrong region parameter", 1)
	}

	logp.Info("Retrieve app list ...")
	appList, err := api.GetAppList(region)
	if err != nil {
		return err
	}

	if len(apps.GetRegionCache()) == 0 {
		return cli.NewExitError("Please create an app first.", 1)
	}

	var sortedAppList []*api.GetAppListResult
	linq.From(appList).OrderBy(func(in interface{}) interface{} {
		return in.(*api.GetAppListResult).AppName[0]
	}).ToSlice(&sortedAppList)

	currentAppID, err := apps.GetCurrentAppID(".")
	if err != nil {
		if err != apps.ErrNoAppLinked && err != apps.ErrMissingRegionCache {
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

	var filtedGroups []*api.GetGroupsResult

	linq.From(groupList).Where(func(group interface{}) bool {
		return group.(*api.GetGroupsResult).Staging.Deployable || group.(*api.GetGroupsResult).Production.Deployable
	}).ToSlice(&filtedGroups)

	var group *api.GetGroupsResult
	if groupName == "" {
		group, err = selectGroup(filtedGroups)
		if err != nil {
			return err
		}
	} else {
		err = func() error {
			for _, group = range filtedGroups {
				if group.GroupName == groupName {
					return nil
				}
			}
			return errors.New("Cannot find group " + groupName)
		}()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Switching to app: %s, group: %s\r\n", app.AppName, group.GroupName)

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
	group := c.String("group")
	region := c.String("region")
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
	logp.Warn("`lean checkout` is deprecated, please use `lean switch` instead")
	return switchAction(c)
}
