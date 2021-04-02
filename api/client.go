package api

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aisk/logp"
	"github.com/aisk/wizard"
	"github.com/cloudfoundry-attic/jibber_jabber"
	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/utils"
	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
)

var dashboardBaseUrls = map[regions.Region]string{
	regions.ChinaNorth: "https://cn-n1-console-api.leancloud.cn",
	regions.USWest:     "https://us-w1-console-api.leancloud.app",
	regions.ChinaEast:  "https://cn-e1-console-api.leancloud.cn",
}

var (
	// Get2FACode is the function to get the user's two-factor-authentication code.
	// You can override it with your custom function.
	Get2FACode = func() (int, error) {
		result := new(string)
		wizard.Ask([]wizard.Question{
			{
				Content: "Please input 2-factor auth code",
				Input: &wizard.Input{
					Result: result,
					Hidden: false,
				},
			},
		})
		code, err := strconv.Atoi(*result)
		if err != nil {
			return 0, errors.New("2-factor auth code should be numerical")
		}
		return code, nil
	}
)

type Client struct {
	CookieJar   *cookiejar.Jar
	Region      regions.Region
	AppID       string
	AccessToken string
}

func NewClientByRegion(region regions.Region) *Client {
	if version.Distribution == "lean" {
		return &Client{
			CookieJar: newCookieJar(),
			Region:    region,
		}
	}

	return &Client{
		AccessToken: accessTokenCache[region],
		Region:      region,
	}
}

func NewClientByApp(appID string) *Client {
	if version.Distribution == "lean" {
		return &Client{
			CookieJar: newCookieJar(),
			AppID:     appID,
		}
	}

	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return &Client{
			AppID: appID,
		}
	}

	return &Client{
		AccessToken: accessTokenCache[region],
		AppID:       appID,
	}
}

func (client *Client) GetBaseURL() string {
	envBaseURL := os.Getenv("LEANCLOUD_DASHBOARD")

	if envBaseURL != "" {
		return envBaseURL
	}

	region := client.Region

	if client.AppID != "" {
		var err error
		region, err = apps.GetAppRegion(client.AppID)

		if err != nil {
			panic(err) // This error should be catch at top level
		}
	}

	if url, ok := dashboardBaseUrls[region]; ok {
		return url
	}
	panic("invalid region")
}

func (client *Client) options() (*grequests.RequestOptions, error) {
	u, err := url.Parse(client.GetBaseURL())
	if err != nil {
		panic(err)
	}

	if version.Distribution == "lean" {
		cookies := client.CookieJar.Cookies(u)
		xsrf := ""
		for _, cookie := range cookies {
			if cookie.Name == "XSRF-TOKEN" {
				xsrf = cookie.Value
				break
			}
		}

		return &grequests.RequestOptions{
			Headers: map[string]string{
				"X-XSRF-TOKEN":    xsrf,
				"Accept-Language": getSystemLanguage(),
			},
			CookieJar:    client.CookieJar,
			UseCookieJar: true,
			UserAgent:    "LeanCloud-CLI/" + version.Version,
		}, nil
	}

	return &grequests.RequestOptions{
		Headers: map[string]string{
			"Accept-Language": getSystemLanguage(),
			"Authorization":   fmt.Sprint("Bearer ", client.AccessToken),
		},
		UserAgent: "TDS-CLI/" + version.Version,
	}, nil
}

func doRequest(client *Client, method string, path string, params map[string]interface{}, options *grequests.RequestOptions) (*grequests.Response, error) {
	var err error
	if options == nil {
		if options, err = client.options(); err != nil {
			return nil, err
		}
	}
	if params != nil {
		options.JSON = params
	}
	var fn func(string, *grequests.RequestOptions) (*grequests.Response, error)
	switch method {
	case "GET":
		fn = grequests.Get
	case "POST":
		fn = grequests.Post
	case "PUT":
		fn = grequests.Put
	case "DELETE":
		fn = grequests.Delete
	case "PATCH":
		fn = grequests.Patch
	default:
		panic("invalid method: " + method)
	}
	resp, err := fn(client.GetBaseURL()+path, options)
	if err != nil {
		return nil, err
	}

	if version.Distribution == "lean" {
		resp, err = client.checkAndDo2FA(resp)
		if err != nil {
			return nil, err
		}
	}

	if !resp.Ok {
		if strings.HasPrefix(strings.TrimSpace(resp.Header.Get("Content-Type")), "application/json") {
			return nil, NewErrorFromResponse(resp)
		}
		return nil, fmt.Errorf("HTTP Error: %d, %s %s", resp.StatusCode, method, path)
	}

	if version.Distribution == "lean" {
		if err = client.CookieJar.Save(); err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// check if the requests need two-factor-authentication and then do it.
func (client *Client) checkAndDo2FA(resp *grequests.Response) (*grequests.Response, error) {
	if resp.StatusCode != 401 || strings.Contains(resp.String(), "User doesn't sign in.") {
		// don't need 2FA
		return resp, nil
	}
	var result struct {
		Token string `json:"token"`
	}
	err := resp.JSON(&result)
	if err != nil {
		return nil, err
	}
	token := result.Token
	if token == "" {
		return resp, nil
	}
	code, err := Get2FACode()
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: filepath.Join(utils.ConfigDir(), "leancloud", "cookies"),
	})
	if err != nil {
		return nil, err
	}

	resp, err = grequests.Post(client.GetBaseURL()+"/1.1/do2fa", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"token": token,
			"code":  code,
		},
		CookieJar: jar,
	})
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		if strings.HasPrefix(strings.TrimSpace(resp.Header.Get("Content-Type")), "application/json") {
			return nil, NewErrorFromResponse(resp)
		}
		return nil, fmt.Errorf("HTTP Error: %d, %s %s", resp.StatusCode, "POST", "/do2fa")
	}

	if err := jar.Save(); err != nil {
		return nil, err
	}

	return resp, nil
}

func (client *Client) get(path string, options *grequests.RequestOptions) (*grequests.Response, error) {
	return doRequest(client, "GET", path, nil, options)
}

func (client *Client) post(path string, params map[string]interface{}, options *grequests.RequestOptions) (*grequests.Response, error) {
	return doRequest(client, "POST", path, params, options)
}

func (client *Client) patch(path string, params map[string]interface{}, options *grequests.RequestOptions) (*grequests.Response, error) {
	return doRequest(client, "PATCH", path, params, options)
}

func (client *Client) put(path string, params map[string]interface{}, options *grequests.RequestOptions) (*grequests.Response, error) {
	return doRequest(client, "PUT", path, params, options)
}

func (client *Client) delete(path string, options *grequests.RequestOptions) (*grequests.Response, error) {
	return doRequest(client, "DELETE", path, nil, options)
}

func newCookieJar() *cookiejar.Jar {
	jarFileDir := filepath.Join(utils.ConfigDir(), "leancloud")

	os.MkdirAll(jarFileDir, 0775)

	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: filepath.Join(jarFileDir, "cookies"),
	})
	if err != nil {
		panic(err)
	}
	return jar
}

func getSystemLanguage() string {
	language, err := jibber_jabber.DetectLanguage()

	if err != nil {
		logp.Info("unsupported locale setting & set to default en_US.UTF-8: ", err)
		language = "en"
	}

	return language
}
