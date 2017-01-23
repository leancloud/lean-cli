package apps

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/leancloud/lean-cli/api"
)

func recentLinkedApps(projectPath string) string {
	return filepath.Join(appDirPath(projectPath), "recent_linked_apps")
}

func setRecentLinkedApp(projectPath string, appID string) error {
	content, err := ioutil.ReadFile(recentLinkedApps(projectPath))
	if os.IsNotExist(err) {
		content = []byte("[]")
	} else if err != nil {
		return err
	}

	var linkedApps []string
	err = json.Unmarshal(content, &linkedApps)
	if err != nil {
		return err
	}

	// the codes below use `goto` statement just for python's 'for ... else ...' function.
	// if you want add more logic codes here, please refactor this in order to remove `goto`

	if len(linkedApps) == 0 {
		// 1, no app linked previous
		linkedApps = append(linkedApps, appID)
		goto SaveLinkedApps
	}

	for i, app := range linkedApps {
		if app == appID {
			// 2, save app linked again, move app to list header
			linkedApps = append(linkedApps[:i], linkedApps[1+i:]...)
			linkedApps = append([]string{app}, linkedApps...)
			goto SaveLinkedApps
		}
	}

	// 3, some apps linked before, excluded this one
	linkedApps = append([]string{appID}, linkedApps...)

	if len(linkedApps) > 5 {
		linkedApps = linkedApps[0:5]
	}

	goto SaveLinkedApps

SaveLinkedApps:
	content, err = json.Marshal(linkedApps)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(recentLinkedApps(projectPath), content, 0644)
}

func getRecentLinkedApps(projectPath string) ([]string, error) {
	content, err := ioutil.ReadFile(recentLinkedApps(projectPath))
	if os.IsNotExist(err) {
		return []string{}, nil
	}

	var linkedApps []string
	err = json.Unmarshal(content, &linkedApps)
	return linkedApps, err
}

// MergeWithRecentApps will adjust the appIDs' order with recent linked apps
func MergeWithRecentApps(projectPath string, apps []*api.GetAppListResult) ([]*api.GetAppListResult, error) {
	linkedAppIDs, err := getRecentLinkedApps(projectPath)
	if err != nil {
		return apps, err
	}

	if len(linkedAppIDs) == 0 {
		return apps, nil
	}

	var appIDsToAppend []string

	for _, linkedAppID := range linkedAppIDs {
		for _, app := range apps {
			if app.AppID == linkedAppID {
				appIDsToAppend = append(appIDsToAppend, app.AppID)
			}
		}
	}

	for i := range appIDsToAppend {
		appIDToAppend := appIDsToAppend[len(appIDsToAppend)-i-1]
		for j, app := range apps {
			if app.AppID == appIDToAppend {
				apps = append(apps[:j], apps[j+1:]...)
				apps = append([]*api.GetAppListResult{app}, apps...)
			}
		}
	}

	return apps, nil
}
