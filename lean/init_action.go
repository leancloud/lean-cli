package main

import (
	"github.com/aisk/wizard"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/boilerplate"
)

func selectApp(appList []*api.GetAppListResult) *api.GetAppListResult {
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
	wizard.Ask([]wizard.Question{question})
	return selectedApp
}

func selectBoilerplate() (*boilerplate.Boilerplate, error) {
	var selectBoil *boilerplate.Boilerplate
	boils, err := boilerplate.GetBoilerplateList()
	if err != nil {
		return nil, err
	}

	question := wizard.Question{
		Content: "请选择需要创建的应用模版：",
		Answers: []wizard.Answer{},
	}
	for _, boil := range boils {
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
	wizard.Ask([]wizard.Question{question})
	return selectBoil, nil
}

func initAction(*cli.Context) error {
	appList, err := api.GetAppList()
	if err != nil {
		return newCliError(err)
	}
	app := selectApp(appList)
	appID := app.AppID

	boil, err := selectBoilerplate()
	if err != nil {
		return newCliError(err)
	}

	appName := app.AppName

	if err := boilerplate.FetchRepo(boil, appName, appID); err != nil {
		return newCliError(err)
	}

	err = apps.LinkApp(app.AppName, app.AppID)
	if err != nil {
		return newCliError(err)
	}
	return nil
}
