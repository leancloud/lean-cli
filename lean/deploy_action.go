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

func uploadProject(appID string, repoPath string) (*api.UploadFileResult, error) {
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
	file, err := api.UploadFile(appID, filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func deployFromLocal(appID string, groupName string) error {
	file, err := uploadProject(appID, "")
	if err != nil {
		return err
	}

	// TODO: remove after deploy finished
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
	// TODO: specific app
	appID, err := apps.GetCurrentAppID("")
	if err == apps.ErrNoAppLinked {
		log.Fatalln("没有关联任何 app，请使用 lean switch 来关联应用。")
	}

	if err != nil {
		return newCliError(err)
	}

	groupName, err := determineGroupName(appID)
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
	deployFromLocal(appID, groupName)
	return nil
}
