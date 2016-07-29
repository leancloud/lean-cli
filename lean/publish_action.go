package main

import (
	"errors"
	"log"
	"os"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/codegangsta/cli"
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

	group, found, err := linq.From(groups).Where(func(group linq.T) (bool, error) {
		return group.(*api.GetGroupsResult).Prod == env, nil
	}).First()
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("group not found")
	}
	return group.(*api.GetGroupsResult), nil
}

func publishAction(c *cli.Context) error {
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		log.Fatalln("没有关联任何 app，请使用 lean checkout 来关联应用。")
	}
	if err != nil {
		return newCliError(err)
	}

	op.Write("获取应用信息")
	info, err := api.GetAppInfo(appID)
	if err != nil {
		op.Failed()
		return newCliError(err)
	}
	op.Successed()

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
