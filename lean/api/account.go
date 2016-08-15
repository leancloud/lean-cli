package api

import (
	"os"
	"path/filepath"

	"github.com/bitly/go-simplejson"
	"github.com/juju/persistent-cookiejar"
	"github.com/leancloud/lean-cli/lean/utils"
	"github.com/levigross/grequests"
)

// Login LeanCloud account
func Login(email string, password string) (*simplejson.Json, error) {
	os.MkdirAll(filepath.Join(utils.ConfigDir(), "leancloud"), 0700)
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: filepath.Join(utils.ConfigDir(), "leancloud", "cookies"),
	})
	if err != nil {
		return nil, err
	}

	options := &grequests.RequestOptions{
		JSON: map[string]string{
			"email":    email,
			"password": password,
		},
		CookieJar:    jar,
		UseCookieJar: true,
	}
	response, err := grequests.Post("https://api.leancloud.cn/1/signin", options)
	if err != nil {
		return nil, err
	}
	if !response.Ok {
		return nil, NewErrorFromBody(response.String())
	}

	if err := jar.Save(); err != nil {
		return nil, err
	}

	return simplejson.NewFromReader(response)
}

// LoginUSRegion will use OAuth2 to login US Region
func LoginUSRegion() error {
	client := NewClient()
	client.Region = RegionUS
	_, err := client.get("/1/oauth2/goto/avoscloud", nil)
	if err != nil {
		return err
	}
	return nil
}

// GetUserInfoResult is the return type of GetUserInfo
type GetUserInfoResult struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
}

// GetUserInfo returns the current logined user info
func GetUserInfo() (*GetUserInfoResult, error) {
	client := NewClient()

	resp, err := client.get("/1.1/clients/self", nil)
	if err != nil {
		return nil, err
	}

	result := new(GetUserInfoResult)
	err = resp.JSON(result)
	return result, err
}
