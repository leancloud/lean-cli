package main

import (
	"log"

	"github.com/aisk/wizard"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
)

func inputAccountInfo() (string, string) {
	var email = new(string)
	var password = new(string)
	wizard.Ask([]wizard.Question{
		{
			Content: "请输入您的邮箱：",
			Input: &wizard.Input{
				Result: email,
				Hidden: false,
			},
		},
		{
			Content: "请输入您的密码：",
			Input: &wizard.Input{
				Result: password,
				Hidden: true,
			},
		},
	})
	return *email, *password
}

func loginAction(c *cli.Context) error {
	email, password := inputAccountInfo()
	err := api.Login(email, password)
	if err != nil {
		switch e := err.(type) {
		case api.Error:
			return cli.NewExitError(e.Content, 1)
		default:
			return cli.NewExitError(e.Error(), 1)
		}
	}
	log.Println("登录成功。")
	return nil
}
