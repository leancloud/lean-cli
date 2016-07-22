package main

import (
	"fmt"

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

func selectRuntime() int {
	runtimeType := 0

	wizard.Ask([]wizard.Question{
		{
			Content: "请选择项目语言:",
			Answers: []wizard.Answer{
				{
					Content: "Python",
					Handler: func() {
						runtimeType = boilerplate.Python
					},
				}, {
					Content: "Node.js",
					Handler: func() {
						runtimeType = boilerplate.NodeJS
					},
				},
				// {
				// 	Content: "PHP",
				// 	Handler: func() {
				// 		runtimeType = runtimePHP
				// 	},
				// },
			},
		},
	})
	return runtimeType
}

func initAction(*cli.Context) error {
	appList, err := api.GetAppList()
	if err != nil {
		return err
	}
	app := selectApp(appList)
	appID := app.AppID

	runtime := selectRuntime()

	appName := app.AppName

	if err := boilerplate.FetchRepo(runtime, appName, appID); err != nil {
		fmt.Println(err)
		return err
	}

	err = apps.LinkApp(app.AppName, app.AppID)

	return newCliError(err)
}
