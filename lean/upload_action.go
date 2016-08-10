package main

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

func uploadAction(c *cli.Context) error {
	if c.NArg() < 1 {
		cli.ShowCommandHelp(c, "upload")
		return cli.NewExitError("", 1)
	}

	filePath := c.Args().First()
	log.Println("> 准备上传文件：" + color.RedString(filePath))

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return newCliError(err)
	}

	file, err := api.UploadFile(appID, filePath)
	if err != nil {
		return newCliError(err)
	}

	log.Println("> 上传成功，文件 URL：" + file.URL)

	return nil
}
