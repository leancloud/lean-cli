package upload

import (
	"encoding/json"
	"errors"
	"fmt"
)

const version = "0.1.0"

// File is the File's return type
type File struct {
	ObjectID string `json:"objectID"`
	URL      string `json:"url"`
}

// Error is the LeanCloud API Server API common error format
type Error struct {
	Code         int    `json:"code"`
	Content      string `json:"error"`
	ErrorEventID string `json:"errorEventID"`
}

func newErrorFromBody(body []byte) error {
	var err Error
	err2 := json.Unmarshal([]byte(body), &err)
	if err2 != nil {
		return errors.New("Upload failed")
	}
	return err
}

func (err Error) Error() string {
	return fmt.Sprintf("LeanCloud API error %d: %s", err.Code, err.Content)
}

// Options is upload options type
type Options struct {
	AppID     string
	AppKey    string
	ServerURL string
}

func (opts *Options) serverURL() string {
	if opts.ServerURL != "" {
		return opts.ServerURL
	}
	return "https://api.leancloud.cn"
}
