package commands

import (
	"os"
	"path/filepath"

	"github.com/aisk/logp"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
	"github.com/urfave/cli"
)

func uploadFile(appID string, filePath string) error {
	logp.Info("上传文件: " + filePath)
	file, err := api.UploadFile(appID, filePath)
	if err != nil {
		return err
	}
	logp.Infof("上传成功，文件 URL：%s\r\n", file.URL)
	return nil
}

func uploadAction(c *cli.Context) error {
	if c.NArg() < 1 {
		cli.ShowCommandHelp(c, "upload")
		return cli.NewExitError("", 1)
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return err
	}

	for _, filePath := range c.Args() {
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			return err
		}
		if stat.IsDir() {
			err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				return uploadFile(appID, path)
			})
			if err != nil {
				return err
			}
		} else {
			err := uploadFile(appID, filePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
