package api

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/levigross/grequests"
)

var (
	// ErrNotLogined means user was not logined
	ErrNotLogined = errors.New("not logined")
)

// Error is the LeanCloud API Server API common error format
type Error struct {
	Code         int    `json:"code"`
	Content      string `json:"error"`
	ErrorEventID string `json:"errorEventID"`
}

func (err Error) Error() string {
	return fmt.Sprintf("LeanCloud API error %d: %s", err.Code, err.Content)
}

// NewErrorFromBody build an error value from JSON string
func NewErrorFromBody(body string) error {
	var err Error
	err2 := json.Unmarshal([]byte(body), &err)
	if err2 != nil {
		panic(err2)
	}
	return err
}

// NewErrorFromResponse build an error value from *grequest.Response
func NewErrorFromResponse(resp *grequests.Response) error {
	contentType := resp.Header.Get("Content-Type")
	if contentType == "application/json" || contentType == "application/json;charset=utf-8" {
		return NewErrorFromBody(resp.String())
	}

	return &Error{
		Code:    resp.StatusCode,
		Content: fmt.Sprintf("Got invalid HTML response, status code: %d", resp.StatusCode),
	}
}
