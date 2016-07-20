package apps

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
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
	err := os.Mkdir(appDirPath(projectPath), 0700)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return ioutil.WriteFile(currentAppIDFilePath(projectPath), []byte(appID), 0700)
}

// GetCurrentAppID will return the content of ${projectPath}/.leancloud/current_app_id
func GetCurrentAppID(projectPath string) (string, error) {
	content, err := ioutil.ReadFile(currentAppIDFilePath(projectPath))
	if err != nil {
		return "", err
	}
	return string(content), nil
}
