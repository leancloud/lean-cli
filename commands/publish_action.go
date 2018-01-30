package commands

import (
	"errors"

	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

func publishAction(c *cli.Context) error {
	version.PrintCurrentVersion()
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		return cli.NewExitError("没有关联任何 app，请使用 lean checkout 来关联应用。", 1)
	}
	if err != nil {
		return err
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}

	logp.Info("获取应用信息 ...")
	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return err
	}
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return err
	}
	engineInfo, err := api.GetEngineInfo(appID)
	if err != nil {
		return err
	}
	group, err := api.GetGroup(appID, groupName)
	if err != nil {
		return err
	}

	if engineInfo.Mode != "prod" {
		return errors.New("免费版应用使用 lean deploy 即可将代码部署到生产环境，无需使用此命令")
	}

	logp.Infof("准备部署应用 %s(%s) 到 %s 节点分组 %s 生产环境\r\n", appInfo.AppName, appID, region, groupName)

	var deployMode string

	if c.Bool("atomic") {
		deployMode = api.DEPLOY_SMOOTHLY
	} else {
		deployMode = api.DEPLOY_SMOOTHLY
	}

	tok, err := api.DeployImage(appID, groupName, 1, group.StagingImage.ImageTag, deployMode)
	ok, err := api.PollEvents(appID, tok)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("部署失败")
	}
	return nil
}
