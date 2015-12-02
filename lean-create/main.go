package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/leancloud/lean-cli/ask"
	"github.com/leancloud/lean-cli/utils"
	"github.com/parnurzeal/gorequest"
)

type AppInfo struct {
	AppId     string `json:"app_id";`
	AppName   string `json:"app_name";`
	AppKey    string `json:"app_key";`
	MasterKey string `json:"master_key";`
}

func useExistApp() bool {
	var useExistApp bool
	ask.Ask([]ask.Question{
		{
			Content: "创建新的 LeanCloud APP 还是使用已存在的 LeanCloud APP？",
			Answers: []ask.Answer{
				{
					Content: "使用现有 LeanCloud APP",
					Handler: func() {
						useExistApp = true
					},
				},
				{
					Content: "创建新的 LeanCloud APP",
					Handler: func() {
						useExistApp = false
					},
				},
			},
		},
	})
	return useExistApp
}

func selectApp(appList []*AppInfo) *AppInfo {
	var selectedApp *AppInfo
	question := ask.Question{
		Content: "请选择 APP",
		Answers: []ask.Answer{},
	}
	for _, app := range appList {
		answer := ask.Answer{
			Content: app.AppName,
		}
		// for scope problem
		func (app *AppInfo) {
			answer.Handler = func() {
				selectedApp = app
			}
		}(app)
		question.Answers = append(question.Answers, answer)
	}
	ask.Ask([]ask.Question{question})
	return selectedApp
}

func getAppList() ([]*AppInfo, error) {
	cookies, err := utils.GetCookies()
	if err != nil {
		return nil, err
	}

	request := gorequest.New()
	request.SetDebug(false)
	resp, body, errs := request.Get("https://leancloud.cn/1/clients/self/apps").
		AddCookies(cookies).
		Set("User-Agent", "leanengine-cli x.x.x"). // TODO
		End()

	if len(errs) != 0 {
		return nil, errs[0]
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(utils.FormatServerErrorResult(body))
	}

	var appList []*AppInfo

	json.Unmarshal([]byte(body), &appList)
	return appList, nil
}

func createNewApp() {

}

func getAppInfo() {

}

func main() {
	var app *AppInfo
	var err error
	if useExistApp() {
		var appList []*AppInfo
		appList, err = getAppList()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		app = selectApp(appList)

	} else {
		// TODO
	}
	fmt.Println(app)

}
