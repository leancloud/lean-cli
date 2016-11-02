package upload

import (
	"errors"
	qiniu "github.com/qiniu/api.v6/io"
	"io"
	"mime"
	"os"
	"path/filepath"

	"github.com/levigross/grequests"
)

// Upload upload specific file to LeanCloud
func Upload(name string, mimeType string, reader io.Reader, opts *Options) (*File, error) {
	if opts.serverURL() == "https://api.leancloud.cn" {
		tokens, err := getFileTokens(name, mimeType, opts)
		if err != nil {
			return nil, err
		}
		putRet := new(qiniu.PutRet)
		err = qiniu.Put(nil, putRet, tokens.Token, tokens.Key, reader, &qiniu.PutExtra{
			MimeType: mimeType,
		})
		if err != nil {
			return nil, err
		}
		file := &File{
			ObjectID: tokens.ObjectID,
			URL:      tokens.URL,
		}
		return file, nil
	}

	reqOpts := &grequests.RequestOptions{
		Headers: map[string]string{
			"X-LC-Id":  opts.AppID,
			"X-LC-Key": opts.AppKey,
		},
		UserAgent:   "LeanCloud-Go-Upload/" + version,
		RequestBody: reader,
	}

	resp, err := grequests.Post(opts.serverURL()+"/1.1/files/"+name, reqOpts)
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, newErrorFromBody(resp.Bytes())
	}

	result := new(File)
	err = resp.JSON(result)
	if result.URL == "" {
		return nil, errors.New("Upload file failed")
	}
	return result, err
}

// UploadFileVerbose will open an file and upload it
func UploadFileVerbose(name string, mimeType string, path string, opts *Options) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return Upload(name, mimeType, f, opts)
}

// UploadFile will open an file and upload it. the file name and mime type is autodetected
func UploadFile(path string, opts *Options) (*File, error) {
	_, name := filepath.Split(path)
	mimeType := mime.TypeByExtension(filepath.Ext(path))
	return UploadFileVerbose(name, mimeType, path, opts)
}
