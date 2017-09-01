package api

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aisk/wizard"
	"github.com/juju/persistent-cookiejar"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/utils"
	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
)

const (
	hostCN  = "https://leancloud.cn"
	hostUS  = "https://us.leancloud.cn"
	hostTAB = "https://tab.leancloud.cn"
)

var (
	// Get2FACode is the function to get the user's two-factor-authentication code.
	// You can override it with your custom function.
	Get2FACode = func() (int, error) {
		result := new(string)
		wizard.Ask([]wizard.Question{
			{
				Content: "请输入二次认证验证码",
				Input: &wizard.Input{
					Result: result,
					Hidden: false,
				},
			},
		})
		code, err := strconv.Atoi(*result)
		if err != nil {
			return 0, errors.New("二次认证验证码应该为数字")
		}
		return code, nil
	}
)

// Client info
type Client struct {
	CookieJar *cookiejar.Jar
	Region    regions.Region
}

// NewClient initilized a new Client
func NewClient(region regions.Region) *Client {
	os.MkdirAll(filepath.Join(utils.ConfigDir(), "leancloud"), 0775)
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: filepath.Join(utils.ConfigDir(), "leancloud", "cookies"),
	})
	if err != nil {
		panic(err)
	}
	return &Client{
		CookieJar: jar,
		Region:    region,
	}
}

func (client *Client) baseURL() string {
	switch client.Region {
	case regions.CN:
		return hostCN
	case regions.US:
		return hostUS
	case regions.TAB:
		return hostTAB
	default:
		panic("invalid region")
	}
}

func (client *Client) options() (*grequests.RequestOptions, error) {
	u, err := url.Parse(client.baseURL())
	if err != nil {
		panic(err)
	}
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
			"X-XSRF-TOKEN": xsrf,
		},
		CookieJar:    client.CookieJar,
		UseCookieJar: true,
		UserAgent:    "LeanCloud-CLI/" + version.Version,
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
	resp, err := fn(client.baseURL()+path, options)
	if err != nil {
		return nil, err
	}

	resp, err = client.checkAndDo2FA(resp)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		if strings.HasPrefix(strings.TrimSpace(resp.Header.Get("Content-Type")), "application/json") {
			return nil, NewErrorFromResponse(resp)
		}
		return nil, fmt.Errorf("HTTP Error: %d, %s %s", resp.StatusCode, method, path)
	}

	if err = client.CookieJar.Save(); err != nil {
		return nil, err
	}

	return resp, nil
}

// check if the requests need two-factor-authentication and then do it.
func (client *Client) checkAndDo2FA(resp *grequests.Response) (*grequests.Response, error) {
	if resp.StatusCode != 401 {
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

	resp, err = grequests.Post(client.baseURL()+"/1.1/do2fa", &grequests.RequestOptions{
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
