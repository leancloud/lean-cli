package commands

import (
	"errors"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/wizard"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/boilerplate"
	"github.com/urfave/cli"
)

func selectApp(appList []*api.GetAppListResult) (*api.GetAppListResult, error) {
	var selectedApp *api.GetAppListResult
	question := wizard.Question{
		Content: "请选择 APP",
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
		Content: "请选择云引擎分组",
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
	var selectBoil *boilerplate.Boilerplate
	boils, err := boilerplate.GetBoilerplateList()
	if err != nil {
		return nil, err
	}

	orderedBoils := []*boilerplate.Boilerplate{}
	linq.From(boils).OrderBy(func(in interface{}) interface{} {
		return in.(*boilerplate.Boilerplate).Name[0]
	}).ToSlice(&orderedBoils)

	question := wizard.Question{
		Content: "请选择需要创建的应用模版",
		Answers: []wizard.Answer{},
	}
	for _, boil := range orderedBoils {
		answer := wizard.Answer{
			Content: boil.Name,
		}
		// for scope problem
		func(boil *boilerplate.Boilerplate) {
			answer.Handler = func() {
				selectBoil = boil
			}
		}(boil)
		question.Answers = append(question.Answers, answer)
	}
	err = wizard.Ask([]wizard.Question{question})
	return selectBoil, err
}

func selectRegion() (regions.Region, error) {
	region := regions.Invalid
	err := wizard.Ask([]wizard.Question{
		{
			Content: "请选择应用节点",
			Answers: []wizard.Answer{
				{
					Content: "国内",
					Handler: func() {
						region = regions.CN
					},
				},
				{
					Content: "美国",
					Handler: func() {
						region = regions.US
					},
				},
				{
					Content: "TAB",
					Handler: func() {
						region = regions.TAB
					},
				},
			},
		},
	})

	return region, err
}

func initAction(c *cli.Context) error {
	groupName := c.String("group")
	var region regions.Region
	var err error
	switch c.String("region") {
	case "cn", "CN":
		region = regions.CN
	case "us", "US":
		region = regions.US
	case "tab", "TAB":
		region = regions.TAB
	case "":
		region, err = selectRegion()
		if err != nil {
			return newCliError(err)
		}
	default:
		return cli.NewExitError("invalid region", 1)
	}

	appList, err := api.GetAppList(region)
	if err != nil {
		return newCliError(err)
	}

	var orderedAppList []*api.GetAppListResult
	linq.From(appList).OrderBy(func(in interface{}) interface{} {
		return in.(*api.GetAppListResult).AppName[0]
	}).ToSlice(&orderedAppList)

	app, err := selectApp(orderedAppList)
	if err != nil {
		return newCliError(err)
	}

	groupList, err := api.GetGroups(app.AppID)
	if err != nil {
		return newCliError(err)
	}
	if groupName == "" {
		group, err := selectGroup(groupList)
		if err != nil {
			return newCliError(err)
		}
		groupName = group.GroupName
	} else {
		err = func() error {
			for _, group := range groupList {
				if group.GroupName == groupName {
					return nil
				}
			}
			return errors.New("找不到分组 " + groupName)
		}()
		if err != nil {
			return newCliError(err)
		}
	}

	boil, err := selectBoilerplate()
	if err != nil {
		return newCliError(err)
	}

	appName := app.AppName

	if err = boilerplate.FetchRepo(boil, appName, app.AppID); err != nil {
		return newCliError(err)
	}

	err = apps.LinkApp(app.AppName, app.AppID)
	if err != nil {
		return newCliError(err)
	}

	err = apps.LinkGroup(app.AppName, groupName)
	if err != nil {
		return newCliError(err)
	}

	return nil
}
