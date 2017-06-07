package utils

import "encoding/json"

// ErrorResult is the LeanCloud API Server API common error format
type ErrorResult struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

// FormatServerErrorResult format LeanCloud Server
func FormatServerErrorResult(body string) string {
	var result ErrorResult
	json.Unmarshal([]byte(body), &result)
	return result.Error
}
