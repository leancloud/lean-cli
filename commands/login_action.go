package commands

import (
	"github.com/aisk/logp"
	"github.com/aisk/wizard"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/urfave/cli"
)

func inputRegion(c *cli.Context, interactiveCllback func(c *cli.Context) (regions.Region, error)) (regions.Region, error) {
	region := c.String("region")

	switch region {
	case "cn", "CN", "cn-n1":
		return regions.ChinaNorth, nil
	case "tab", "TAB", "cn-e1":
		return regions.ChinaEast, nil
	case "us", "US", "us-w1":
		return regions.USWest, nil
	case "":
		return interactiveCllback(c)
	default:
		cli.ShowCommandHelp(c, "login")
		return regions.Invalid, cli.NewExitError("Wrong region parameter", 1)
	}
}

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
	region, err := inputRegion(c, func(c *cli.Context) (regions.Region, error) {
		regionSelected, err := selectRegion([]regions.Region{regions.ChinaNorth, regions.USWest, regions.ChinaEast})
		if err != nil {
			return regions.Invalid, err
		}
		return regionSelected, nil
	})
	if err != nil {
		return err
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
