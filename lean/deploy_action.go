package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/leancloud/go-upload"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/runtimes"
)

func determineGroupName(appID string) (string, error) {
	spinner := chrysanthemum.New("获取应用信息").Start()

	info, err := api.GetAppInfo(appID)
	if err != nil {
		spinner.Failed()
		return "", err
	}
	spinner.Successed()
	chrysanthemum.Printf("准备部署至目标应用：%s (%s)\r\n", color.RedString(info.AppName), appID)
	mode := info.LeanEngineMode

	spinner = chrysanthemum.New("获取应用分组信息").Start()
	groups, err := api.GetGroups(appID)
	if err != nil {
		spinner.Failed()
		return "", err
	}
	spinner.Successed()

	groupName := linq.From(groups).Where(func(group interface{}) bool {
		groupName := group.(*api.GetGroupsResult).GroupName
		if mode == "free" {
			return groupName != "staging"
		}
		return groupName == "staging"
	}).Select(func(group interface{}) interface{} {
		return group.(*api.GetGroupsResult).GroupName
	}).First()
	return groupName.(string), nil
}

func uploadProject(appID string, repoPath string, isDeployFromJavaWar bool, ignoreFilePath string) (*upload.File, error) {
	fileDir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return nil, err
	}

	archiveFile := filepath.Join(fileDir, "leanengine.zip")

	runtime, err := runtimes.DetectRuntime(repoPath)
	if err != nil {
		return nil, err
	}

	runtime.ArchiveUploadFiles(archiveFile, isDeployFromJavaWar, ignoreFilePath)

	file, err := api.UploadFile(appID, archiveFile)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func deployFromLocal(isDeployFromJavaWar bool, ignoreFilePath string, keepFile bool, opts *deployOptions) error {
	file, err := uploadProject(opts.appID, ".", isDeployFromJavaWar, ignoreFilePath)
	if err != nil {
		return err
	}

	if !keepFile {
		defer func() {
			spinner := chrysanthemum.New("删除临时文件").Start()
			err = api.DeleteFile(opts.appID, file.ObjectID)
			if err != nil {
				spinner.Failed()
			} else {
				spinner.Successed()
			}
		}()
	}

	eventTok, err := api.DeployAppFromFile(opts.appID, ".", opts.groupName, file.URL, opts.message, opts.noDepsCache)
	ok, err := api.PollEvents(opts.appID, eventTok, os.Stdout)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}

func deployFromGit(revision string, opts *deployOptions) error {
	eventTok, err := api.DeployAppFromGit(opts.appID, ".", opts.groupName, revision, opts.noDepsCache)
	if err != nil {
		return err
	}
	ok, err := api.PollEvents(opts.appID, eventTok, os.Stdout)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}

func deployAction(c *cli.Context) error {
	isDeployFromGit := c.Bool("g")
	isDeployFromJavaWar := c.Bool("war")
	ignoreFilePath := c.String("leanignore")
	noDepsCache := c.Bool("no-cache")
	message := c.String("message")
	keepFile := c.Bool("keep-deploy-file")
	revision := c.String("revision")

	println("revision:", revision)

	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		return cli.NewExitError("没有关联任何 app，请使用 lean checkout 来关联应用。", 1)
	}
	if err != nil {
		return newCliError(err)
	}

	groupName, err := determineGroupName(appID)
	if err != nil {
		return newCliError(err)
	}

	if groupName == "staging" {
		chrysanthemum.Printf("准备部署应用到预备环境\r\n")
	} else {
		chrysanthemum.Printf("准备部署应用到生产环境: %s\r\n", groupName)
	}

	opts := &deployOptions{
		appID:       appID,
		groupName:   groupName,
		message:     message,
		noDepsCache: noDepsCache,
	}

	if isDeployFromGit {
		err = deployFromGit(revision, opts)
		if err != nil {
			return newCliError(err)
		}
	} else {
		err = deployFromLocal(isDeployFromJavaWar, ignoreFilePath, keepFile, opts)
		if err != nil {
			return newCliError(err)
		}
	}
	return nil
}

type deployOptions struct {
	appID       string
	groupName   string
	message     string
	noDepsCache bool
}
