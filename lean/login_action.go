package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aisk/cookieparser"
	"github.com/aisk/wizard"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/utils"
	"github.com/levigross/grequests"
)

func inputAccountInfo() (string, string) {
	var username = new(string)
	var password = new(string)
	wizard.Ask([]wizard.Question{
		{
			Content: "请输入您的邮箱：",
			Input: &wizard.Input{
				Result: username,
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
	return *username, *password
}

func login(email string, password string) ([]*http.Cookie, error) {
	options := &grequests.RequestOptions{
		JSON: map[string]string{
			"email":    email,
			"password": password,
		},
	}
	response, err := grequests.Post("https://leancloud.cn/1/signin", options)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, api.NewErrorFromBody(response.String())
	}
	return response.RawResponse.Cookies(), err
}

func saveCookie(cookies []*http.Cookie) error {
	os.Mkdir(filepath.Join(utils.HomeDir(), ".leancloud"), 0700)

	content := []byte(cookieparser.ToString(cookies))
	return ioutil.WriteFile(filepath.Join(utils.HomeDir(), ".leancloud", "cookies"), content, 0600)
}

func loginAction(c *cli.Context) error {
	email, password := inputAccountInfo()
	cookies, err := login(email, password)
	if err != nil {
		switch e := err.(type) {
		case api.Error:
			return cli.NewExitError(e.Content, 1)
		default:
			return cli.NewExitError(e.Error(), 1)
		}
	}
	err = saveCookie(cookies)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	log.Println("登录成功。")
	return nil
}
