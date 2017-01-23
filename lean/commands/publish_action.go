package commands

import (
	"os"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
)

const (
	stag = 0
	prod = 1
)

func getDefaultGroup(appID string, env int) (*api.GetGroupsResult, error) {
	if env != stag && env != prod {
		panic("Invalid prod params")
	}
	groups, err := api.GetGroups(appID)
	if err != nil {
		return nil, err
	}

	group := linq.From(groups).Where(func(group interface{}) bool {
		return group.(*api.GetGroupsResult).Prod == env
	}).First()
	if err != nil {
		return nil, err
	}
	return group.(*api.GetGroupsResult), nil
}

func publishAction(c *cli.Context) error {
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		return cli.NewExitError("没有关联任何 app，请使用 lean checkout 来关联应用。", 1)
	}
	if err != nil {
		return newCliError(err)
	}

	spinner := chrysanthemum.New("获取应用信息").Start()
	info, err := api.GetAppInfo(appID)
	if err != nil {
		spinner.Failed()
		return newCliError(err)
	}
	spinner.Successed()
	chrysanthemum.Printf("准备部署至目标应用：%s (%s)\r\n", color.RedString(info.AppName), appID)

	if info.LeanEngineMode == "free" {
		return cli.NewExitError("免费版应用使用 lean deploy 即可将代码部署到生产环境，无需使用此命令。", 1)
	}

	prodGroup, err := getDefaultGroup(appID, prod)
	if err != nil {
		return newCliError(err)
	}
	stagGroup, err := getDefaultGroup(appID, stag)
	if err != nil {
		return newCliError(err)
	}

	tok, err := api.DeployImage(appID, prodGroup.GroupName, stagGroup.CurrentImage.ImageTag)
	ok, err := api.PollEvents(appID, tok, os.Stdout)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}
