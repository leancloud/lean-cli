package utils

import (
	"io/ioutil"
	"net/http"
	"path"
	"os/user"

	"github.com/leancloud/lean-cli/cookieparser"
)

// GetCookies returns the current user's lean cli cookie from
// $HOME/.leancloud/cookies
func GetCookies() ([]*http.Cookie, error) {
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
