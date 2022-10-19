package commands

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aisk/logp"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/console"
	"github.com/leancloud/lean-cli/runtimes"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

var (
	errDoNotSupportCloudCode = cli.NewExitError(`This tool no long supports cloudcode 2.0 projects. Please update your project according to:
https://leancloud.cn/docs/leanengine_upgrade_3.html`, 1)
)

// get the console port. now console port is just runtime port plus one.
func getConsolePort(runtimePort int) int {
	return runtimePort + 1
}

func upAction(c *cli.Context) error {
	version.PrintVersionAndEnvironment()
	customArgs := c.Args()
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

	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return err
	}

	apiServer := api.GetAppAPIURL(region, appID)

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

	if rtm.Name == "cloudcode" {
		return errDoNotSupportCloudCode
	}

	logp.Info("Retrieving app info ...")
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return err
	}
	logp.Infof("Current app: %s (%s)\r\n", color.RedString(appInfo.AppName), appID)

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}
	groupInfo, err := api.GetGroup(appID, groupName)
	if err != nil {
		return err
	}

	haveStaging := "false"

	if groupInfo.Staging.Deployable {
		haveStaging = "true"
	}

	rtm.Envs = append(rtm.Envs, []string{
		"LC_APP_ID=" + appInfo.AppID,
		"LC_APP_KEY=" + appInfo.AppKey,
		"LC_APP_MASTER_KEY=" + appInfo.MasterKey,
		"LC_APP_PORT=" + strconv.Itoa(rtmPort),
		"LC_API_SERVER=" + apiServer,
		"LEANCLOUD_APP_ID=" + appInfo.AppID,
		"LEANCLOUD_APP_KEY=" + appInfo.AppKey,
		"LEANCLOUD_APP_MASTER_KEY=" + appInfo.MasterKey,
		"LEANCLOUD_APP_HOOK_KEY=" + appInfo.HookKey,
		"LEANCLOUD_APP_PORT=" + strconv.Itoa(rtmPort),
		"LEANCLOUD_API_SERVER=" + apiServer,
		"LEANCLOUD_APP_ENV=" + "development",
		"LEANCLOUD_REGION=" + region.EnvString(),
		"LEANCLOUD_APP_DOMAIN=" + groupInfo.Domain,
		"LEAN_CLI_HAVE_STAGING=" + haveStaging,
 		"LEANCLOUD_APP_GROUP=" + groupName,
	}...)

	if c.Bool("fetch-env") {
		for k, v := range groupInfo.Environments {
			localVar := os.Getenv(k)
			if localVar == "" {
				logp.Info("Exporting custome environment variables from LeanEngine: ", k)
				rtm.Envs = append(rtm.Envs, fmt.Sprintf("%s=%s", k, v))
			} else {
				logp.Info("Using local environment variables: ", k)
				rtm.Envs = append(rtm.Envs, fmt.Sprintf("%s=%s", k, localVar))
			}
		}
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
