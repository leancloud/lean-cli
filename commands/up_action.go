package commands

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aisk/chrysanthemum"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/console"
	"github.com/leancloud/lean-cli/runtimes"
	"github.com/urfave/cli"
)

var (
	errDoNotSupportCloudCode = cli.NewExitError(`命令行工具不再支持 cloudcode 2.0 项目，请参考此文档对您的项目进行升级：
https://leancloud.cn/docs/leanengine_upgrade_3.html`, 1)
)

// get the console port. now console port is just runtime port plus one.
func getConsolePort(runtimePort int) int {
	return runtimePort + 1
}

func upAction(c *cli.Context) error {
	customArgs := c.Args()
	watchChanges := c.Bool("watch")
	customCommand := c.String("cmd")
	rtmPort := c.Int("port")
	consPort := c.Int("console-port")
	if consPort == 0 {
		consPort = getConsolePort(rtmPort)
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	region, err := api.GetAppRegion(appID)
	if err != nil {
		return err
	}

	rtm, err := runtimes.DetectRuntime("")
	if err != nil {
		return err
	}
	rtm.Port = strconv.Itoa(rtmPort)
	rtm.Args = append(rtm.Args, customArgs...)
	if customCommand != "" {
		customCommand = strings.TrimSpace(customCommand)
		cmds := regexp.MustCompile(" +").Split(customCommand, -1)
		rtm.Exec = cmds[0]
		rtm.Args = cmds[1:]
	}

	if watchChanges {
		printDeprecatedWatchWarning(rtm)
	}

	if rtm.Name == "cloudcode" {
		return errDoNotSupportCloudCode
	}

	bar := chrysanthemum.New("获取应用信息").Start()
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		bar.Failed()
		return err
	}
	bar.Successed()
	fmt.Printf("当前应用：%s (%s)\r\n", color.RedString(appInfo.AppName), appID)

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}
	spinner := chrysanthemum.New("获取运引擎分组 " + groupName + " 信息").Start()
	groupInfo, err := api.GetGroup(appID, groupName)
	if err != nil {
		spinner.Failed()
		return err
	}
	spinner.Successed()

	engineInfo, err := api.GetEngineInfo(appID)
	if err != nil {
		return err
	}
	haveStaging := "false"
	if engineInfo.Mode == "prod" {
		haveStaging = "true"
	}

	rtm.Envs = []string{
		"LC_APP_ID=" + appInfo.AppID,
		"LC_APP_KEY=" + appInfo.AppKey,
		"LC_APP_MASTER_KEY=" + appInfo.MasterKey,
		"LC_APP_PORT=" + strconv.Itoa(rtmPort),
		"LC_API_SERVER=" + region.APIServerURL(),
		"LEANCLOUD_APP_ID=" + appInfo.AppID,
		"LEANCLOUD_APP_KEY=" + appInfo.AppKey,
		"LEANCLOUD_APP_MASTER_KEY=" + appInfo.MasterKey,
		"LEANCLOUD_APP_HOOK_KEY=" + appInfo.HookKey,
		"LEANCLOUD_APP_PORT=" + strconv.Itoa(rtmPort),
		"LEANCLOUD_API_SERVER=" + region.APIServerURL(),
		"LEANCLOUD_APP_ENV=" + "development",
		"LEANCLOUD_REGION=" + region.String(),
		"LEAN_CLI_HAVE_STAGING=" + haveStaging,
	}

	for k, v := range groupInfo.Environments {
		chrysanthemum.Successed("从服务器导出自定义环境变量:", k)
		rtm.Envs = append(rtm.Envs, fmt.Sprintf("%s=%s", k, v))
	}

	cons := &console.Server{
		AppID:       appInfo.AppID,
		AppKey:      appInfo.AppKey,
		MasterKey:   appInfo.MasterKey,
		HookKey:     appInfo.HookKey,
		RemoteURL:   "http://localhost:" + strconv.Itoa(rtmPort),
		ConsolePort: strconv.Itoa(consPort),
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

func printDeprecatedWatchWarning(rtm *runtimes.Runtime) {
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
