package boilerplate

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
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

// FetchRepo will download the boilerplate from remote and extract to ${dest}/folder
func FetchRepo(boil *Boilerplate, dest string, appID string, region regions.Region) error {
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
	switch region {
	case regions.ChinaNorth, regions.ChinaEast, regions.ChinaTDS1:
		downloadURLs = []string{"https://releases.leanapp.cn", "https://api.github.com/repos"}
	case regions.USWest, regions.APSG:
		downloadURLs = []string{"https://api.github.com/repos", "https://releases.leanapp.cn"}
	}
	err = DownloadToFile(downloadURLs[0]+boil.URL, zipFilePath)
	if err != nil {
		err = DownloadToFile(downloadURLs[1]+boil.URL, zipFilePath)
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
	// TODO: Change value of boil.Homepage for English site.
	logp.Info("Creating", boil.Name, "succeeded. Please refer to the website", boil.Homepage, "for documentation")
	return nil
}

type Boilerplate struct {
	Name     string `json:"name"`
	Homepage string `json:"homepage"`
	URL      string `json:"url"`
}

// DownloadToFile allows you to download the contents of the URL to a file
func DownloadToFile(url string, fileName string) error {
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

var Boilerplates = []Boilerplate{
	{
		Name:     "Express - Node.js",
		URL:      "/leancloud/node-js-getting-started/zipball/latest",
		Homepage: "http://expressjs.com/",
	},
	{
		Name:     "Koa - Node.js",
		URL:      "/leancloud/koa-getting-started/zipball/latest",
		Homepage: "http://koajs.com/",
	},
	{
		Name:     "Flask - Python",
		URL:      "/leancloud/python-getting-started/zipball/latest",
		Homepage: "http://flask.pocoo.org/",
	},
	{
		Name:     "Django - Python",
		URL:      "/leancloud/django-getting-started/zipball/latest",
		Homepage: "https://www.djangoproject.com/",
	},
	{
		Name:     "Serlvet - Java",
		URL:      "/leancloud/java-war-getting-started/zipball/latest",
		Homepage: "https://jcp.org/en/jsr/detail?id=340",
	},
	{
		Name:     "Spring Boot - Java",
		URL:      "/leancloud/spring-boot-getting-started/zipball/latest",
		Homepage: "https://spring.io/projects/spring-boot",
	},
	{
		Name:     "Slim - PHP",
		URL:      "/leancloud/slim-getting-started/zipball/latest",
		Homepage: "http://www.slimframework.com/",
	},
	{
		Name:     ".NET Core - .Net",
		URL:      "/leancloud/dotnet-core-getting-started/zipball/latest",
		Homepage: "https://dotnet.microsoft.com/",
	},
	{
		Name:     "Echo - Go",
		URL:      "/leancloud/golang-getting-started/zipball/latest",
		Homepage: "https://echo.labstack.com/",
	},
	{
		Name:     "Static Site",
		URL:      "/leancloud/static-getting-started/zipball/latest",
		Homepage: "https://github.com/cloudhead/node-static",
	},
}
