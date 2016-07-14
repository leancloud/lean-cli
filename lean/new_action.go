package main

import (
	"fmt"

	"github.com/aisk/wizard"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/boilerplate"
	"github.com/leancloud/lean-cli/lean/utils"
)

func selectApp(appList []interface{}) map[string]interface{} {
	var selectedApp map[string]interface{}
	question := wizard.Question{
		Content: "请选择 APP",
		Answers: []wizard.Answer{},
	}
	for _, _app := range appList {
		app := _app.(map[string]interface{})
		answer := wizard.Answer{
			Content: app["app_name"].(string),
		}
		// for scope problem
		func(app map[string]interface{}) {
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

func newAction(*cli.Context) error {
	appList, err := api.GetAppList()
	if err != nil {
		return err
	}
	app := selectApp(appList)
	appID := app["app_id"].(string)
	masterKey := app["master_key"].(string)

	runtime := selectRuntime()

	client := api.NewKeyAuthClient(appID, masterKey)

	detail, err := client.AppDetail()
	utils.CheckError(err)
	appName := detail.Get("app_name").MustString()

	if err := boilerplate.FetchRepo(runtime, appName, appID); err != nil {
		return err
	}

	if err := apps.AddApp(appName, appName, appID); err != nil {
		fmt.Println(err)
		return err
	}

	if err := apps.SwitchApp(appName, appName); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
