package main

import (
	"log"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
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

	appID, err := apps.GetCurrentAppID("")
	if err != nil {
		return newCliError(err)
	}

	rtm, err := console.DetectRuntime("")
	log.Println(rtm, err)
	if err != nil {
		return newCliError(err)
	}

	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return newCliError(err)
	}

	envs := []string{
		"LC_APP_ID=" + appInfo.AppID,
		"LC_APP_KEY" + appInfo.AppKey,
		"LC_APP_MASTER_KEY" + appInfo.MasterKey,
		"LC_APP_PORT" + port,
		"LC_API_SERVER" + apiServerURL,
		"LEANCLOUD_APP_ID" + appInfo.AppID,
		"LEANCLOUD_APP_KEY" + appInfo.AppKey,
		"LEANCLOUD_APP_MASTER_KEY" + appInfo.MasterKey,
		"LEANCLOUD_APP_PORT" + port,
		"LEANCLOUD_API_SERVER" + apiServerURL,
	}
	for _, env := range envs {
		rtm.Envs = append(envs, env)
	}

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
