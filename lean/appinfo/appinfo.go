package appinfo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

type AppInfo struct {
	AppId     string `json:"appId"`
	AppKey    string `json:"appKey"`
	MasterKey string `json:"masterKey"`
}

func NewFromInput() error {
	var appId string
	var appKey string
	var masterKey string

	fmt.Println("Please input your APP ID:")
	fmt.Scanf("%s", &appId)

	fmt.Println("Please input your APP KEY:")
	fmt.Scanf("%s", &appKey)

	fmt.Println("Please input your MASTER KEY:")
	fmt.Scanf("%s", &masterKey)

	appInfo := &AppInfo{
		AppId:     appId,
		AppKey:    appKey,
		MasterKey: masterKey,
	}
	if !appInfo.validate() {
		return errors.New("Invalid input")
	}

	appInfo.Store()
	return nil
}

func PrintFromLocal() error {
	appInfo, err := loadAppInfoFromLocal()
	if err != nil {
		return err
	}
	fmt.Println("Current App ID: ", appInfo.AppId)
	return err
}

func loadAppInfoFromLocal() (*AppInfo, error) {
	filePath := ".leanengine" + string(os.PathSeparator) + "keys.json"
	content, err := ioutil.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil, errors.New(filePath + " not exist, please make sure your are in a LeanEngine project folder.")
	}

	var appInfo AppInfo
	err = json.Unmarshal(content, &appInfo)
	return &appInfo, err
}

func (appInfo *AppInfo) validate() bool {
	return true // TODO
}

func (appInfo *AppInfo) Store() error {
	value, err := json.Marshal(appInfo)
	if err != nil {
		return err
	}

	fmt.Println(string(value))

	err = os.Mkdir(".leanengine", 0777)
	if !os.IsExist(err) && err != nil {
		return nil
	}

	file, err := os.Create(".leanengine" + string(os.PathSeparator) + "keys.json")
	if err != nil {
		return nil
	}
	defer file.Close()

	file.Write(value)

	return nil
}
