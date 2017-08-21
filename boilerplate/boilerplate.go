package boilerplate

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/logger"
	"github.com/leancloud/lean-cli/utils"
	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
	"github.com/mattn/go-colorable"
)

// don't know why archive/zip.Reader.File[0].FileInfo().IsDir() always return true,
// this is a trick hack to void this.
func isDir(path string) bool {
	return os.IsPathSeparator(path[len(path)-1])
}

func extractAndWriteFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(dest, f.Name)

	if isDir(f.Name) {
		if err := os.MkdirAll(path, f.Mode()); err != nil {
			return err
		}
	} else {
		// Use os.Create() since Zip don't store file permissions.
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}
	return nil
}

// FetchRepo will download the boilerplate from remote and extract to ${appName}/folder
func FetchRepo(boil *Boilerplate, appName string, appID string) error {
	if err := os.Mkdir(appName, 0775); err != nil {
		return err
	}

	repoURL := "https://lcinternal-cloud-code-update.leanapp.cn" + boil.URL

	dir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	resp, err := grequests.Get(repoURL, &grequests.RequestOptions{
		UserAgent: "LeanCloud-CLI/" + version.Version,
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode != 200 {
		return errors.New(utils.FormatServerErrorResult(resp.String()))
	}
	zipFilePath := filepath.Join(dir, "getting-started.zip")
	err = DownloadToFile(resp, zipFilePath)
	if err != nil {
		return err
	}

	logger.Info("正在创建项目...")

	zipFile, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	for _, f := range zipFile.File {
		err := extractAndWriteFile(f, appName)
		if err != nil {
			return err
		}
	}

	logger.Info("创建", boil.Name, "项目成功，更多关于", boil.Name, "的文档请参考官网：", boil.Homepage)
	return nil
}

type Category struct {
	Name         string        `json:"name"`
	Boilerplates []Boilerplate `json:"boilerplates"`
}

type Boilerplate struct {
	Name     string `json:"name"`
	Homepage string `json:"homepage"`
	URL      string `json:"url"`
}

// GetBoilerplates returns all the boilerplate with name and url
func GetBoilerplates() ([]Category, error) {
	resp, err := grequests.Get("https://lcinternal-cloud-code-update.leanapp.cn/boilerplates.json", &grequests.RequestOptions{
		UserAgent: "LeanCloud-CLI/" + version.Version,
	})
	if err != nil {
		return nil, err
	}
	var result []Category
	err = resp.JSON(&result)
	return result, err
}

// DownloadToFile allows you to download the contents of the response to a file
func DownloadToFile(r *grequests.Response, fileName string) error {

	if r.Error != nil {
		return r.Error
	}

	fd, err := os.Create(fileName)

	if err != nil {
		return err
	}

	defer r.Close() // This is a noop if we use the internal ByteBuffer
	defer fd.Close()

	if length, err := strconv.Atoi(r.Header.Get("Content-Length")); err == nil {
		bar := pb.New(length).SetUnits(pb.U_BYTES).SetMaxWidth(80)
		bar.Output = colorable.NewColorableStderr()
		bar.Prefix(color.GreenString("[INFO]") + " 下载模版文件")
		bar.Start()
		defer bar.Finish()
		reader := bar.NewProxyReader(r)
		if _, err := io.Copy(fd, reader); err != nil && err != io.EOF {
			return err
		}
	} else {
		if _, err := io.Copy(fd, r); err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}
