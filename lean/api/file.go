package api

import (
	"encoding/base64"
	"io/ioutil"
	"time"
)

// File ...
type File struct {
	ID  string
	URL string
}

// UploadFile ...
func (client *Client) UploadFile(filePath string) (*File, error) {
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

	jsonObj, err := client.post("/files/"+fileName, params, nil)
	if err != nil {
		return nil, err
	}

	return &File{
		ID:  jsonObj.Get("objectId").MustString(),
		URL: jsonObj.Get("url").MustString(),
	}, nil
}

// DeleteFile ...
func (client *Client) DeleteFile(ID string) error {
	_, err := client.delete("/files/"+ID, nil)
	return err
}
