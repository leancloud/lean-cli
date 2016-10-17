package api

import (
	"os"
	"path/filepath"

	"github.com/juju/persistent-cookiejar"
	"github.com/leancloud/lean-cli/lean/api/regions"
	"github.com/leancloud/lean-cli/lean/utils"
	"github.com/leancloud/lean-cli/lean/version"
	"github.com/levigross/grequests"
)

// Login LeanCloud account
func Login(email string, password string) (*GetUserInfoResult, error) {
	os.MkdirAll(filepath.Join(utils.ConfigDir(), "leancloud"), 0775)
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
		UserAgent:    "LeanCloud-CLI/" + version.Version,
	}
	resp, err := grequests.Post("https://leancloud.cn/1/signin", options)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromResponse(resp)
	}

	if err := jar.Save(); err != nil {
		return nil, err
	}

	result := new(GetUserInfoResult)
	err = resp.JSON(result)
	return result, err
}

// LoginUSRegion will use OAuth2 to login US Region
func LoginUSRegion() error {
	client := NewClient(regions.US)
	_, err := client.get("/1/oauth2/goto/avoscloud", nil)
	if err != nil {
		return err
	}
	return nil
}

// LoginTABRegion will use qrcode to login TAB Region
func LoginTABRegion(token string) (*GetUserInfoResult, error) {
	client := NewClient(regions.TAB)
	resp, err := client.post("/1.1/signinByToken?token="+token, nil, nil)
	result := new(GetUserInfoResult)
	err = resp.JSON(result)
	return result, err
}

// GetUserInfoResult is the return type of GetUserInfo
type GetUserInfoResult struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
}

// GetUserInfo returns the current logined user info
func GetUserInfo() (*GetUserInfoResult, error) {
	client := NewClient(regions.CN)

	resp, err := client.get("/1.1/clients/self", nil)
	if err != nil {
		return nil, err
	}

	result := new(GetUserInfoResult)
	err = resp.JSON(result)
	return result, err
}
