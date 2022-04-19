package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/wizard"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/boilerplate"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

func selectApp(appList []*api.GetAppListResult) (*api.GetAppListResult, error) {
	var selectedApp *api.GetAppListResult
	question := wizard.Question{
		Content: "Please select an app: ",
		Answers: []wizard.Answer{},
	}
	for _, app := range appList {
		answer := wizard.Answer{
			Content: app.AppName,
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

func selectGroup(groupList []*api.GetGroupsResult) (*api.GetGroupsResult, error) {
	if len(groupList) == 1 {
		return groupList[0], nil
	}

	var selectedGroup *api.GetGroupsResult
	question := wizard.Question{
		Content: "Please select a LeanEngine group",
		Answers: []wizard.Answer{},
	}
	for _, group := range groupList {
		answer := wizard.Answer{
			Content: group.GroupName,
		}
		func(group *api.GetGroupsResult) {
			answer.Handler = func() {
				selectedGroup = group
			}
		}(group)
		question.Answers = append(question.Answers, answer)
	}
	err := wizard.Ask([]wizard.Question{question})
	return selectedGroup, err
}

func selectBoilerplate() (*boilerplate.Boilerplate, error) {
	var selectedBoilerplate boilerplate.Boilerplate

	question := wizard.Question{
		Content: "Please select an app template: ",
		Answers: []wizard.Answer{},
	}
	for _, boil := range boilerplate.Boilerplates {
		answer := wizard.Answer{
			Content: boil.Name,
		}
		func(boil boilerplate.Boilerplate) {
			answer.Handler = func() {
				selectedBoilerplate = boil
			}
		}(boil)
		question.Answers = append(question.Answers, answer)
	}
	err := wizard.Ask([]wizard.Question{question})
	return &selectedBoilerplate, err
}

func selectRegion(loginedRegions []regions.Region) (regions.Region, error) {
	region := regions.Invalid
	question := wizard.Question{
		Content: "Please select a region: ",
		Answers: []wizard.Answer{},
	}

	for _, r := range loginedRegions {
		answer := wizard.Answer{
			Content: r.Description(),
		}
		func(r regions.Region) {
			answer.Handler = func() {
				region = r
			}
		}(r)
		question.Answers = append(question.Answers, answer)
	}
	err := wizard.Ask([]wizard.Question{question})
	return region, err
}

func newAction(c *cli.Context) error {
	groupName := c.String("group")
	regionString := c.String("region")
	if c.NArg() < 1 {
		return cli.NewExitError(fmt.Sprintf("You must specify a directory name like `%s new engine-project`", os.Args[0]), 1)
	}
	dest := c.Args()[0]

	boil, err := selectBoilerplate()
	if err != nil {
		return err
	}

	var region regions.Region
	if regionString == "" {
		loginedRegions := regions.GetLoginedRegions(version.AvailableRegions)
		if len(loginedRegions) == 0 {
			return cli.NewExitError("Please log in first.", 1)
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
		cli.NewExitError("Invalid region", 1)
	}

	appList, err := api.GetAppList(region)
	if err != nil {
		return err
	}

	if len(apps.GetRegionCache()) == 0 {
		return cli.NewExitError("Please create an app first.", 1)
	}

	var orderedAppList []*api.GetAppListResult
	linq.From(appList).OrderBy(func(in interface{}) interface{} {
		return in.(*api.GetAppListResult).AppName[0]
	}).ToSlice(&orderedAppList)

	app, err := selectApp(orderedAppList)
	if err != nil {
		return err
	}

	groupList, err := api.GetGroups(app.AppID)
	if err != nil {
		return err
	}
	if groupName == "" {
		group, err := selectGroup(groupList)
		if err != nil {
			return err
		}
		groupName = group.GroupName
	} else {
		err = func() error {
			for _, group := range groupList {
				if group.GroupName == groupName {
					return nil
				}
			}
			return errors.New("Failed to find group " + groupName)
		}()
		if err != nil {
			return err
		}
	}

	if err = boilerplate.CreateProject(boil, dest, app.AppID, region); err != nil {
		return err
	}

	err = apps.LinkApp(dest, app.AppID)
	if err != nil {
		return err
	}

	err = apps.LinkGroup(dest, groupName)
	if err != nil {
		return err
	}

	return nil
}
