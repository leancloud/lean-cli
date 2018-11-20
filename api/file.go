package api

import (
	"errors"
	"mime"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/leancloud/go-upload"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
	"github.com/mattn/go-colorable"
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

	region, err := apps.GetAppRegion(appID)

	if err != nil {
		return nil, err
	}

	return UploadFileEx(appInfo.AppID, appInfo.AppKey, region, filePath)
}

// UploadFileEx upload specific file to LeanCloud
func UploadFileEx(appID string, appKey string, region regions.Region, filePath string) (*upload.File, error) {
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
	bar.Output = colorable.NewColorableStderr()
	bar.Prefix(color.GreenString("[INFO]") + " 上传文件")
	bar.Start()

	// qiniu want a io.ReadSeeker to get file's size
	readSeeker := &fileBarReaderSeeker{
		file:   f,
		reader: bar.NewProxyReader(f),
	}

	file, err := upload.Upload(fileName, mimeType, readSeeker, &upload.Options{
		AppID:     appID,
		AppKey:    appKey,
		APIServer: GetAppAPIURL(region, appID),
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

// upload code zip package to a concentrated app in specific region
func UploadToRepoStorage(region regions.Region, filePath string) (*upload.File, error) {
	appID, appKey, uploadRegion := getRepoStorageInfo(region)
	return UploadFileEx(appID, appKey, uploadRegion, filePath)
}

func DeleteFromRepoStorage(region regions.Region, objectID string) error {
	appID, appKey, uploadRegion := getRepoStorageInfo(region)

	client := NewClientByRegion(uploadRegion)
	opts, err := client.options()
	if err != nil {
		return err
	}
	opts.Headers["X-LC-Id"] = appID
	opts.Headers["X-LC-Key"] = appKey
	_, err = client.delete("/1.1/files/"+objectID, opts)
	return err
}

func getRepoStorageInfo(region regions.Region) (appId string, appKey string, uploadRegion regions.Region) {
	switch region {
	case regions.US:
		return "iuuztdrr4mj683kbsmwoalt1roaypb5d25eu0f23lrfsthgn", "exhqkdnvtjw34p5670r4zlofdsc91likhzfxmr9jz7vnbc07", regions.US
	default:
		return "x7WmVG0x63V6u8MCYM8qxKo8-gzGzoHsz", "PcDNOjiEpYc0DTz2E9kb5fvu", regions.CN
	}
}
