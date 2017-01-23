package upload

import (
	"math/rand"
	"path/filepath"
	"time"

	"github.com/levigross/grequests"
)

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

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

func getFileTokens(name string, mime string, opts *Options) (*fileTokens, error) {
	key := getFileKey(name)
	reqOpts := &grequests.RequestOptions{
		Headers: map[string]string{
			"X-LC-Id":  opts.AppID,
			"X-LC-Key": opts.AppKey,
		},
		JSON: map[string]interface{}{
			"name":      name,
			"key":       key,
			"mime_type": mime,
			"metaData":  map[string]interface{}{},
		},
	}
	resp, err := grequests.Post(opts.serverURL()+"/1.1/fileTokens", reqOpts)
	if err != nil {
		return nil, err
	}
	result := new(fileTokens)
	err = resp.JSON(result)
	if err == nil {
		result.Key = key
	}
	return result, err
}
