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
	"github.com/leancloud/lean-cli/lean/output"
	"github.com/leancloud/lean-cli/lean/utils"
	"github.com/levigross/grequests"
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
		os.MkdirAll(path, f.Mode())
	} else {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
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
	op := output.NewOutput(os.Stdout)
	utils.CheckError(os.Mkdir(appName, 0775))

	repoURL := "https://lcinternal-cloud-code-update.leanapp.cn/" + boil.URL

	dir, err := ioutil.TempDir("", "leanengine")
	utils.CheckError(err)
	defer os.RemoveAll(dir)

	resp, err := grequests.Get(repoURL, nil)
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode != 200 {
		return errors.New(utils.FormatServerErrorResult(resp.String()))
	}
	zipFilePath := filepath.Join(dir, "getting-started.zip")
	DownloadToFile(resp, zipFilePath)

	op.Write("正在创建项目...")

	zipFile, err := zip.OpenReader(zipFilePath)
	utils.CheckError(err)
	defer zipFile.Close()
	for _, f := range zipFile.File {
		err := extractAndWriteFile(f, appName)
		if err != nil {
			op.Failed()
			return err
		}
	}

	op.Successed()

	return nil
}

// Boilerplate is GetBoilerplateList's result type
type Boilerplate struct {
	Name string
	URL  string
}

// GetBoilerplateList returns all the boilerplate with name and url
func GetBoilerplateList() ([]*Boilerplate, error) {
	resp, err := grequests.Get("https://lcinternal-cloud-code-update.leanapp.cn/", nil)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*Boilerplate)
	err = resp.JSON(&result)
	if err != nil {
		return nil, err
	}
	var boils []*Boilerplate
	for _, boil := range result {
		boils = append(boils, boil)
	}
	return boils, nil
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
		bar.Prefix("> 下载模版文件")
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
