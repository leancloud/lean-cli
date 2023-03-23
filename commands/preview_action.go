package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/aisk/logp"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli"
)

func getEnvInfo(c *cli.Context) (name, commit, url string, err error) {
	var pr, ciCommit, ciUrl string
	if os.Getenv("GITLAB_CI") == "true" {
		pr = os.Getenv("CI_MERGE_REQUEST_IID")
		ciCommit = os.Getenv("CI_COMMIT_SHA")
		projectUrl := os.Getenv("CI_PROJECT_URL")
		ciUrl = fmt.Sprintf("%s/merge_request/%s", projectUrl, pr)
	} else if os.Getenv("GITHUB_ACTIONS") == "true" {
		// e.g. "refs/pull/123/merge"
		ref := os.Getenv("GITHUB_REF")
		pr = strings.Split(ref, "/")[2]
		ciCommit = os.Getenv("GITHUB_SHA")
		repo := os.Getenv("GITHUB_REPOSITORY")
		ciUrl = fmt.Sprintf("https://github.com/%s/pull/%s", repo, pr)
	}

	commit = c.String("commit")
	if commit == "" {
		commit = ciCommit
	}
	url = c.String("url")
	if url == "" {
		url = ciUrl
	}
	name = c.String("name")
	if name == "" {
		if pr == "" {
			err = cli.NewExitError("Not running in GitLab CI / GitHub Actions. Please set `--name`", 1)
			return
		}
		name = fmt.Sprintf("pr-%s", pr)
	} else if name == "1" || name == "0" {
		err = cli.NewExitError("Preview environment name can't be 1 or 0", 1)
		return
	}
	return
}

func deployPreviewAction(c *cli.Context) error {
	name, commit, url, err := getEnvInfo(c)
	if err != nil {
		return err
	}
	buildLogs := c.Bool("build-logs")
	isDeployFromGit := c.Bool("g")
	noDepsCache := c.Bool("no-cache")
	isDeployFromJavaWar := c.Bool("war")
	ignoreFilePath := c.String("leanignore")

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	appInfo, err := api.GetAppInfo(appID)
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

	logp.Info(fmt.Sprintf("Current app: %s (%s), group: %s, region: %s", color.GreenString(appInfo.AppName), appID, color.GreenString(groupName), region))
	logp.Info(fmt.Sprintf("Deploying %s to preview environment %s", color.GreenString(commit), color.GreenString(name)))

	opts := &api.DeployOptions{
		Commit:      commit,
		Url:         url,
		NoDepsCache: noDepsCache,
		BuildLogs:   buildLogs,
	}

	if isDeployFromGit {
		err = deployFromGit(appID, groupName, name, commit, opts)
		if err != nil {
			return err
		}
	} else {
		err = deployFromLocal(appID, groupName, name, isDeployFromJavaWar, ignoreFilePath, false, opts)
		if err != nil {
			return err
		}
	}

	domainBindings, err := api.GetDomainBindings(appID, api.EnginePreview, groupName)
	if err != nil {
		return err
	}
	if len(domainBindings) == 0 {
		logp.Warn("There are no preview domains associated with this group. Please bind one first.")
	} else {
		getUrl := func(domain api.DomainBinding) string {
			proto := "http"
			if domain.SslType != "none" {
				proto = "https"
			}
			return fmt.Sprintf("%s://%s.%s", proto, name, strings.TrimPrefix(domain.Domain, "*."))
		}
		for _, domain := range domainBindings {
			logp.Info("Preview URL:", color.GreenString(getUrl(domain)))
		}
		// Print preview URL to *stdout* when used as URL=$(lean preview deploy ...)
		if !isatty.IsTerminal(1) {
			fmt.Println(getUrl(domainBindings[0]))
		}
	}

	return nil
}

func deletePreviewAction(c *cli.Context) error {
	name, _, _, err := getEnvInfo(c)
	if err != nil {
		return err
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	groupName, err := apps.GetCurrentGroup(".")
	if err != nil {
		return err
	}

	err = api.DeleteEnvironment(appID, groupName, name)
	if err != nil {
		return err
	}

	logp.Infof("Deleted preview environment %s", name)
	return nil
}
