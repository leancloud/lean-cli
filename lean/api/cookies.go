package api

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aisk/cookieparser"
	"github.com/leancloud/lean-cli/lean/api/regions"
	"github.com/leancloud/lean-cli/lean/utils"
)

func cookiesFilePath(region int) string {
	switch region {
	case regions.CN:
		return filepath.Join(utils.ConfigDir(), "leancloud", "cn_region_cookies")
	case regions.US:
		return filepath.Join(utils.ConfigDir(), "leancloud", "us_region_cookies")
	default:
		panic("invalid region")
	}
}

// saveCookies saves the cookies to `${HOME}/.leancloud/cookies`
func saveCookies(cookies []*http.Cookie, region int) error {
	os.MkdirAll(filepath.Join(utils.ConfigDir(), "leancloud"), 0775)

	content := []byte(cookieparser.ToString(cookies))
	return ioutil.WriteFile(cookiesFilePath(region), content, 0664)
}

func getCookies(region int) ([]*http.Cookie, error) {
	content, err := ioutil.ReadFile(cookiesFilePath(region))
	if err != nil {
		return nil, err
	}
	cookies := cookieparser.Parse(string(content))
	return cookies, nil
}
