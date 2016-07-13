package boilerplate

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/utils"
	"github.com/levigross/grequests"
)

// stands for different language runtime boilerplate
const (
	invalid = iota
	Python
	NodeJS
	PHP
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

// FetchRepo will download the boilerplate from remote and extract to ${appName}/ folder
func FetchRepo(t int, appName string, appID string) error {
	utils.CheckError(os.Mkdir(appName, 0700))

	repoURL := map[int]string{
		Python: "http://lcinternal-cloud-code-update.leanapp.cn/python-getting-started.zip",
		NodeJS: "http://lcinternal-cloud-code-update.leanapp.cn/node-js-getting-started.zip",
	}[t]

	dir, err := ioutil.TempDir("", "leanengine")
	utils.CheckError(err)
	defer os.RemoveAll(dir)

	log.Println("正在下载项目模版...")

	resp, err := grequests.Get(repoURL, nil)
	if err != nil {
		return err
	}
	defer resp.Close()
	if resp.StatusCode != 200 {
		return errors.New(utils.FormatServerErrorResult(resp.String()))
	}

	log.Println("下载完成")

	zipFilePath := filepath.Join(dir, "getting-started.zip")
	resp.DownloadToFile(zipFilePath)

	log.Println("正在创建项目...")

	zipFile, err := zip.OpenReader(zipFilePath)
	utils.CheckError(err)
	defer zipFile.Close()
	for _, f := range zipFile.File {
		err := extractAndWriteFile(f, appName)
		if err != nil {
			return err
		}
	}

	if err := apps.AddApp(appName, appName, appID); err != nil {
		return err
	}

	log.Println("创建项目完成")

	return nil
}
