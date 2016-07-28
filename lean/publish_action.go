package main

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func publishAction(c *cli.Context) error {
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		log.Fatalln("没有关联任何 app，请使用 lean switch 来关联应用。")
	}
	if err != nil {
		return newCliError(err)
	}

	info, err := api.GetAppInfo(appID)
	if err != nil {
		return newCliError(err)
	}

	if info.LeanEngineMode == "free" {
		return cli.NewExitError("免费版应用使用 lean deploy 即可将代码部署到生产环境，无需使用此命令。", 1)
	}

	return nil
}
