package commands

import (
	"fmt"

	"github.com/aisk/chrysanthemum"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
)

func publishAction(c *cli.Context) error {
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		return cli.NewExitError("没有关联任何 app，请使用 lean checkout 来关联应用。", 1)
	}
	if err != nil {
		return newCliError(err)
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return newCliError(err)
	}

	spinner := chrysanthemum.New("获取应用信息").Start()
	region, err := api.GetAppRegion(appID)
	if err != nil {
		spinner.Failed()
		return newCliError(err)
	}
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		spinner.Failed()
		return newCliError(err)
	}
	engineInfo, err := api.GetEngineInfo(appID)
	if err != nil {
		spinner.Failed()
		return newCliError(err)
	}
	group, err := api.GetGroup(appID, groupName)
	if err != nil {
		spinner.Failed()
		return newCliError(err)
	}
	spinner.Successed()

	if engineInfo.Mode != "prod" {
		return cli.NewExitError("免费版应用使用 lean deploy 即可将代码部署到生产环境，无需使用此命令。", 1)
	}

	fmt.Printf("准备部署应用 %s(%s) 到 %s 节点分组 %s 生产环境\r\n", appInfo.AppName, appID, region, groupName)

	tok, err := api.DeployImage(appID, groupName, 1, group.StagingImage.ImageTag)
	ok, err := api.PollEvents(appID, tok)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}
