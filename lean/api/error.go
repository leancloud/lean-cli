package api

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	// ErrNotLogined means user was not logined
	ErrNotLogined = errors.New("not logined")
)

// Error is the LeanCloud API Server API common error format
type Error struct {
	Code    int    `json:"code"`
	Content string `json:"error"`
}

func (err Error) Error() string {
	return fmt.Sprintf("LeanCloud API error %d: %s", err.Code, err.Content)
}

// NewErrorFromBody format LeanCloud Server
func NewErrorFromBody(body string) error {
	var err Error
	err2 := json.Unmarshal([]byte(body), &err)
	if err2 != nil {
		panic(err2)
	}
	return err
}
