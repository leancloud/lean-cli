package commands

import (
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
			Content: "Email: ",
			Input: &wizard.Input{
				Result: email,
				Hidden: false,
			},
		},
		{
			Content: "Password: ",
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

	return api.Login(username, password, region)
}

func loginAction(c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	regionStr := c.String("region")
	var region regions.Region
	var err error
	switch regionStr {
	case "cn", "CN", "cn-n1":
		region = regions.CN
	case "us", "US", "us-w1":
		region = regions.US
	case "tab", "TAB", "cn-e1":
		region = regions.TAB
	case "":
		region, err = selectRegion([]regions.Region{regions.CN, regions.US, regions.TAB})
		if err != nil {
			return err
		}
	default:
		cli.ShowCommandHelp(c, "login")
		return cli.NewExitError("Wrong region parameter", 1)
	}
	userInfo, err := loginWithPassword(username, password, region)
	if err != nil {
		return err
	}
	_, err = api.GetAppList(region) // load region cache
	if err != nil {
		return err
	}
	logp.Info("Login succeeded: ")
	logp.Infof("Username: %s\r\n", userInfo.UserName)
	logp.Infof("Email: %s\r\n", userInfo.Email)
	return nil
}
