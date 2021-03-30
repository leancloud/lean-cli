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

func inputAcessKey() (string, error) {
	accessKey := new(string)
	err := wizard.Ask([]wizard.Question{
		{
			Content: "AccessKey: ",
			Input: &wizard.Input{
				Result: accessKey,
				Hidden: false,
			},
		},
	})

	return *accessKey, err
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

func loginWithAccessKey(key string, region regions.Region) (*api.GetUserInfoResult, error) {
	if key == "" {
		var err error
		key, err = inputAcessKey()
		if err != nil {
			return nil, err
		}
	}

	return api.LoginNext(key, region)
}

func loginAction(c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	regionString := c.String("region")
	var region regions.Region
	var err error
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

func loginActionNext(c *cli.Context) error {
	accessKey := c.String("key")
	region := regions.CN
	/*
		if strings.HasSuffix(accessKey, "-cn") {
			region = regions.CN
		} else if strings.HasSuffix(accessKey, "-us") {
			region = regions.US
		} else {
			return cli.NewExitError("Bad format of AccessKey", 1)
		}
	*/
	userInfo, err := loginWithAccessKey(accessKey, region)
	if err != nil {
		return err
	}

	logp.Info("Login succeeded: ")
	logp.Infof("Username: %s\r\n", userInfo.UserName)
	logp.Infof("Email: %s\r\n", userInfo.Email)

	return nil
}
