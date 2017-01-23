package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aisk/chrysanthemum"
	"github.com/codegangsta/cli"
	"github.com/leancloud/lean-cli/api"
	"github.com/leancloud/lean-cli/apps"
)

func uploadFile(appID string, filePath string) error {
	chrysanthemum.Println("上传文件:", filePath)
	file, err := api.UploadFile(appID, filePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf(" %s 上传成功，文件 URL：%s\r\n", chrysanthemum.Success, file.URL)
	return nil
}

func uploadAction(c *cli.Context) error {
	if c.NArg() < 1 {
		cli.ShowCommandHelp(c, "upload")
		return cli.NewExitError("", 1)
	}

	appID, err := apps.GetCurrentAppID(".")
	if err != nil {
		return newCliError(err)
	}

	for _, filePath := range c.Args() {
		f, err := os.Open(filePath)
		if err != nil {
			return newCliError(err)
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			return newCliError(err)
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
				return newCliError(err)
			}
		} else {
			err := uploadFile(appID, filePath)
			if err != nil {
				return newCliError(err)
			}
		}
	}

	return nil
}
