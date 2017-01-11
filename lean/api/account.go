package api

import (
	"net/url"
	"os"
	"path/filepath"
	"time"

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
	token = url.QueryEscape(token)
	resp, err := client.post("/1.1/signinByToken?token="+token, nil, nil)
	if err != nil {
		return nil, err
	}
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
func GetUserInfo(region regions.Region) (*GetUserInfoResult, error) {
	client := NewClient(region)

	resp, err := client.get("/1.1/clients/self", nil)
	if err != nil {
		return nil, err
	}

	result := new(GetUserInfoResult)
	err = resp.JSON(result)
	return result, err
}

// GetLoginedRegion returns all regions which is logined
func GetLoginedRegion() (result []regions.Region, err error) {
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: filepath.Join(utils.ConfigDir(), "leancloud", "cookies"),
	})
	if err != nil {
		return nil, err
	}

	cookies := jar.AllCookies()

	for _, cookie := range cookies {
		if cookie.Name != "uluru_user" {
			continue
		}
		if cookie.Expires.Before(time.Now()) {
			continue
		}
		switch cookie.Domain {
		case "leancloud.cn":
			result = append(result, regions.CN)
		case "us.leancloud.cn":
			result = append(result, regions.US)
		case "tab.leancloud.cn":
			result = append(result, regions.TAB)
		}
	}

	return
}
