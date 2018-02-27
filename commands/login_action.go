package commands

import (
	"strings"

	"github.com/aisk/logp"
	"github.com/aisk/wizard"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
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

func loginWithPassword(username string, password string, region regions.Region) (*api.GetUserInfoResult, error) {
	if username == "" || password == "" {
		var err error
		username, password, err = inputAccountInfo()
		if err != nil {
			return nil, err
		}
	}
	r := region
	if r == regions.US {
		r = regions.CN
	}
	// US 节点登录时，需要先用用户名／密码登录 CN 节点，然后调用 api.LoginUSRegion 进行登录。
	info, err := api.Login(username, password, r)
	if err != nil {
		return nil, err
	}

	if region != regions.US {
		return info, nil
	}

	err = api.LoginUSRegion()
	if err != nil {
		return nil, err
	}

	return info, err
}

func loginAction(c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	regionStr := strings.ToUpper(c.String("region"))
	var region regions.Region
	switch regionStr {
	case "CN":
		region = regions.CN
	case "US":
		region = regions.US
	case "TAB":
		region = regions.TAB
	default:
		cli.ShowCommandHelp(c, "login")
		return cli.NewExitError("错误的 region 参数", 1)
	}
	userInfo, err := loginWithPassword(username, password, region)
	if err != nil {
		return err
	}
	_, err = api.GetAppList(region) // load region cache
	if err != nil {
		return err
	}
	logp.Info("登录成功：")
	logp.Infof("用户名: %s\r\n", userInfo.UserName)
	logp.Infof("邮箱: %s\r\n", userInfo.Email)
	return nil
}
