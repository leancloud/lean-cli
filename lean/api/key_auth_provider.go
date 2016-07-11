package api

import "github.com/levigross/grequests"

type keyAuthProvider struct {
	AppID     string
	MasterKey string
	Region    int
}

func (provider *keyAuthProvider) baseURL() string {
	switch provider.Region {
	case RegionCN:
		return hostCN + "/" + apiVersion
	case RegionUS:
		return hostUS + "/" + apiVersion
	default:
		panic("invalid region")
	}
}

func (provider *keyAuthProvider) options() *grequests.RequestOptions {
	return &grequests.RequestOptions{
		Headers: map[string]string{
			"X-AVOSCloud-Application-Id":         provider.AppID,
			"X-AVOSCloud-Master-Key":             provider.MasterKey,
			"X-AVOSCloud-Application-Production": "1",
			"Content-Type":                       "application/json",
		},
	}
}

// NewKeyAuthClient creates a app-id / app-key based api client
func NewKeyAuthClient(appID string, masterKey string) *Client {
	return &Client{
		provider: &keyAuthProvider{
			AppID:     appID,
			MasterKey: masterKey,
			Region:    RegionCN,
		},
	}
}
