package commands

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/aisk/chrysanthemum"
	"github.com/cbroglie/mustache"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
)

var (
	defaultBashEnvTemplateString = "export {{name}}={{value}}"
	defaultDOSEnvTemplateString  = "SET {{name}}={{value}}"
)

func detectDOS() bool {
	// this function is not reliable
	if runtime.GOOS != "windows" {
		return false
	}
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "bash") ||
		strings.Contains(shell, "zsh") ||
		strings.Contains(shell, "fish") ||
		strings.Contains(shell, "csh") ||
		strings.Contains(shell, "ksh") ||
		strings.Contains(shell, "ash") {
		return false
	}
	return true
}

func envAction(c *cli.Context) error {
	port := strconv.Itoa(c.Int("port"))
	tmplString := c.String("template")
	if tmplString == "" {
		if detectDOS() {
			tmplString = defaultDOSEnvTemplateString
		} else {
			tmplString = defaultBashEnvTemplateString
		}
	}

	tmpl, err := mustache.ParseString(tmplString)
	if err != nil {
		return err
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	region, err := api.GetAppRegion(appID)
	if err != nil {
		return err
	}

	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return err
	}

	envs := []map[string]string{
		map[string]string{"name": "LC_APP_ID", "value": appInfo.AppID},
		map[string]string{"name": "LC_APP_KEY", "value": appInfo.AppKey},
		map[string]string{"name": "LC_APP_MASTER_KEY", "value": appInfo.MasterKey},
		map[string]string{"name": "LC_APP_PORT", "value": port},
		map[string]string{"name": "LC_API_SERVER", "value": region.APIServerURL()},
		map[string]string{"name": "LEANCLOUD_APP_ID", "value": appInfo.AppID},
		map[string]string{"name": "LEANCLOUD_APP_KEY", "value": appInfo.AppKey},
		map[string]string{"name": "LEANCLOUD_APP_MASTER_KEY", "value": appInfo.MasterKey},
		map[string]string{"name": "LEANCLOUD_APP_HOOK_KEY", "value": appInfo.HookKey},
		map[string]string{"name": "LEANCLOUD_APP_PORT", "value": port},
		map[string]string{"name": "LEANCLOUD_API_SERVER", "value": region.APIServerURL()},
		map[string]string{"name": "LEANCLOUD_APP_ENV", "value": "development"},
		map[string]string{"name": "LEANCLOUD_REGION", "value": region.String()},
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}
	groupInfo, err := api.GetGroup(appID, groupName)
	if err != nil {
		return err
	}

	for name, value := range groupInfo.Environments {
		envs = append(envs, map[string]string{"name": name, "value": value})
	}

	for _, env := range envs {
		result, err := tmpl.Render(env)
		if err != nil {
			return err
		}
		fmt.Println(result)
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
		return errors.New("请不要设置 `LEANCLOUD` 开头的环境变量")
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	bar := chrysanthemum.New("获取云引擎信息").Start()
	engineInfo, err := api.GetEngineInfo(appID)
	if err != nil {
		bar.Failed()
		return err
	}
	group, err := apps.GetCurrentGroup(".")
	if err != nil {
		bar.Failed()
		return err
	}
	bar.Successed()

	envs := engineInfo.Environments
	envs[envName] = envValue
	bar = chrysanthemum.New("更新云引擎 " + group + " 分组环境变量").Start()
	err = api.PutEnvironments(appID, group, envs)
	if err != nil {
		bar.Failed()
		return err
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
		return errors.New("请不要移除 `LEANCLOUD` 开头的环境变量")
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	bar := chrysanthemum.New("获取云引擎信息").Start()
	group, err := apps.GetCurrentGroup(".")
	if err != nil {
		bar.Failed()
		return err
	}
	engineInfo, err := api.GetEngineInfo(appID)
	if err != nil {
		bar.Failed()
		return err
	}
	bar.Successed()

	envs := engineInfo.Environments
	delete(envs, env)

	bar = chrysanthemum.New("更新云引擎 " + group + " 分组环境变量").Start()
	err = api.PutEnvironments(appID, group, envs)
	if err != nil {
		bar.Failed()
		return err
	}
	bar.Successed()
	return nil
}
