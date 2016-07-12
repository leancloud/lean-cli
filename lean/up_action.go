package main

import (
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/console"
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
	// TODO: get port from args
	port := "3000"
	consPort, err := getConsolePort(port)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	// TODO:
	apiServerURL := "https://api.leancloud.cn"

	appInfo, err := apps.CurrentAppInfo(".")
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	rtm, err := apps.DetectRuntime(".")
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	rtm.Envs["LC_APP_ID"] = appInfo.AppID
	rtm.Envs["LC_APP_KEY"] = appInfo.AppKey
	rtm.Envs["LC_APP_MASTER_KEY"] = appInfo.MasterKey
	rtm.Envs["LC_APP_PORT"] = port
	rtm.Envs["LC_API_SERVER"] = apiServerURL
	rtm.Envs["LEANCLOUD_APP_ID"] = appInfo.AppID
	rtm.Envs["LEANCLOUD_APP_KEY"] = appInfo.AppKey
	rtm.Envs["LEANCLOUD_APP_MASTER_KEY"] = appInfo.MasterKey
	rtm.Envs["LEANCLOUD_APP_PORT"] = port
	rtm.Envs["LEANCLOUD_API_SERVER"] = apiServerURL

	go func() {
		err := rtm.Run()
		if err != nil {
			panic(err)
		}
	}()

	cons := &console.Server{
		AppID:       appInfo.AppID,
		AppKey:      appInfo.AppKey,
		MasterKey:   appInfo.MasterKey,
		AppPort:     port,
		ConsolePort: consPort,
	}

	cons.Run()
	return nil
}
