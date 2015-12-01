package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"

	"github.com/leancloud/lean-cli/cookieparser"
	"github.com/parnurzeal/gorequest"
)

type errorResult struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func formatServerErrorResult(body string) string {
	var result errorResult
	json.Unmarshal([]byte(body), &result)
	return result.Error
}

type UserInfo struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

func getCookies() ([]*http.Cookie, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}

	fileName := path.Join(currentUser.HomeDir, ".leancloud", "cookies")
	raw, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	cookies := cookieparser.Parse(string(raw))
	return cookies, nil
}

func getMyInfo(cookies []*http.Cookie) (*UserInfo, []error) {
	var info UserInfo
	request := gorequest.New()
	request.SetDebug(false)
	resp, body, errs := request.Get("https://leancloud.cn/1/clients/self").
		AddCookies(cookies).
		Set("User-Agent", "leanengine-cli x.x.x"). // TODO
		End()

	if len(errs) != 0 {
		return nil, errs
	}

	if resp.StatusCode != 200 {
		return nil, []error{errors.New(formatServerErrorResult(body))}
	}

	if err := json.Unmarshal([]byte(body), &info); err != nil {
		return nil, []error{err}
	}

	return &info, nil
}

func main() {
	cookies, err := getCookies()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Not logined, please use 'lean login' to login.")
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\r\n", err)
		}
		os.Exit(1)
	}
	myinfo, errs := getMyInfo(cookies)
	if len(errs) != 0 {
		fmt.Fprintf(os.Stderr, "Error: %s\r\nMaybe session is expired, pelase use 'lean login' to relogin\r\n", errs[0])
		os.Exit(1)
	}

	fmt.Printf("username: %s\r\nemail: %s\r\n", myinfo.Username, myinfo.Email)
}
