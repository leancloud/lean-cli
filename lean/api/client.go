package api

import (
	"github.com/bitly/go-simplejson"
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

const apiVersion = "1.1"

type authProvider interface {
	options() *grequests.RequestOptions
	baseURL() string
}

// Client info
type Client struct {
	provider authProvider
	// AppID     string
	// MasterKey string
	// Region    int
}

func fetchRouter() error {
	// TODO: fetch router from server
	return nil
}

func (client *Client) get(path string, options *grequests.RequestOptions) (*simplejson.Json, error) {
	if options == nil {
		options = client.provider.options()
	}
	resp, err := grequests.Get(client.provider.baseURL()+path, options)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromBody(resp.String())
	}

	return simplejson.NewFromReader(resp)
}

func (client *Client) getJSON(path string, options *grequests.RequestOptions) (interface{}, error) {
	if options == nil {
		options = client.provider.options()
	}
	resp, err := grequests.Get(client.provider.baseURL()+path, options)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromBody(resp.String())
	}

	var result interface{}
	err = resp.JSON(&result)
	return result, err
}

func (client *Client) post(path string, params map[string]interface{}, options *grequests.RequestOptions) (*simplejson.Json, error) {
	if options == nil {
		options = client.provider.options()
	}
	options.JSON = params
	resp, err := grequests.Post(client.provider.baseURL()+path, options)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromBody(resp.String())
	}
	return simplejson.NewFromReader(resp)
}

func (client *Client) delete(path string, options *grequests.RequestOptions) (*simplejson.Json, error) {
	if options == nil {
		options = client.provider.options()
	}
	resp, err := grequests.Delete(client.provider.baseURL()+path, options)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromBody(resp.String())
	}
	return simplejson.NewFromReader(resp)
}
