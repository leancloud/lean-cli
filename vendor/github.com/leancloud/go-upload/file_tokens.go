package upload

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"
)

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterBytes   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var client = &http.Client{
	Timeout: 14*time.Second + 1*time.Second,
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randStringBytesMask(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

func getFileKey(name string) string {
	return randStringBytesMask(40) + filepath.Ext(name)
}

type fileTokens struct {
	ObjectID  string `json:"objectId"`
	URL       string `json:"url"`
	Provider  string `json:"provider"`
	UploadURL string `json:"upload_url"`
	Token     string `json:"token"`
	Key       string
}

func getFileTokens(name string, mime string, size int64, opts *Options) (*fileTokens, error) {
	key := getFileKey(name)
	data := map[string]interface{}{
		"name":      name,
		"key":       key,
		"mime_type": mime,
		"metaData": map[string]interface{}{
			"owner": "unknown",
			"size":  size,
		},
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	url := opts.serverURL() + "/1.1/fileTokens"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-LC-Id", opts.AppID)
	request.Header.Set("X-LC-Key", opts.AppKey)
	request.Header.Set("User-Agent", "LeanCloud-Go-Upload/"+version)
	request.Header.Set("Content-Type", "Application/JSON")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err = ioutil.ReadAll(response.Body)

	if response.StatusCode != 201 {
		return nil, errors.New(string(body))
	}

	result := new(fileTokens)
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	result.Key = key // key is not in Server response
	return result, err
}
