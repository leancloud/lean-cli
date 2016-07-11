package api

import (
	"errors"
	"net/http"

	"github.com/levigross/grequests"
)

var (
	// ErrNotLogined means no user logined
	ErrNotLogined = errors.New("Not Logined")
)

type cookieAuthProvider struct {
	cookies []*http.Cookie
}

func (provider *cookieAuthProvider) baseURL() string {
	// TODO: multi region support
	return "https://leancloud.cn/1"
}

func (provider *cookieAuthProvider) options() *grequests.RequestOptions {
	return &grequests.RequestOptions{
		Cookies: provider.cookies,
	}
}

// NewCookieAuthClient returns a cookie based api client
func NewCookieAuthClient() (*Client, error) {
	cookies, err := getCookies()
	if err != nil {
		return nil, ErrNotLogined
	}
	return &Client{
		provider: &cookieAuthProvider{
			cookies: cookies,
		},
	}, nil
}
