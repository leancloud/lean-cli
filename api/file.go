package api

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/leancloud/go-upload"
	"mime"
)

type fileBarReaderSeeker struct {
	file   *os.File
	reader *pb.Reader
}

func (f *fileBarReaderSeeker) Seek(offset int64, whence int) (ret int64, err error) {
	return f.file.Seek(offset, whence)
}

func (f *fileBarReaderSeeker) Read(b []byte) (n int, err error) {
	return f.reader.Read(b)
}

// UploadFile upload specific file to LeanCloud
func UploadFile(appID string, filePath string) (*upload.File, error) {
	appInfo, err := GetAppInfo(appID)
	if err != nil {
		return nil, err
	}

	region, err := GetAppRegion(appID)
	if err != nil {
		return nil, err
	}

	_, fileName := filepath.Split(filePath)
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	bar := pb.New(int(stat.Size())).SetUnits(pb.U_BYTES).SetMaxWidth(80)
	bar.Prefix(color.GreenString("[INFO]") + " 上传文件")
	bar.Start()

	// qiniu want a io.ReadSeeker to get file's size
	readSeeker := &fileBarReaderSeeker{
		file:   f,
		reader: bar.NewProxyReader(f),
	}

	file, err := upload.Upload(fileName, mimeType, readSeeker, &upload.Options{
		AppID:     appInfo.AppID,
		AppKey:    appInfo.MasterKey + ",master",
		ServerURL: NewClient(region).baseURL(),
	})
	if err != nil {
		return nil, err
	}

	bar.Finish()
	if err != nil {
		return nil, err
	}

	if file.URL == "" {
		return nil, errors.New("文件上传失败")
	}
	return file, err
}

// DeleteFile will delete the specific file
func DeleteFile(appID string, objectID string) error {
	region, err := GetAppRegion(appID)
	if err != nil {
		return nil
	}

	appInfo, err := GetAppInfo(appID)
	if err != nil {
		return err
	}
	client := NewClient(region)
	opts, err := client.options()
	if err != nil {
		return err
	}
	opts.Headers["X-LC-Id"] = appInfo.AppID
	opts.Headers["X-LC-Key"] = appInfo.MasterKey + ",master"
	_, err = client.delete("/1.1/files/"+objectID, opts)
	return err
}
