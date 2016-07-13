package main

import (
	"log"

	"github.com/aisk/wizard"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/boilerplate"
	"github.com/leancloud/lean-cli/lean/utils"
)

func askNewAppInfo() (string, string, int) {
	appID := new(string)
	masterKey := new(string)
	runtimeType := 0

	log.Println("开始输入应用信息，这些信息可以从'开发者平台的应用设置 -> 应用 key'里找到。")

	wizard.Ask([]wizard.Question{
		{
			Content: "请输入应用的 Application ID:",
			Input: &wizard.Input{
				Hidden: false,
				Result: appID,
			},
		},
	})

	// TODO: get the masterKey from local first
	wizard.Ask([]wizard.Question{
		{
			Content: "请输入应用的 Master Key:",
			Input: &wizard.Input{
				Hidden: true,
				Result: masterKey,
			},
		},
	})

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
	return *appID, *masterKey, runtimeType
}

func newAction(*cli.Context) {
	appID, masterKey, runtime := askNewAppInfo()

	client := api.NewKeyAuthClient(appID, masterKey)

	detail, err := client.AppDetail()
	utils.CheckError(err)
	appName := detail.Get("app_name").MustString()

	err = boilerplate.FetchRepo(runtime, appName, appID)
	utils.CheckError(err)
}
