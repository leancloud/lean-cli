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
			Content: func() string {
				if version.Distribution == "lean" {
					return "Paste AccessToken from LeanCloud Console => your App => LeanEngine => Deploy of your group => Deploy using CLI: "
				} else {
					return "Paste AccessToken from TapTap Developer Center => your Game => Game Services => Cloud Services => Cloud Engine => Deploy of your group => Deploy using CLI: "
				}
			}(),
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
	useToken := c.Bool("use-token")
	token := c.String("token")
	var region regions.Region
	var err error
	var userInfo *api.GetUserInfoResult

	if len(version.AvailableRegions) > 1 {
		if regionString == "" {
			region, err = selectRegion(version.AvailableRegions)
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
	} else {
		region = version.AvailableRegions[0]
	}

	if version.LoginViaAccessTokenOnly || useToken || token != "" {
		if token == "" {
			token, err = inputAcessToken()
			if err != nil {
				return err
			}
		}

		userInfo, err = loginWithAccessToken(token, region)
		if err != nil {
			return err
		}
	} else {
		userInfo, err = loginWithPassword(username, password, region)
		if err != nil {
			return err
		}
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
