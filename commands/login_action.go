package commands

import (
	"github.com/aisk/logp"
	"github.com/aisk/wizard"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/version"
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

func inputAcessToken() (string, error) {
	accessToken := new(string)
	err := wizard.Ask([]wizard.Question{
		{
			Content: "AccessToken: ",
			Input: &wizard.Input{
				Result: accessToken,
				Hidden: false,
			},
		},
	})

	return *accessToken, err
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

func loginWithAccessToken(token string, region regions.Region) (*api.GetUserInfoResult, error) {
	if token == "" {
		var err error
		token, err = inputAcessToken()
		if err != nil {
			return nil, err
		}
	}

	return api.LoginWithAccessToken(token, region)
}

func loginAction(c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	regionString := c.String("region")
	var region regions.Region
	var err error
	var userInfo *api.GetUserInfoResult
	if version.Distribution == "lean" {
		if regionString == "" {
			region, err = selectRegion([]regions.Region{regions.ChinaNorth, regions.USWest, regions.ChinaEast})
			if err != nil {
				return err
			}
		} else {
			region = regions.Parse(regionString)
		}

		if region == regions.Invalid {
			cli.ShowCommandHelp(c, "login")
			return cli.NewExitError("Wrong region parameter", 1)
		}

		userInfo, err = loginWithPassword(username, password, region)
		if err != nil {
			return err
		}
	} else {
		region = regions.ChinaNorth
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
