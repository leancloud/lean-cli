package apps

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bitly/go-simplejson"
)

// App ...
type App struct {
	Name string
	ID   string
}

func appDirPath(projectPath string) string {
	return filepath.Join(projectPath, ".avoscloud")
}

func appFilePath(projectPath string) string {
	return filepath.Join(appDirPath(projectPath), "apps.json")
}

// GetApps returns the current project's linked apps
func GetApps(projectPath string) (apps []App, err error) {
	content, err := ioutil.ReadFile(appFilePath(projectPath))
	if os.IsNotExist(err) {
		return apps, nil
	}
	if err != nil {
		return
	}

	json, err := simplejson.NewJson(content)
	if err != nil {
		return
	}

	for name, _ID := range json.MustMap() {
		ID := _ID.(string)
		apps = append(apps, App{Name: name, ID: ID})
	}

	return
}

// AddApp add new app into project's linked apps
func AddApp(projectPath string, name string, ID string) error {
	apps, err := GetApps(projectPath)
	if err != nil {
		return err
	}
	apps = append(apps, App{Name: name, ID: ID})

	err = os.Mkdir(appDirPath(projectPath), 0700)
	if err != nil && !os.IsExist(err) {
		return err
	}

	jsonApps := map[string]string{}
	for _, app := range apps {
		jsonApps[app.Name] = app.ID
	}
	data, err := json.Marshal(jsonApps)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(appFilePath(projectPath), data, 0700); err != nil {
		return err
	}

	return nil
}
