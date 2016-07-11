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

// UserInfo returns the current logined user info
func UserInfo() (*simplejson.Json, error) {
	client, err := NewCookieAuthClient()
	if err != nil {
		return nil, err
	}
	return client.get("/clients/self", nil)
}
