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
		UserAgent:    version.GetUserAgent(),
	}
	client := NewClientByRegion(region)

	resp, err := grequests.Post(client.GetBaseURL()+"/client-center/2/signin", options)
	if resp.StatusCode == 401 {
		var result struct {
			Token string `json:"token"`
		}

		if err = resp.JSON(&result); err != nil {
			return nil, err
		}
		token := result.Token
		if token != "" {
			code, err := Get2FACode()
			if err != nil {
				return nil, err
			}
			options.JSON = map[string]string{
				"email":    email,
				"password": password,
				"code":     code,
			}
			resp, err = grequests.Post(client.GetBaseURL()+"/client-center/2/signin", options)
		}
	}

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

func LoginWithAccessToken(accessToken string, region regions.Region) (*GetUserInfoResult, error) {
	client := NewClientByRegion(region)
	client.AccessToken = accessToken

	resp, err := client.get("/client-center/2/clients/self", nil)
	if err != nil {
		return nil, err
	}

	userInfo := new(GetUserInfoResult)
	if err := resp.JSON(userInfo); err != nil {
		return nil, err
	}

	if err := accessTokenCache.Add(accessToken, region).Save(); err != nil {
		return nil, err
	}
	regions.SetRegionLoginStatus(region)
	if err := regions.SaveRegionLoginStatus(); err != nil {
		return nil, err
	}

	return userInfo, nil
}

// GetUserInfoResult is the return type of GetUserInfo
type GetUserInfoResult struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
}

// GetUserInfo returns the current logined user info
func GetUserInfo(region regions.Region) (*GetUserInfoResult, error) {
	client := NewClientByRegion(region)

	resp, err := client.get("/client-center/2/clients/self", nil)
	if err != nil {
		return nil, err
	}

	result := new(GetUserInfoResult)
	err = resp.JSON(result)
	return result, err
}
