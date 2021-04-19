package commands

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aisk/logp"
	"github.com/leancloud/go-upload"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/apps"
	"github.com/leancloud/lean-cli/runtimes"
	"github.com/leancloud/lean-cli/utils"
	"github.com/leancloud/lean-cli/version"
	"github.com/urfave/cli"
)

func deployAction(c *cli.Context) error {
	version.PrintCurrentVersion()
	isDeployFromGit := c.Bool("g")
	isDeployFromJavaWar := c.Bool("war")
	ignoreFilePath := c.String("leanignore")
	noDepsCache := c.Bool("no-cache")
	overwriteFuncs := c.Bool("overwrite-functions")
	message := c.String("message")
	keepFile := c.Bool("keep-deploy-file")
	revision := c.String("revision")
	prodString := c.String("prod")
	isDirect := c.Bool("direct")

	var prod int

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}

	logp.Info("Retrieving app info ...")

	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return err
	}

	appInfo, err := api.GetAppInfo(appID)
	if err != nil {
		return err
	}

	if prodString == "" {
		groupInfo, err := api.GetGroup(appID, groupName)
		if err != nil {
			return err
		}

		if groupInfo.Staging.Deployable {
			prod = 0
		} else {
			prod = 1
		}
	} else {
		prod, err = strconv.Atoi(prodString)
		if err != nil {
			return err
		}
	}

	if prod == 1 {
		logp.Infof("Preparing to deploy %s(%s) to region: %s group: %s production\r\n", appInfo.AppName, appID, region, groupName)
	} else {
		logp.Infof("Preparing to deploy %s(%s) to region: %s group: %s staging\r\n", appInfo.AppName, appID, region, groupName)
	}

	opts := &api.DeployOptions{
		NoDepsCache:    noDepsCache,
		OverwriteFuncs: overwriteFuncs,
		Options:        c.String("options"),
	}

	if isDeployFromGit {
		err = deployFromGit(appID, groupName, prod, revision, opts)
		if err != nil {
			return err
		}
	} else {
		opts.Message = getCommentMessage(message)
		opts.DirectUpload = isDirect
		err = deployFromLocal(appID, groupName, prod, isDeployFromJavaWar, ignoreFilePath, keepFile, opts)
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

func uploadProject(appID string, region regions.Region, repoPath string, ignoreFilePath string) (*upload.File, error) {
	archiveFile, err := packageProject(repoPath, ignoreFilePath)
	if err != nil {
		return nil, err
	}

	file, err := api.UploadToRepoStorage(region, archiveFile)
	if err != nil {
		return nil, err
	}

	return file, nil
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

	fileDir, err := ioutil.TempDir("", "leanengine")
	if err != nil {
		return "", err
	}
	archivePath := filepath.Join(fileDir, "ROOT.war.zip")

	file := []struct{ Name, Path string }{{
		Name: "ROOT.war",
		Path: warPath,
	}}
	if err = utils.ArchiveFiles(archivePath, file); err != nil {
		return "", err
	}

	return archivePath, nil
}

func uploadWar(appID string, region regions.Region, repoPath string) (*upload.File, error) {
	archivePath, err := packageWar(repoPath)
	if err != nil {
		return nil, err
	}
	return api.UploadToRepoStorage(region, archivePath)
}

func deployFromLocal(appID string, group string, prod int, isDeployFromJavaWar bool, ignoreFilePath string, keepFile bool, opts *api.DeployOptions) error {
	region, err := apps.GetAppRegion(appID)
	if err != nil {
		return err
	}

	var file *upload.File
	var archiveFilePath string
	if isDeployFromJavaWar {
		if opts.DirectUpload {
			archiveFilePath, err = packageWar(".")
		} else {
			file, err = uploadWar(appID, region, ".")
		}
		if err != nil {
			return err
		}
	} else {
		if opts.DirectUpload {
			archiveFilePath, err = packageProject(".", ignoreFilePath)
		} else {
			file, err = uploadProject(appID, region, ".", ignoreFilePath)
		}
		if err != nil {
			return err
		}
	}

	if !keepFile {
		defer func() {
			logp.Info("Deleting temporary files")
			err := api.DeleteFromRepoStorage(region, file.ObjectID)
			if err != nil {
				logp.Error(err)
			}
		}()
	}

	var eventTok string
	if opts.DirectUpload {
		eventTok, err = api.DeployAppFromFile(appID, group, prod, file.URL, opts)
	} else {
		eventTok, err = api.DeployAppFromFile(appID, group, prod, archiveFilePath, opts)
	}
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

func deployFromGit(appID string, group string, prod int, revision string, opts *api.DeployOptions) error {
	eventTok, err := api.DeployAppFromGit(appID, group, prod, revision, opts)
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
