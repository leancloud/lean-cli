package main

import (
	"fmt"
	"log"

	"github.com/aisk/wizard"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/utils"
)

func appRemoveAction(c *cli.Context) {
	listOfApps, err := apps.LinkedApps("")
	utils.CheckError(err)
	if len(listOfApps) == 0 {
		log.Fatal("没有关联任何应用")
	}

	answers := []wizard.Answer{}
	appToRemove := new(string)
	for _, app := range listOfApps {
		answers = append(answers, wizard.Answer{
			Content: fmt.Sprintf("%s - %s", app.AppName, app.AppID),
			Handler: func() {
				appToRemove = &app.AppName
			},
		})
	}

	wizard.Ask([]wizard.Question{
		{
			Content: "请选择要移除关联的应用：",
			Answers: answers,
		},
	})

	log.Println("移除关联的应用：", *appToRemove)
	apps.RemoveApp("", *appToRemove)
}
