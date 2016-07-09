package api

import (
	"github.com/levigross/grequests"
)

// Login LeanCloud account
func Login(email string, password string) error {
	options := &grequests.RequestOptions{
		JSON: map[string]string{
			"email":    email,
			"password": password,
		},
	}
	response, err := grequests.Post("https://leancloud.cn/1/signin", options)
	if err != nil {
		return err
	}
	if !response.Ok {
		return NewErrorFromBody(response.String())
	}

	cookies := response.RawResponse.Cookies()

	return saveCookies(cookies)
}

// UserInfo returns the current logined user info
func UserInfo() error {
	cookies, err := getCookies()
	if err != nil {
		return err
	}
	_ = cookies
	return nil
}
