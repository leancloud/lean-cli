package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aisk/logp"
	"github.com/leancloud/go-upload"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/runtimes"
	"github.com/leancloud/lean-cli/utils"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

const (
	uploadRepoAppID  = "x7WmVG0x63V6u8MCYM8qxKo8-gzGzoHsz"
	uploadRepoAppKey = "PcDNOjiEpYc0DTz2E9kb5fvu"
	uploadRepoRegion = regions.CN
)

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

	file, err := api.UploadFileEx(uploadRepoAppID, uploadRepoAppKey, uploadRepoRegion, archiveFile)
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

	return api.UploadFileEx(uploadRepoAppID, uploadRepoAppKey, uploadRepoRegion, archivePath)
}

func deployFromLocal(isDeployFromJavaWar bool, ignoreFilePath string, keepFile bool, opts *deployOptions) error {
	var file *upload.File
	var err error
	if isDeployFromJavaWar {
		file, err = uploadWar(opts.appID, ".")
	} else {
		file, err = uploadProject(opts.appID, ".", ignoreFilePath)
		if err != nil {
			return err
		}
	}

	if !keepFile {
		defer func() {
			logp.Info("删除临时文件")
			err := api.DeleteFileEx(uploadRepoAppID, uploadRepoAppKey, uploadRepoRegion, file.ObjectID)
			if err != nil {
				logp.Error(err)
			}
		}()
	}

	eventTok, err := api.DeployAppFromFile(opts.appID, opts.groupName, opts.prod, file.URL, opts.message, opts.noDepsCache, opts.mode)
	if err != nil {
		return err
	}
	ok, err := api.PollEvents(opts.appID, eventTok)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}

func deployFromGit(revision string, opts *deployOptions) error {
	eventTok, err := api.DeployAppFromGit(opts.appID, opts.groupName, opts.prod, revision, opts.noDepsCache, opts.mode)
	if err != nil {
		return err
	}
	ok, err := api.PollEvents(opts.appID, eventTok)
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

	if message == "" {
		_, err := exec.LookPath("git")

		if err == nil {
			messageBuf, err := exec.Command("git", "log", "-1", "--no-color", "--pretty=%B").CombinedOutput()
			messageStr := string(messageBuf)

			if err != nil && strings.Contains(messageStr, "Not a git repository") {
				// Ignore
			} else if err != nil {
				logp.Error(err)
			} else {
				message = "WIP on: " + strings.TrimSpace(messageStr)
			}
		}
	}

	if message == "" {
		message = "从命令行工具构建"
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}

	logp.Info("获取应用信息 ...")
	region, err := apps.GetAppRegion(appID)
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

	var deployMode string

	if c.Bool("atomic") {
		deployMode = api.DEPLOY_ATOMIC
	} else {
		deployMode = api.DEPLOY_SMOOTHLY
	}

	opts := &deployOptions{
		appID:       appID,
		groupName:   groupName,
		message:     message,
		noDepsCache: noDepsCache,
		prod:        prod,
		mode:        deployMode,
	}

	if isDeployFromGit {
		err = deployFromGit(revision, opts)
		if err != nil {
			return err
		}
	} else {
		err = deployFromLocal(isDeployFromJavaWar, ignoreFilePath, keepFile, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

type deployOptions struct {
	appID       string
	groupName   string
	message     string
	noDepsCache bool
	prod        int
	mode        string
}
