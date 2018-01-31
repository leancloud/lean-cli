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
	var selectedBoilerplate boilerplate.Boilerplate
	var selectedCategory boilerplate.Category
	categories, err := boilerplate.GetBoilerplates()
	if err != nil {
		return nil, err
	}

	question := wizard.Question{
		Content: "请选择需要创建应用的编程语言",
		Answers: []wizard.Answer{},
	}
	for _, category := range categories {
		answer := wizard.Answer{
			Content: category.Name,
		}
		func(category boilerplate.Category) {
			answer.Handler = func() {
				selectedCategory = category
			}
		}(category)
		question.Answers = append(question.Answers, answer)
	}
	err = wizard.Ask([]wizard.Question{question})

	question = wizard.Question{
		Content: "请选择需要创建的应用模版",
		Answers: []wizard.Answer{},
	}
	for _, boil := range selectedCategory.Boilerplates {
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
	err = wizard.Ask([]wizard.Question{question})
	return &selectedBoilerplate, err
}

func selectRegion(loginedRegions []regions.Region) (regions.Region, error) {
	region := regions.Invalid
	question := wizard.Question{
		Content: "请选择应用节点",
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
		loginedRegions := apps.GetLoginedRegions()
		if len(loginedRegions) == 0 {
			return cli.NewExitError("没有登录", 1)
		} else if len(loginedRegions) == 1 {
			region = loginedRegions[0]
		} else {
			region, err = selectRegion(loginedRegions)
			if err != nil {
				return err
			}
		}
	default:
		return cli.NewExitError("invalid region", 1)
	}

	appList, err := api.GetAppList(region)
	if err != nil {
		return err
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
			return errors.New("找不到分组 " + groupName)
		}()
		if err != nil {
			return err
		}
	}

	boil, err := selectBoilerplate()
	if err != nil {
		return err
	}

	var dest string

	if c.NArg() > 0 {
		dest = c.Args()[0]
	} else {
		dest = app.AppName
	}

	if err = boilerplate.FetchRepo(boil, dest, app.AppID); err != nil {
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
