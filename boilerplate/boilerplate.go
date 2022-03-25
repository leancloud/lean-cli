package boilerplate

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aisk/logp"
	"github.com/cheggaaa/pb"
	"github.com/fatih/color"
	"github.com/leancloud/lean-cli/api/regions"
	"github.com/leancloud/lean-cli/utils"
	"github.com/leancloud/lean-cli/version"
	"github.com/levigross/grequests"
	"github.com/mattn/go-colorable"
)

func CreateProject(boil *Boilerplate, dest string, appID string, region regions.Region) error {
	if boil.DownloadURL != "" {
		if err := os.Mkdir(dest, 0775); err != nil {
			return err
		}

		dir, err := ioutil.TempDir("", "leanengine")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		zipFilePath := filepath.Join(dir, "getting-started.zip")

		var downloadURLs []string
		if region.InChina() {
			downloadURLs = []string{"https://releases.leanapp.cn", "https://api.github.com/repos"}
		} else {
			downloadURLs = []string{"https://api.github.com/repos", "https://releases.leanapp.cn"}
		}
		err = downloadToFile(downloadURLs[0]+boil.DownloadURL, zipFilePath)
		if err != nil {
			err = downloadToFile(downloadURLs[1]+boil.DownloadURL, zipFilePath)
			if err != nil {
				return err
			}
		}

		logp.Info("Creating project...")

		zipFile, err := zip.OpenReader(zipFilePath)
		if err != nil {
			return err
		}
		defer zipFile.Close()
		for _, f := range zipFile.File {
			// Remove outer directory name.
			f.Name = f.Name[strings.Index(f.Name, "/"):]
			err := extractAndWriteFile(f, dest)
			if err != nil {
				return err
			}
		}
	}

	if boil.CMD != nil {
		args := boil.CMD(dest)
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		stdin, err := cmd.StdinPipe()

		if err != nil {
			return err
		}

		go func() {
			defer stdin.Close()
			io.Copy(stdin, os.Stdin)
		}()

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if boil.Files != nil {
		for name, body := range boil.Files {
			if err := ioutil.WriteFile(filepath.Join(dest, name), []byte(body), 0644); err != nil {
				return err
			}
		}
	}

	logp.Info(fmt.Printf("Created %s project in `%s`", boil.Name, dest))

	if boil.Message != "" {
		logp.Info(boil.Message)
	}

	return nil
}

type Boilerplate struct {
	Name        string
	Message     string
	DownloadURL string
	CMD         func(dest string) []string
	Files       map[string]string
}

var Boilerplates = []Boilerplate{
	{
		Name:        "Node.js - Express",
		Message:     "Lean how to use Express at https://expressjs.com",
		DownloadURL: "/leancloud/node-js-getting-started/zipball/master",
	},
	{
		Name:        "Node.js - Koa",
		DownloadURL: "/leancloud/koa-getting-started/zipball/master",
		Message:     "Lean how to use Koa at https://koajs.com",
	},
	{
		Name:        "Python - Flask",
		DownloadURL: "/leancloud/python-getting-started/zipball/master",
		Message:     "Lean how to use Flask at https://flask.palletsprojects.com",
	},
	{
		Name:        "Python - Django",
		DownloadURL: "/leancloud/django-getting-started/zipball/master",
		Message:     "Lean how to use Django at https://docs.djangoproject.com",
	},
	{
		Name:        "Java - Serlvet",
		DownloadURL: "/leancloud/servlet-getting-started/zipball/master",
	},
	{
		Name:        "Java - Spring Boot",
		DownloadURL: "/leancloud/spring-boot-getting-started/zipball/master",
		Message:     "Lean how to use Spring Boot at https://spring.io/projects/spring-boot",
	},
	{
		Name:        "PHP - Slim",
		DownloadURL: "/leancloud/slim-getting-started/zipball/master",
		Message:     "Lean how to use Slim at https://www.slimframework.com",
	},
	{
		Name:        ".NET Core",
		DownloadURL: "/leancloud/dotnet-core-getting-started/zipball/master",
		Message:     "Lean how to use .NET Core at https://docs.microsoft.com/aspnet/core/",
	},
	{
		Name:        "Go - Echo",
		DownloadURL: "/leancloud/golang-getting-started/zipball/master",
		Message:     "Lean how to use Echo at https://echo.labstack.com/",
	},
	{
		Name:  "Web App - React (via create-react-app, require NPM installed)",
		Files: prepareWebAppFiles("build"),
		CMD: func(dest string) []string {
			return []string{"npx", "create-react-app", dest, "--use-npm"}
		},
	},
	{
		Name:  "Web App - Vue (via @vue/cli, require NPM installed)",
		Files: prepareWebAppFiles("dist"),
		CMD: func(dest string) []string {
			return []string{"npx", "@vue/cli", "create", "--default", "--packageManager", "npm", dest}
		},
	},
}

// don't know why archive/zip.Reader.File[0].FileInfo().IsDir() always return true,
// this is a trick hack to void this.
func isDir(path string) bool {
	return os.IsPathSeparator(path[len(path)-1])
}

func extractAndWriteFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(dest, f.Name)

	if isDir(f.Name) {
		if err := os.MkdirAll(path, f.Mode()); err != nil {
			return err
		}
	} else {
		// Use os.Create() since Zip don't store file permissions.
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(f, rc)
		if err != nil {
			return err
		}
	}
	return nil
}

// downloadToFile allows you to download the contents of the URL to a file
func downloadToFile(url string, fileName string) error {
	resp, err := grequests.Get(url, &grequests.RequestOptions{
		UserAgent: "LeanCloud-CLI/" + version.Version,
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.StatusCode != 200 {
		return errors.New(utils.FormatServerErrorResult(resp.String()))
	}
	if resp.Error != nil {
		return resp.Error
	}

	fd, err := os.Create(fileName)

	if err != nil {
		return err
	}

	defer resp.Close() // This is a noop if we use the internal ByteBuffer
	defer fd.Close()

	if length, err := strconv.Atoi(resp.Header.Get("Content-Length")); err == nil {
		bar := pb.New(length).SetUnits(pb.U_BYTES).SetMaxWidth(80)
		bar.Output = colorable.NewColorableStderr()
		bar.Prefix(color.GreenString("[INFO]") + " Downloading templates")
		bar.Start()
		defer bar.Finish()
		reader := bar.NewProxyReader(resp)
		if _, err := io.Copy(fd, reader); err != nil && err != io.EOF {
			return err
		}
	} else {
		if _, err := io.Copy(fd, resp); err != nil && err != io.EOF {
			return err
		}
	}

	return nil
}

func prepareWebAppFiles(webRoot string) map[string]string {
	return map[string]string{
		"leanengine.yaml": "build: npm run build",
		"static.json": fmt.Sprintf(`{
  "public": "%s",
  "rewrites": [
    { "source": "**", "destination": "/index.html" }
  ]
}`, webRoot),
	}
}
