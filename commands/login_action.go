package commands

import (
	"fmt"
	"strings"

	"github.com/aisk/wizard"
	"github.com/leancloud/lean-cli/api"
	"github.com/urfave/cli"
)

func inputAccountInfo() (string, string, error) {
	email := new(string)
	password := new(string)
	err := wizard.Ask([]wizard.Question{
		{
			Content: "请输入您的邮箱",
			Input: &wizard.Input{
				Result: email,
				Hidden: false,
			},
		},
		{
			Content: "请输入您的密码",
			Input: &wizard.Input{
				Result: password,
				Hidden: true,
			},
		},
	})

	return *email, *password, err
}

func loginWithPassword(username string, password string, region string) (*api.GetUserInfoResult, error) {
	if username == "" || password == "" {
		var err error
		username, password, err = inputAccountInfo()
		if err != nil {
			return nil, newCliError(err)
		}
	}
	info, err := api.Login(username, password)
	if err != nil {
		return nil, newCliError(err)
	}

	if region == "CN" {
		return info, nil
	}

	err = api.LoginUSRegion()
	if err != nil {
		return nil, newCliError(err)
	}

	return info, err
}

func loginWithSessionToken() (*api.GetUserInfoResult, error) {
	sessionToken := new(string)
	err := wizard.Ask([]wizard.Question{
		{
			Content: "请在浏览器打开：https://console.qcloud.com/tab?goto=cli-login-token，并输入页面给出的 token",
			Input: &wizard.Input{
				Result: sessionToken,
				Hidden: false,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return api.LoginTABRegion(*sessionToken)
}

func loginAction(c *cli.Context) error {
	region := strings.ToUpper(c.String("region"))

	var userInfo *api.GetUserInfoResult
	var err error

	switch region {
	case "CN":
		fallthrough
	case "US":
		username := c.String("username")
		password := c.String("password")
		userInfo, err = loginWithPassword(username, password, region)
		if err != nil {
			return newCliError(err)
		}
	case "TAB":
		userInfo, err = loginWithSessionToken()
		if err != nil {
			return newCliError(err)
		}
	default:
		cli.ShowCommandHelp(c, "login")
		return cli.NewExitError("", 1)
	}

	fmt.Println("登录成功：")
	fmt.Printf("用户名: %s\r\n", userInfo.UserName)
	fmt.Printf("邮箱: %s\r\n", userInfo.Email)
	return nil
}
