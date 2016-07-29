package api

import (
	"github.com/bitly/go-simplejson"
	"github.com/levigross/grequests"
)

// Login LeanCloud account
func Login(email string, password string) (*simplejson.Json, error) {
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
		return nil, NewErrorFromBody(response.String())
	}

	cookies := response.RawResponse.Cookies()

	if err := saveCookies(cookies); err != nil {
		return nil, err
	}

	return simplejson.NewFromReader(response)
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
