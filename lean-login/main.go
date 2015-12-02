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
	"strings"

	"github.com/howeyc/gopass"
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

func getLoginInfoFromInput() (string, string) {
	var email string
	var password string

	fmt.Print("Your LeanCloud login email: ")
	fmt.Scanln(&email)

	fmt.Print("Your LeanCloud login password (will hidden while input): ")
	password = string(gopass.GetPasswd())

	return email, password
}

func login(email string, password string) ([]*http.Cookie, error) {
	request := gorequest.New()
	resp, body, errs := request.Post("https://leancloud.cn/1/signin").
		Set("User-Agent", "leanengine-cli x.x.x"). // TODO
		Send(fmt.Sprintf(`{"email": "%s", "password": "%s"}`, email, password)).
		End()

	if len(errs) != 0 {
		return nil, errs[0]
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(formatServerErrorResult(body))
	}

	cookies := request.Client.Jar.Cookies(resp.Request.URL)
	return cookies, nil
}

func cookiesToString(cookies []*http.Cookie) string {
	var cookieStrings []string
	for _, cookie := range cookies {
		cookieStrings = append(cookieStrings, cookie.String())
	}
	return strings.Join(cookieStrings, ",")
}

func saveCookies(cookies []*http.Cookie) error {
	cookieStrings := cookiesToString(cookies)

	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	fileName := path.Join(currentUser.HomeDir, ".leancloud", "cookies")

	return ioutil.WriteFile(fileName, []byte(cookieStrings), 0600)
}

func main() {
	email, password := getLoginInfoFromInput()
	cookies, err := login(email, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\r\n", err)
		os.Exit(1)
	}
	err = saveCookies(cookies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\r\n", err)
		os.Exit(1)
	}
	fmt.Println("login ok")
}
