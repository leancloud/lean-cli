package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
)

func envAction(c *cli.Context) error {
	port := strconv.Itoa(c.Int("port"))

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return newCliError(err)
	}

	region, err := api.GetAppRegion(appID)
	if err != nil {
		return newCliError(err)
	}

	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return newCliError(err)
	}

	envs := []string{
		"LC_APP_ID=" + appInfo.AppID,
		"LC_APP_KEY=" + appInfo.AppKey,
		"LC_APP_MASTER_KEY=" + appInfo.MasterKey,
		"LC_APP_PORT=" + port,
		"LC_API_SERVER=" + region.APIServerURL(),
		"LEANCLOUD_APP_ID=" + appInfo.AppID,
		"LEANCLOUD_APP_KEY=" + appInfo.AppKey,
		"LEANCLOUD_APP_MASTER_KEY=" + appInfo.MasterKey,
		"LEANCLOUD_APP_HOOK_KEY=" + appInfo.HookKey,
		"LEANCLOUD_APP_PORT=" + port,
		"LEANCLOUD_API_SERVER=" + region.APIServerURL(),
		"LEANCLOUD_APP_ENV=" + "development",
		"LEANCLOUD_REGION=" + region.String(),
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return newCliError(err)
	}
	spinner := chrysanthemum.New("获取运引擎分组 " + groupName + " 信息").Start()
	groupInfo, err := api.GetGroup(appID, groupName)
	if err != nil {
		spinner.Failed()
		return newCliError(err)
	}
	spinner.Successed()

	for k, v := range groupInfo.Environments {
		envs = append(envs, k+"="+v)
	}

	for _, env := range envs {
		fmt.Println("export", env)
	}

	return nil
}

func envSetAction(c *cli.Context) error {
	if c.NArg() != 2 {
		cli.ShowSubcommandHelp(c)
		return cli.NewExitError("", 1)
	}
	envName := c.Args()[0]
	envValue := c.Args()[1]

	if strings.HasPrefix(strings.ToUpper(envName), "LEANCLOUD") {
		return newCliError(errors.New("请不要设置 `LEANCLOUD` 开头的环境变量"))
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return newCliError(err)
	}

	bar := chrysanthemum.New("获取云引擎信息").Start()
	engineInfo, err := api.GetEngineInfo(appID)
	if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	bar.Successed()

	envs := engineInfo.Environments
	envs[envName] = envValue
	bar = chrysanthemum.New("更新云引擎环境变量").Start()
	err = api.PutEnvironments(appID, envs)
	if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	bar.Successed()
	return nil
}

func envUnsetAction(c *cli.Context) error {
	if c.NArg() != 1 {
		cli.ShowSubcommandHelp(c)
		return cli.NewExitError("", 1)
	}
	env := c.Args()[0]

	if strings.HasPrefix(strings.ToUpper(env), "LEANCLOUD") {
		return newCliError(errors.New("请不要移除 `LEANCLOUD` 开头的环境变量"))
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return newCliError(err)
	}

	bar := chrysanthemum.New("获取云引擎信息").Start()
	engineInfo, err := api.GetEngineInfo(appID)
	if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	bar.Successed()

	envs := engineInfo.Environments
	delete(envs, env)

	bar = chrysanthemum.New("更新云引擎环境变量").Start()
	err = api.PutEnvironments(appID, envs)
	if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	bar.Successed()
	return nil
}
