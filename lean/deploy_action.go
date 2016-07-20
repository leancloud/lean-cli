package main

import (
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/ahmetalpbalkan/go-linq"
	"github.com/codegangsta/cli"
	"github.com/jhoonb/archivex"
	"github.com/leancloud/lean-cli/lean/api"
	"github.com/leancloud/lean-cli/lean/apps"
	"github.com/leancloud/lean-cli/lean/utils"
)

func determineGroupName(appID string) (string, error) {
	info, err := api.GetAppInfo(appID)
	if err != nil {
		return "", err
	}
	mode := info.LeanEngineMode

	groups, err := api.GetGroups(appID)
	if err != nil {
		return "", err
	}

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

func uploadProject(appInfo *apps.AppInfo, repoPath string) (*api.UploadFileResult, error) {
	// TODO: ignore files

	fileDir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(fileDir, "leanengine.zip")
	println(filePath)

	log.Println("压缩项目文件 ...")
	zip := new(archivex.ZipFile)
	func() {
		defer zip.Close()
		zip.Create(filePath)
		zip.AddAll(repoPath, false)
	}()

	log.Println("上传项目文件 ...")
	file, err := api.UploadFile(appInfo.AppID, filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func deployFromLocal(appInfo *apps.AppInfo, groupName string) error {
	file, err := uploadProject(appInfo, "")
	if err != nil {
		return err
	}

	// defer func() {
	// 	err := api.DeleteFile(appInfo.AppID, file.ObjectID)
	// 	if err != nil {
	// 		log.Println("删除临时文件失败：", err)
	// 	} else {
	// 		log.Println("删除临时文件成功")
	// 	}
	// }()

	tok, err := api.DeployAppFromFile("", groupName, file.URL)
	log.Println(tok)
	return err
}

func deployAction(*cli.Context) error {
	_apps, err := apps.LinkedApps(".")
	utils.CheckError(err)
	if len(_apps) == 0 {
		log.Fatalln("没有关联任何 app，请使用 lean app add 来关联应用。")
	}

	// TODO: specific app
	app := _apps[0]

	appInfo, err := apps.GetAppInfo(app.AppID)
	if err != nil {
		return newCliError(err)
	}

	groupName, err := determineGroupName(appInfo.AppID)
	if err != nil {
		return newCliError(err)
	}

	if groupName == "staging" {
		log.Println("准备部署应用到预备环境")
	} else {
		log.Println("准备部署应用到生产环境: " + groupName)
	}

	if isDeployFromGit {
		eventTok, err := api.DeployAppFromGit("", groupName)
		if err != nil {
			return newCliError(err)
		}
		log.Println(eventTok)
		return nil
	}
	deployFromLocal(appInfo, groupName)
	return nil
}
