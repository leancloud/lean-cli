package api

import (
	"encoding/base64"
	"io/ioutil"
	"time"
)

// UploadFileResult is the UploadFile's return type
type UploadFileResult struct {
	ObjectID string `json:"objectID"`
	URL      string `json:"url"`
}

// UploadFile upload specific file to LeanCloud
func UploadFile(appID string, filePath string) (*UploadFileResult, error) {
	appInfo, err := GetAppInfo(appID)
	if err != nil {
		return nil, err
	}

	fileName := "leanengine" + time.Now().Format("20060102150405") + ".zip"

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	params := map[string]interface{}{
		"base64":       base64.StdEncoding.EncodeToString(content),
		"_ContentType": "application/zip, application/octet-stream",
		"mime_type":    "application/zip",
	}

	client := NewClient()
	opts, err := client.options()
	if err != nil {
		return nil, err
	}
	opts.Headers["X-LC-Id"] = appInfo.AppID
	opts.Headers["X-LC-Key"] = appInfo.MasterKey + ",master"
	resp, err := client.postX("/1.1/files/"+fileName, params, opts)
	if err != nil {
		return nil, err
	}

	result := new(UploadFileResult)
	err = resp.JSON(result)
	return result, err
}

// DeleteFile will delete the specific file
func DeleteFile(appID string, objectID string) error {
	appInfo, err := GetAppInfo(appID)
	if err != nil {
		return err
	}
	client := NewClient()
	opts, err := client.options()
	if err != nil {
		return err
	}
	opts.Headers["X-LC-Id"] = appInfo.AppID
	opts.Headers["X-LC-Key"] = appInfo.MasterKey + ",master"
	_, err = client.delete("/1.1/files/"+objectID, opts)
	return err
}
