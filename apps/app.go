package apps

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

func currentGroupFilePath(projectPath string) string {
	return filepath.Join(appDirPath(projectPath), "current_group")
}

// LinkApp will write the specific appID to ${projectPath}/.leancloud/current_app_id
func LinkApp(projectPath string, appID string) error {
	err := os.Mkdir(appDirPath(projectPath), 0775)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return ioutil.WriteFile(currentAppIDFilePath(projectPath), []byte(appID), 0644)
}

// LinkGroup will write the specific groupName to ${projectPath}/.leancloud/current_group
func LinkGroup(projectPath string, groupName string) error {
	err := os.Mkdir(appDirPath(projectPath), 0775)
	if err != nil && !os.IsExist(err) {
		return err
	}

	return ioutil.WriteFile(currentGroupFilePath(projectPath), []byte(groupName), 0644)
}

// GetCurrentAppID will return the content of ${projectPath}/.leancloud/current_app_id
func GetCurrentAppID(projectPath string) (string, error) {
	content, err := ioutil.ReadFile(currentAppIDFilePath(projectPath))
	if err != nil && os.IsNotExist(err) {
		return "", ErrNoAppLinked
	} else if err != nil {
		return "", err
	}
	appID := strings.TrimSpace(string(content))
	if appID == "" {
		msg := "Invalid app, please check the `.leancloud/current_app_id`'s content."
		return "", errors.New(msg)
	}

	if _, err = GetAppRegion(appID); err != nil {
		return "", err
	}

	return appID, nil
}

// GetCurrentGroup returns the content of ${projectPath}/.leancloud/current_group if it exists,
// or migrate the project's primary group.
func GetCurrentGroup(projectPath string) (string, error) {
	content, err := ioutil.ReadFile(currentGroupFilePath(projectPath))
	if err != nil {
		return "", err
	}
	groupName := strings.TrimSpace(string(content))
	if groupName == "" {
		msg := "Invalid group, please check the `.leancloud/current_group`'s content."
		return "", errors.New(msg)
	}
	return groupName, nil
}
