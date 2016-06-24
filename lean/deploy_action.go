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

func deployGroupName(appInfo apps.AppInfo) (string, error) {
	client := api.Client{
		AppID:     appInfo.AppID,
		MasterKey: appInfo.MasterKey,
		Region:    api.RegionCN,
	}

	engineInfo, err := client.EngineInfo()
	if err != nil {
		return "", err
	}
	mode := engineInfo.Get("mode").MustString()

	groups, err := client.Groups()
	if err != nil {
		return "", err
	}

	groupName, found, err := linq.From(groups.MustArray()).Where(func(_group linq.T) (bool, error) {
		groupName := _group.(map[string]interface{})["groupName"].(string)
		if mode == "free" {
			return groupName != "staging", nil
		}
		return groupName == "staging", nil
	}).Select(func(group linq.T) (linq.T, error) {
		return group.(map[string]interface{})["groupName"], nil
	}).First()
	if err != nil {
		return "", err
	}
	if !found {
		return "", errors.New("group not found")
	}
	return groupName.(string), nil
}

func uploadProject(appInfo apps.AppInfo, repoPath string) (*api.File, error) {
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
	client := api.Client{
		AppID:     appInfo.AppID,
		MasterKey: appInfo.MasterKey,
		Region:    api.RegionCN,
	}
	file, err := client.UploadFile(filePath)
	utils.CheckError(err)

	return file, nil
}

func deployFromLocal(appID string) {
	appInfo, err := apps.GetAppInfo(appID)
	utils.CheckError(err)

	groupName, err := deployGroupName(appInfo)
	utils.CheckError(err)

	if groupName == "staging" {
		log.Println("准备部署应用到预备环境")
	} else {
		log.Println("准备部署应用到生产环境: " + groupName)
	}

	file, err := uploadProject(appInfo, "")
	utils.CheckError(err)

	client := api.Client{
		AppID:     appInfo.AppID,
		MasterKey: appInfo.MasterKey,
		Region:    api.RegionCN,
	}

	_, err = client.BuildAndDeploy(groupName, file.URL)
	utils.CheckError(err)

	err = client.DeleteFile(file.ID)
	utils.CheckError(err)
}

func deployAction(*cli.Context) {
	_apps, err := apps.GetApps(".")
	utils.CheckError(err)
	if len(_apps) == 0 {
		log.Fatalln("没有关联任何 app，请使用 lean app add 来关联应用。")
	}

	// TODO: specific app
	app := _apps[0]

	deployFromLocal(app.AppID)
}
