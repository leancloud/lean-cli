package commands

import (
	"os"

	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
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

	chrysanthemum.Printf("准备部署至目标分组：%s\r\n", color.RedString(groupName))

	tok, err := api.DeployImage(appID, groupName, 1, group.StagingImage.ImageTag)
	ok, err := api.PollEvents(appID, tok, os.Stdout)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}
