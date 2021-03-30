package api

import (
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
)

// Login LeanCloud account
func Login(email string, password string, region regions.Region) (*GetUserInfoResult, error) {
	jar := newCookieJar()

	options := &grequests.RequestOptions{
		JSON: map[string]string{
			"email":    email,
			"password": password,
		},
		CookieJar:    jar,
		UseCookieJar: true,
		UserAgent:    "LeanCloud-CLI/" + version.Version,
	}
	client := NewClientByRegion(region)

	resp, err := grequests.Post(client.GetBaseURL()+"/1/signin", options)
	if err != nil {
		return nil, err
	}

	resp, err = client.checkAndDo2FA(resp)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, NewErrorFromResponse(resp)
	}

	if err := jar.Save(); err != nil {
		return nil, err
	}

	regions.SetRegionLoginStatus(region)
	if err := regions.SaveRegionLoginStatus(); err != nil {
		return nil, err
	}

	result := new(GetUserInfoResult)
	err = resp.JSON(result)
	return result, err
}

func LoginNext(accessKey string, region regions.Region) (*GetUserInfoResult, error) {
	client := NewClientByRegion(region)
	client.AccessKey = accessKey

	resp, err := client.get("/1.1/clients/self", nil)
	if err != nil {
		return nil, err
	}

	userInfo := new(GetUserInfoResult)
	if err := resp.JSON(userInfo); err != nil {
		return nil, err
	}

	if err := accessKeyCache.Add(accessKey, region).Save(); err != nil {
		return nil, err
	}

	return &GetUserInfoResult{}, nil
}

// GetUserInfoResult is the return type of GetUserInfo
type GetUserInfoResult struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
}

// GetUserInfo returns the current logined user info
func GetUserInfo(region regions.Region) (*GetUserInfoResult, error) {
	client := NewClientByRegion(region)

	resp, err := client.get("/1.1/clients/self", nil)
	if err != nil {
		return nil, err
	}

	result := new(GetUserInfoResult)
	err = resp.JSON(result)
	return result, err
}
