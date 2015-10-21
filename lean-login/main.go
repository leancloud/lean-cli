package main

import (
	"encoding/json"
	"fmt"
	"github.com/howeyc/gopass"
	"github.com/parnurzeal/gorequest"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

type errorResult struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type authResult struct {
	UserName     string `json:"username"`
	SessionToken string `json:"session_token"`
}

func printServerErrorResult(body string) {
	var result errorResult
	json.Unmarshal([]byte(body), &result)
	fmt.Println(result.Error)
}

func getLoginInfoFromInput() (string, string) {
	var email string
	var password string

	fmt.Print("Your LeanCloud login email: ")
	fmt.Scanln(&email)

	fmt.Print("Your LeanCloud login password (will hidden while input): ")
	password = string(gopass.GetPasswd())

	// fmt.Printf("email: %s, password: %s\r\n", email, password)

	return email, password
}

func saveAuthResult(authRst authResult) error {
	raw, err := json.MarshalIndent(authRst, "", "    ")
	if err != nil {
		return err
	}

	// fmt.Println(string(raw))

	user, err := user.Current()
	if err != nil {
		return err
	}

	if err = os.Mkdir(path.Join(user.HomeDir, ".lean"), 0700); err != nil {
		if os.IsNotExist(err) {
			return err
		}
	}

	fileName := path.Join(user.HomeDir, ".lean", "user.json")

	err = ioutil.WriteFile(fileName, raw, 0644)

	return err
}

func getAuthResult(email string, password string) authResult {
	request := gorequest.New()
	resp, body, err := request.Post("https://leancloud.cn/1/signin").
		Set("User-Agent", "leanengine-cli x.x.x"). // TODO
		Send(fmt.Sprintf(`{"email": "%s", "password": "%s"}`, email, password)).
		End()

	if err != nil {
		fmt.Fprint(os.Stderr, "Error in login: ")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if resp.StatusCode != 200 {
		fmt.Fprint(os.Stderr, "Login failed: ")
		printServerErrorResult(body)
		os.Exit(1)
	}

	var result authResult
	json.Unmarshal([]byte(body), &result)
	// fmt.Println(result)

	return result
}

func main() {
	email, password := getLoginInfoFromInput()
	authRst := getAuthResult(email, password)
	err := saveAuthResult(authRst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\r\n", err)
	}
}
