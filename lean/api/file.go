package api

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb"
	"github.com/levigross/grequests"
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

	_, fileName := filepath.Split(filePath)

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	bar := pb.New(int(stat.Size())).SetUnits(pb.U_BYTES).SetMaxWidth(80)
	bar.Prefix("> 上传项目文件")
	bar.Start()
	reader := bar.NewProxyReader(f)

	opts := &grequests.RequestOptions{
		Headers: map[string]string{
			"X-LC-Id":      appInfo.AppID,
			"X-LC-Key":     appInfo.MasterKey + ",master",
			"Content-Type": "application/zip, application/octet-stream",
		},
		RequestBody: reader,
	}
	resp, err := grequests.Post(NewClient().baseURL()+"/1.1/files/"+fileName, opts)
	bar.Finish()
	if err != nil {
		return nil, err
	}
	if !resp.Ok {
		return nil, NewErrorFromBody(resp.String())
	}

	result := new(UploadFileResult)
	err = resp.JSON(result)
	if result.URL == "" {
		return nil, errors.New("文件上传失败")
	}
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
