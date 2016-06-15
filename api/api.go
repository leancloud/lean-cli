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
	RegionCN = iota
	RegionUS
)

const apiVersion = "1.1"

// Client info
type Client struct {
	AppID     string
	MasterKey string
	Region    int
}

func (client *Client) baseURL() string {
	switch client.Region {
	case RegionCN:
		return hostCN + "/" + apiVersion
	case RegionUS:
		return hostUS + "/" + apiVersion
	default:
		panic("invalid region")
	}
}

func (client *Client) options() *grequests.RequestOptions {
	return &grequests.RequestOptions{
		Headers: map[string]string{
			"X-AVOSCloud-Application-Id":         client.AppID,
			"X-AVOSCloud-Master-Key":             client.MasterKey,
			"X-AVOSCloud-Application-Production": "1",
			"Content-Type":                       "application/json",
		},
	}
}

func (client *Client) get(path string) (*simplejson.Json, error) {
	resp, err := grequests.Get(client.baseURL()+path, client.options())
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromBody(resp.String())
	}
	return simplejson.NewFromReader(resp)
}

// AppDetail returns the app's detail infomation
func (client *Client) AppDetail() (*simplejson.Json, error) {
	return client.get("/__leancloud/apps/appDetail")
}
