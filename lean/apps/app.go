package apps

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/leancloud/lean-cli/lean/output"
)

var (
	// ErrNoAppLinked means no app was linked to the project
	ErrNoAppLinked = errors.New("No Leancloud Application was linked to the project")
)

func appDirPath(projectPath string) string {
	return filepath.Join(projectPath, ".leancloud")
}

func currentAppIDFilePath(projectPath string) string {
	return filepath.Join(appDirPath(projectPath), "current_app_id")
}

// LinkApp will write the specific appID to ${projectPath}/.leancloud/current_app_id
func LinkApp(projectPath string, appID string) error {
	err := os.Mkdir(appDirPath(projectPath), 0775)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return ioutil.WriteFile(currentAppIDFilePath(projectPath), []byte(appID), 0644)
}

// GetCurrentAppID will return the content of ${projectPath}/.leancloud/current_app_id
func GetCurrentAppID(projectPath string) (string, error) {
	content, err := ioutil.ReadFile(currentAppIDFilePath(projectPath))
	if os.IsNotExist(err) {
		return migrateLegencyProjectConfig(projectPath)
	}
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func getLegencyAppID(projectPath string) (string, error) {
	content, err := ioutil.ReadFile(filepath.Join(projectPath, ".avoscloud", "apps.json"))
	if err != nil {
		return "", ErrNoAppLinked
	}

	var apps map[string]string
	err = json.Unmarshal(content, &apps)
	if err != nil {
		return "", ErrNoAppLinked
	}

	if len(apps) == 0 {
		return "", ErrNoAppLinked
	}

	if len(apps) == 1 {
		for _, v := range apps {
			return v, nil
		}
	}

	content, err = ioutil.ReadFile(filepath.Join(projectPath, ".avoscloud", "curr_app"))
	if err != nil {
		return "", ErrNoAppLinked
	}
	appName := string(content)

	appID, ok := apps[appName]
	if !ok {
		return "", ErrNoAppLinked
	}
	return appID, nil
}

func migrateLegencyProjectConfig(projectPath string) (string, error) {
	appID, err := getLegencyAppID(projectPath)
	if err != nil {
		return "", err
	}

	op := output.NewOutput(os.Stdout)
	op.Write("检测到旧版命令行工具项目配置，正在迁移")

	err = LinkApp(projectPath, appID)
	if err != nil {
		op.Failed()
		return "", err
	}
	op.Successed()

	log.Printf("> 迁移完毕，`%s`可进行删除\r\n", filepath.Join(projectPath, ".avoscloud"))

	return appID, nil
}
