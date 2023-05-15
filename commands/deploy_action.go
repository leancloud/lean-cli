package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aisk/logp"
	"github.com/aisk/wizard"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/runtimes"
	"github.com/leancloud/lean-cli/utils"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

func deployAction(c *cli.Context) error {
	version.PrintVersionAndEnvironment()
	isDeployFromGit := c.Bool("g")
	isDeployFromJavaWar := c.Bool("war")
	ignoreFilePath := c.String("leanignore")
	noDepsCache := c.Bool("no-cache")
	overwriteFuncs := c.Bool("overwrite-functions")
	message := c.String("message")
	keepFile := c.Bool("keep-deploy-file")
	revision := c.String("revision")
	prodBool := c.Bool("prod")
	staging := c.Bool("staging")
	isDirect := c.Bool("direct")
	directUpload := &isDirect
	if !c.IsSet("direct") {
		directUpload = nil
	}
	buildLogs := c.Bool("build-logs")

	var env string

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}

	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return err
	}

	if staging && prodBool {
		return cli.NewExitError("`--prod` and `--staging` flags are mutually exclusive", 1)
	}
	if staging {
		env = "0"
	} else if prodBool {
		env = "1"
	} else {
		logp.Info("`lean deploy` now has no default target. Specify the environment by `--prod` or `--staging` flag to avoid this prompt:")
		question := wizard.Question{
			Content: "Please select the environment: ",
			Answers: []wizard.Answer{
				{
					Content: "Production",
					Handler: func() {
						env = "1"
					},
				},
				{
					Content: "Staging",
					Handler: func() {
						env = "0"
					},
				},
			},
		}
		err = wizard.Ask([]wizard.Question{question})
		if err != nil {
			return err
		}
	}

	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return err
	}

	envText := "production"

	if env == "0" {
		envText = "staging"
	}

	logp.Info(fmt.Sprintf("Current app: %s (%s), group: %s, region: %s", color.GreenString(appInfo.AppName), appID, color.GreenString(groupName), region))
	logp.Info(fmt.Sprintf("Deploying new version to %s", color.GreenString(envText)))

	groupInfo, err := api.GetGroup(appID, groupName)
	if err != nil {
		return err
	}

	if env == "0" && !groupInfo.Staging.Deployable {
		return cli.NewExitError("Deployment failed: no staging instance", 1)
	} else if env == "1" && !groupInfo.Production.Deployable {
		return cli.NewExitError("Deployment failed: no production instance", 1)
	}

	opts := &api.DeployOptions{
		NoDepsCache:    noDepsCache,
		OverwriteFuncs: overwriteFuncs,
		BuildLogs:      buildLogs,
		Options:        c.String("options"),
	}

	if isDeployFromGit {
		err = deployFromGit(appID, groupName, env, revision, opts)
		if err != nil {
			return err
		}
	} else {
		opts.Message = getCommentMessage(message)
		err = deployFromLocal(appID, groupName, env, isDeployFromJavaWar, ignoreFilePath, keepFile, directUpload, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func packageProject(repoPath, ignoreFilePath string) (string, error) {
	fileDir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return "", err
	}

	archiveFile := filepath.Join(fileDir, "leanengine.zip")

	runtime, err := runtimes.DetectRuntime(repoPath)
	if err == runtimes.ErrRuntimeNotFound {
		logp.Warn("Failed to recognize project type. Please inspect the directory structure if the deployment failed.")
	} else if err != nil {
		return "", err
	}

	if err := runtime.ArchiveUploadFiles(archiveFile, ignoreFilePath); err != nil {
		return "", err
	}

	return archiveFile, nil
}

func packageWar(repoPath string) (string, error) {
	var warPath string
	files, err := ioutil.ReadDir(filepath.Join(repoPath, "target"))
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".war") && !file.IsDir() {
			warPath = filepath.Join(repoPath, "target", file.Name())
		}
	}
	if warPath == "" {
		return "", errors.New("cannot find .war file in ./target")
	}

	logp.Info("Found .war file:", warPath)

	file := []struct{ Name, Path string }{{
		Name: "ROOT.war",
		Path: warPath,
	}}

	for _, filename := range []string{"leanengine.yaml", "system.properties"} {
		path := filepath.Join(repoPath, filename)
		if utils.IsFileExists(path) {
			file = append(file, struct{ Name, Path string }{
				Name: filename,
				Path: path,
			})
		}
	}

	fileDir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return "", err
	}
	archivePath := filepath.Join(fileDir, "ROOT.war.zip")
	if err = utils.ArchiveFiles(archivePath, file); err != nil {
		return "", err
	}

	return archivePath, nil
}

func deployFromLocal(appID string, group string, env string, isDeployFromJavaWar bool, ignoreFilePath string, keepFile bool, directUpload *bool, opts *api.DeployOptions) error {
	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return err
	}

	var archiveFilePath string
	if isDeployFromJavaWar {
		archiveFilePath, err = packageWar(".")
	} else {
		archiveFilePath, err = packageProject(".", ignoreFilePath)
	}
	if directUpload != nil {
		opts.DirectUpload = *directUpload
	} else {
		if region != regions.USWest {
			opts.DirectUpload = false
		} else {
			fileInfo, err := os.Stat(archiveFilePath)
			if err != nil {
				return err
			}
			if fileInfo.Size() < 100*1024*1024 {
				opts.DirectUpload = true
			} else {
				opts.DirectUpload = false
			}
		}
	}
	var eventTok string
	if opts.DirectUpload {
		eventTok, err = api.DeployAppFromFile(appID, group, env, archiveFilePath, opts)
		if err != nil {
			return err
		}
	} else {
		file, err := api.UploadToRepoStorage(region, archiveFilePath)
		if err != nil {
			return err
		}
		eventTok, err = api.DeployAppFromFile(appID, group, env, file.URL, opts)
		if err != nil {
			return err
		}
		if !keepFile {
			defer func() {
				err := api.DeleteFromRepoStorage(region, file.ObjectID)
				if err != nil {
					logp.Error(err)
				}
			}()
		}
	}

	ok, err := api.PollEvents(appID, eventTok)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("Deployment failed", 1)
	}
	return nil
}

func deployFromGit(appID string, group string, env string, revision string, opts *api.DeployOptions) error {
	eventTok, err := api.DeployAppFromGit(appID, group, env, revision, opts)
	if err != nil {
		return err
	}
	ok, err := api.PollEvents(appID, eventTok)
	if err != nil {
		return err
	}
	if !ok {
		return cli.NewExitError("Deployment failed", 1)
	}
	return nil
}

func getCommentMessage(message string) string {
	if message == "" {
		_, err := exec.LookPath("git")

		if err == nil {
			if _, err := os.Stat("./.git"); !os.IsNotExist(err) {
				messageBuf, err := exec.Command("git", "log", "-1", "--no-color", "--pretty=%B").CombinedOutput()
				messageStr := string(messageBuf)

				if err != nil {
					logp.Error("failed to load git message: ", err)
				} else {
					message = "WIP on: " + strings.TrimSpace(messageStr)
				}
			}
		}
	}

	if message == "" {
		message = "Creating from the CLI"
	}

	return message
}
