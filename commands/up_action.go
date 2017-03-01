package commands

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/console"
	"github.com/leancloud/lean-cli/runtimes"
)

// get the console port. now console port is just runtime port plus one.
func getConsolePort(runtimePort string) (string, error) {
	port, err := strconv.Atoi(runtimePort)
	if err != nil {
		return "", nil
	}
	return strconv.Itoa(port + 1), nil
}

func upAction(c *cli.Context) error {
	watchChanges := c.Bool("watch")
	port := strconv.Itoa(c.Int("port"))
	consPort, err := getConsolePort(port)
	if err != nil {
		return newCliError(err)
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return newCliError(err)
	}

	region, err := api.GetAppRegion(appID)
	if err != nil {
		return newCliError(err)
	}

	rtm, err := runtimes.DetectRuntime("")
	if err != nil {
		return newCliError(err)
	}
	rtm.Port = port

	if watchChanges {
		fmt.Fprintf(
			color.Output,
			" %s [WARNING] --watch 选项不再被支持，请使用项目代码本身实现此功能\r\n",
			chrysanthemum.Fail,
		)
		if rtm.Name == "python" {
			fmt.Println("   [WARNING] 可以参考此 Pull Request 来给现有项目增加调试时自动重启功能：")
			fmt.Println("   [WARNING] https://github.com/leancloud/python-getting-started/pull/12/files")
		}
		if rtm.Name == "node.js" {
			fmt.Println("   [WARNING] 可以参考此 Pull Request 来给现有项目增加调试时自动重启功能：")
			fmt.Println("   [WARNING] https://github.com/leancloud/node-js-getting-started/pull/26/files")
		}
	}

	if rtm.Name == "cloudcode" {
		return cli.NewExitError(`命令行工具不再支持 cloudcode 2.0 项目，请参考此文档对您的项目进行升级：
https://leancloud.cn/docs/leanengine_upgrade_3.html`, 1)
	}

	bar := chrysanthemum.New("获取应用信息").Start()
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	bar.Successed()
	chrysanthemum.Printf("当前应用：%s (%s)\r\n", color.RedString(appInfo.AppName), appID)

	bar = chrysanthemum.New("获取云引擎信息").Start()
	engineInfo, err := api.GetEngineInfo(appID)
	if err != nil {
		bar.Failed()
		return newCliError(err)
	}
	bar.Successed()

	rtm.Envs = []string{
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

	for k, v := range engineInfo.Environments {
		fmt.Println("   从服务器导出自定义环境变量:", k)
		rtm.Envs = append(rtm.Envs, fmt.Sprintf("%s=%s", k, v))
	}

	cons := &console.Server{
		AppID:       appInfo.AppID,
		AppKey:      appInfo.AppKey,
		MasterKey:   appInfo.MasterKey,
		AppPort:     port,
		ConsolePort: consPort,
		Errors:      make(chan error),
	}

	rtm.Run()
	time.Sleep(time.Millisecond * 300)
	cons.Run()

	for {
		select {
		case err = <-cons.Errors:
			panic(err)
		case err = <-rtm.Errors:
			if _, ok := err.(*exec.ExitError); ok {
				return cli.NewExitError("", 1)
			}
			panic(err)
		}
	}
}
