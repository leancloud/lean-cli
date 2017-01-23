package api

import (
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
	"net/url"
)

// ExecuteCQLResult is ExecuteCQL's result type
type ExecuteCQLResult struct {
	ClassName string                   `json:"className"`
	Results   []map[string]interface{} `json:"results"`
	Count     int                      `json:"count"`
}

// ExecuteCQL will execute the cql, and returns' the result
func ExecuteCQL(appID string, masterKey string, region regions.Region, cql string) (*ExecuteCQLResult, error) {
	opts := &grequests.RequestOptions{
		Headers: map[string]string{
			"X-LC-Id":      appID,
			"X-LC-Key":     masterKey + ",master",
			"Content-Type": "application/zip, application/octet-stream",
		},
		UserAgent: "LeanCloud-CLI/" + version.Version,
	}
	resp, err := grequests.Get(NewClient(region).baseURL()+"/1.1/cloudQuery?cql="+url.QueryEscape(cql), opts)
	if err != nil {
		return nil, err
	}

	if !resp.Ok {
		return nil, NewErrorFromResponse(resp)
	}

	result := new(ExecuteCQLResult)
	result.Count = -1 // means no count returned
	err = resp.JSON(result)
	return result, err
}
