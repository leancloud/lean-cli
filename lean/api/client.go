package api

import (
	"fmt"
	"strings"

	"github.com/bitly/go-simplejson"
	"github.com/leancloud/lean-cli/lean/version"
	"github.com/levigross/grequests"
)

const (
	hostCN = "https://api.leancloud.cn"
	hostUS = "https://us-api.leancloud.cn"
)

// API server regions
const (
	RegionInvalid = iota
	RegionCN
	RegionUS
)

// Client info
type Client struct {
}

// NewClient initilized a new Client
func NewClient() *Client {
	return &Client{}
}

func (client *Client) fetchRouter() error {
	// TODO: fetch router from server
	return nil
}

func (client *Client) baseURL() string {
	// TODO: return base URL per region
	return hostCN
}

func (client *Client) options() (*grequests.RequestOptions, error) {
	cookies, err := getCookies()
	if err != nil {
		return nil, ErrNotLogined
	}

	xsrfTok := ""
	for _, cookie := range cookies {
		if cookie.Name == "XSRF-TOKEN" {
			xsrfTok = cookie.Value
			break
		}
	}

	return &grequests.RequestOptions{
		Cookies: cookies,
		Headers: map[string]string{
			"X-XSRF-TOKEN": xsrfTok,
		},
		UserAgent: "LeanCloud-CLI/" + version.Version,
	}, nil
}

func (client *Client) get(path string, options *grequests.RequestOptions) (*simplejson.Json, error) {
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
		return nil, NewErrorFromBody(resp.String())
	}

	return simplejson.NewFromReader(resp)
}

func (client *Client) getX(path string, options *grequests.RequestOptions) (*grequests.Response, error) {
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

	return resp, nil
}

func (client *Client) post(path string, params map[string]interface{}, options *grequests.RequestOptions) (*simplejson.Json, error) {
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
	return simplejson.NewFromReader(resp)
}

func (client *Client) postX(path string, params map[string]interface{}, options *grequests.RequestOptions) (*grequests.Response, error) {
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
	return resp, nil
}

func (client *Client) putX(path string, params map[string]interface{}, options *grequests.RequestOptions) (*grequests.Response, error) {
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
	return resp, nil
}

func (client *Client) delete(path string, options *grequests.RequestOptions) (*simplejson.Json, error) {
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
	return simplejson.NewFromReader(resp)
}
