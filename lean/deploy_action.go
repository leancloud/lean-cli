package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/jhoonb/archivex"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/console"
	"github.com/leancloud/lean-cli/lean/utils"
)

func determineGroupName(appID string) (string, error) {
	op.Write("获取应用信息")
	info, err := api.GetAppInfo(appID)
	if err != nil {
		op.Failed()
		return "", err
	}
	op.Successed()
	log.Println("> 准备部署至目标应用：" + color.RedString(info.AppName) + " (" + appID + ")")
	mode := info.LeanEngineMode

	op.Write("获取应用分组信息")
	groups, err := api.GetGroups(appID)
	if err != nil {
		op.Failed()
		return "", err
	}
	op.Successed()

	groupName, found, err := linq.From(groups).Where(func(group linq.T) (bool, error) {
		groupName := group.(*api.GetGroupsResult).GroupName
		if mode == "free" {
			return groupName != "staging", nil
		}
		return groupName == "staging", nil
	}).Select(func(group linq.T) (linq.T, error) {
		return group.(*api.GetGroupsResult).GroupName, nil
	}).First()
	if err != nil {
		return "", err
	}
	if !found {
		return "", errors.New("group not found")
	}
	return groupName.(string), nil
}

func uploadProject(appID string, repoPath string) (*api.UploadFileResult, error) {
	// TODO: ignore files

	fileDir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(fileDir, "leanengine.zip")

	runtime, err := console.DetectRuntime(repoPath)
	log.Println(repoPath)
	files, err := utils.MatchFiles(repoPath, runtime.DeployFiles.Includes, runtime.DeployFiles.Excludes)
	if err != nil {
		return nil, err
	}
	op.Write("压缩项目文件 ...")
	zip := new(archivex.ZipFile)
	func() {
		defer zip.Close()
		zip.Create(filePath)
		zip.AddAll(repoPath, false)
		for _, f := range files {
			err := zip.AddFile(f)
			if err != nil {
				panic(err)
			}
		}
	}()
	op.Successed()

	file, err := api.UploadFile(appID, filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func deployFromLocal(appID string, groupName string) error {
	file, err := uploadProject(appID, ".")
	if err != nil {
		return err
	}

	defer func() {
		op.Write("删除临时文件")
		err := api.DeleteFile(appID, file.ObjectID)
		if err != nil {
			op.Failed()
		} else {
			op.Successed()
		}
	}()

	eventTok, err := api.DeployAppFromFile("", groupName, file.URL)
	ok, err := api.PollEvents(appID, eventTok, os.Stdout)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}

func deployFromGit(appID string, groupName string) error {
	eventTok, err := api.DeployAppFromGit("", groupName)
	if err != nil {
		return err
	}
	ok, err := api.PollEvents(appID, eventTok, os.Stdout)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("部署失败", 1)
	}
	return nil
}

func deployAction(*cli.Context) error {
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		log.Fatalln("没有关联任何 app，请使用 lean checkout 来关联应用。")
	}
	if err != nil {
		return newCliError(err)
	}

	groupName, err := determineGroupName(appID)
	if err != nil {
		op.Failed()
		return newCliError(err)
	}

	if groupName == "staging" {
		log.Println("> 准备部署应用到预备环境")
	} else {
		log.Println("> 准备部署应用到生产环境: " + groupName)
	}

	if isDeployFromGit {
		err = deployFromGit(appID, groupName)
		if err != nil {
			return newCliError(err)
		}
	} else {
		err = deployFromLocal(appID, groupName)
		if err != nil {
			return newCliError(err)
		}
	}
	return nil
}
