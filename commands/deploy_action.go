package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/aisk/logp"
	"github.com/leancloud/go-upload"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/runtimes"
	"github.com/leancloud/lean-cli/utils"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

func monitorInterrupt(appId, eventTok string, signalCh chan os.Signal) {
	for i := 0; i < 2; i++ {
		_, ok := <-signalCh
		if !ok{
			return
		}
		switch i {
		case 0:
			logp.Warn("正在取消部署...")
			go func() {
				err := api.CancelDeployByToken(appId, eventTok)
				if err != nil {
					logp.Error(err)
				} else {
					logp.Info("取消部署成功")
				}
				os.Exit(0)
			}()
		case 1:
			os.Exit(0)
		}
	}
}

func uploadProject(appID string, repoPath string, ignoreFilePath string) (*upload.File, error) {
	fileDir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return nil, err
	}

	archiveFile := filepath.Join(fileDir, "leanengine.zip")

	runtime, err := runtimes.DetectRuntime(repoPath)
	if err != nil {
		return nil, err
	}

	err = runtime.ArchiveUploadFiles(archiveFile, ignoreFilePath)
	if err != nil {
		return nil, err
	}

	file, err := api.UploadFile(appID, archiveFile)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func uploadWar(appID string, repoPath string) (*upload.File, error) {
	var warPath string
	files, err := ioutil.ReadDir(filepath.Join(repoPath, "target"))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".war") && !file.IsDir() {
			warPath = filepath.Join(repoPath, "target", file.Name())
		}
	}
	if warPath == "" {
		return nil, errors.New("在 ./target 目录没有找到 war 文件")
	}

	logp.Info("找到默认的 war 文件：", warPath)

	fileDir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return nil, err
	}
	archivePath := filepath.Join(fileDir, "ROOT.war.zip")

	file := []struct{ Name, Path string }{{
		Name: "ROOT.war",
		Path: warPath,
	}}
	if err = utils.ArchiveFiles(archivePath, file); err != nil {
		return nil, err
	}

	return api.UploadFile(appID, archivePath)
}

func deployFromLocal(isDeployFromJavaWar bool, ignoreFilePath string, keepFile bool, opts *deployOptions) (string, error) {
	var file *upload.File
	var err error
	if isDeployFromJavaWar {
		file, err = uploadWar(opts.appID, ".")
	} else {
		file, err = uploadProject(opts.appID, ".", ignoreFilePath)
		if err != nil {
			return "", err
		}
	}

	if !keepFile {
		defer func() {
			logp.Info("删除临时文件")
			err := api.DeleteFile(opts.appID, file.ObjectID)
			if err != nil {
				logp.Error(err)
			}
		}()
	}
	return api.DeployAppFromFile(opts.appID, opts.groupName, opts.prod, file.URL, opts.message, opts.noDepsCache)
}

func deployFromGit(revision string, opts *deployOptions) (string, error) {
	return api.DeployAppFromGit(opts.appID, opts.groupName, opts.prod, revision, opts.noDepsCache)
}

func pollEvents(appID, eventTok string) error{
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh)
	go monitorInterrupt(appID, eventTok, signalCh)
	defer func(){
		signal.Stop(signalCh)
		close(signalCh)
	}()
	ok, err := api.PollEvents(appID, eventTok)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}

func deployAction(c *cli.Context) error {
	version.PrintCurrentVersion()
	isDeployFromGit := c.Bool("g")
	isDeployFromJavaWar := c.Bool("war")
	ignoreFilePath := c.String("leanignore")
	noDepsCache := c.Bool("no-cache")
	message := c.String("message")
	keepFile := c.Bool("keep-deploy-file")
	revision := c.String("revision")

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}

	logp.Info("获取应用信息 ...")
	region, err := api.GetAppRegion(appID)
	if err != nil {
		return err
	}
	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return err
	}
	engineInfo, err := api.GetEngineInfo(appID)
	if err != nil {
		return err
	}

	prod := 0
	if engineInfo.Mode == "prod" {
		logp.Infof("准备部署应用 %s(%s) 到 %s 节点分组 %s 预备环境\r\n", appInfo.AppName, appID, region, groupName)
	} else if engineInfo.Mode == "free" {
		prod = 1
		logp.Infof("准备部署应用 %s(%s) 到 %s 节点分组 %s 生产环境\r\n", appInfo.AppName, appID, region, groupName)
	} else {
		panic(fmt.Sprintf("invalid engine mode: %s", engineInfo.Mode))
	}

	opts := &deployOptions{
		appID:       appID,
		groupName:   groupName,
		message:     message,
		noDepsCache: noDepsCache,
		prod:        prod,
	}
	var eventTok string
	if isDeployFromGit {
		eventTok, err = deployFromGit(revision, opts)
		if err != nil {
			return err
		}
	} else {
		eventTok, err = deployFromLocal(isDeployFromJavaWar, ignoreFilePath, keepFile, opts)
		if err != nil {
			return err
		}
	}
	return pollEvents(opts.appID, eventTok)
}

type deployOptions struct {
	appID       string
	groupName   string
	message     string
	noDepsCache bool
	prod        int
}
