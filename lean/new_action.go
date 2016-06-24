package main

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aisk/wizard"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/utils"
	"github.com/levigross/grequests"
)

const (
	runtimePython = iota
	runtimeNodeJS
	runtimePHP
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

func askNewAppInfo() (string, string, int) {
	appID := new(string)
	masterKey := new(string)
	runtimeType := 0

	log.Println("开始输入应用信息，这些信息可以从'开发者平台的应用设置 -> 应用 key'里找到。")

	wizard.Ask([]wizard.Question{
		{
			Content: "请输入应用的 Application ID:",
			Input: &wizard.Input{
				Hidden: false,
				Result: appID,
			},
		},
	})

	// TODO: get the masterKey from local first
	wizard.Ask([]wizard.Question{
		{
			Content: "请输入应用的 Master Key:",
			Input: &wizard.Input{
				Hidden: true,
				Result: masterKey,
			},
		},
	})

	wizard.Ask([]wizard.Question{
		{
			Content: "请选择项目语言:",
			Answers: []wizard.Answer{
				{
					Content: "Python",
					Handler: func() {
						runtimeType = runtimePython
					},
				}, {
					Content: "Node.js",
					Handler: func() {
						runtimeType = runtimeNodeJS
					},
				},
				// {
				// 	Content: "PHP",
				// 	Handler: func() {
				// 		runtimeType = runtimePHP
				// 	},
				// },
			},
		},
	})
	return *appID, *masterKey, runtimeType
}

func fetchRepo(t int, appName string, appID string) error {
	utils.CheckError(os.Mkdir(appName, 0700))

	repoURL := map[int]string{
		runtimePython: "http://lcinternal-cloud-code-update.leanapp.cn/python-getting-started.zip",
		runtimeNodeJS: "http://lcinternal-cloud-code-update.leanapp.cn/node-js-getting-started.zip",
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
		utils.CheckError(extractAndWriteFile(f, appName))
	}

	if err := apps.AddApp(appName, appName, appID); err != nil {
		return err
	}

	log.Println("创建项目完成")

	return nil
}

func newAction(*cli.Context) {
	appID, masterKey, runtime := askNewAppInfo()

	client := api.Client{AppID: appID, MasterKey: masterKey, Region: api.RegionCN}

	detail, err := client.AppDetail()
	utils.CheckError(err)
	appName := detail.Get("app_name").MustString()

	err = fetchRepo(runtime, appName, appID)
	utils.CheckError(err)
}
