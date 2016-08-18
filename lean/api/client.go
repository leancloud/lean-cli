package api

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/juju/persistent-cookiejar"
	"github.com/leancloud/lean-cli/lean/api/regions"
	"github.com/leancloud/lean-cli/lean/utils"
	"github.com/leancloud/lean-cli/lean/version"
	"github.com/levigross/grequests"
)

const (
	hostCN = "https://leancloud.cn"
	hostUS = "https://us.leancloud.cn"
)

// Client info
type Client struct {
	CookieJar *cookiejar.Jar
	Region    int
}

// NewClient initilized a new Client
func NewClient(region int) *Client {
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

func (client *Client) fetchRouter() error {
	// TODO: fetch router from server
	return nil
}

func (client *Client) baseURL() string {
	switch client.Region {
	case regions.CN:
		return hostCN
	case regions.US:
		return hostUS
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

func (client *Client) get(path string, options *grequests.RequestOptions) (*grequests.Response, error) {
	var err error
	if options == nil {
		if options, err = client.options(); err != nil {
			return nil, err
		}
	}
	resp, err := grequests.Get(client.baseURL()+path, options)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
			return nil, NewErrorFromBody(resp.String())
		}
		return nil, fmt.Errorf("HTTP Error: %d", resp.StatusCode)
	}

	if err = client.CookieJar.Save(); err != nil {
		return resp, err
	}

	return resp, nil
}

func (client *Client) post(path string, params map[string]interface{}, options *grequests.RequestOptions) (*grequests.Response, error) {
	var err error
	if options == nil {
		if options, err = client.options(); err != nil {
			return nil, err
		}
	}
	options.JSON = params
	resp, err := grequests.Post(client.baseURL()+path, options)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromBody(resp.String())
	}

	if err = client.CookieJar.Save(); err != nil {
		return resp, err
	}

	return resp, nil
}

func (client *Client) put(path string, params map[string]interface{}, options *grequests.RequestOptions) (*grequests.Response, error) {
	var err error
	if options == nil {
		if options, err = client.options(); err != nil {
			return nil, err
		}
	}
	options.JSON = params
	resp, err := grequests.Put(client.baseURL()+path, options)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromBody(resp.String())
	}

	if err = client.CookieJar.Save(); err != nil {
		return resp, err
	}

	return resp, nil
}

func (client *Client) delete(path string, options *grequests.RequestOptions) (*grequests.Response, error) {
	var err error
	if options == nil {
		if options, err = client.options(); err != nil {
			return nil, err
		}
	}
	resp, err := grequests.Delete(client.baseURL()+path, options)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromBody(resp.String())
	}

	if err = client.CookieJar.Save(); err != nil {
		return resp, err
	}

	return resp, nil
}
