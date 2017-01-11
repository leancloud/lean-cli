package runtimes

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aisk/chrysanthemum"
	"github.com/facebookgo/parseignore"
	"github.com/facebookgo/symwalk"
	"github.com/fsnotify/fsnotify"
	"github.com/jhoonb/archivex"
	"github.com/leancloud/lean-cli/lean/utils"
)

// ErrInvalidRuntime means the project's structure is not a valid LeanEngine project
var ErrInvalidRuntime = errors.New("invalid runtime")

type filesPattern struct {
	Includes []string
	Excludes []string
}

// Runtime stands for a language runtime
type Runtime struct {
	command     *exec.Cmd
	ProjectPath string
	Name        string
	Exec        string
	Args        []string
	WatchFiles  []string
	Envs        []string
	Port        string
	// DeployFiles is the patterns for source code to deploy to the remote server
	DeployFiles filesPattern
	// Errors is the channel that receives the command's error result
	Errors chan error
}

// Run the project, and watch file changes
func (runtime *Runtime) Run() {
	go func() {
		for {
			runtime.command = exec.Command(runtime.Exec, runtime.Args...)
			runtime.command.Env = os.Environ()
			runtime.command.Stdout = os.Stdout
			runtime.command.Stderr = os.Stderr
			runtime.command.Env = os.Environ()

			for _, env := range runtime.Envs {
				runtime.command.Env = append(runtime.command.Env, env)
			}

			chrysanthemum.Printf("项目已启动，请使用浏览器访问：http://localhost:%s\r\n", runtime.Port)
			err := runtime.command.Run()
			// TODO: this maybe not portable
			if err.Error() == "signal: killed" {
				continue
			} else {
				runtime.Errors <- err
				break
			}
		}
	}()
}

// Watch file changes
func (runtime *Runtime) Watch(interval time.Duration) error {

	// watch file changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	lastFiredTime := time.Now()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				_ = event
				now := time.Now()
				if now.Sub(lastFiredTime) > interval {
					err = runtime.command.Process.Kill()
					if err != nil {
						runtime.Errors <- err
					}
					lastFiredTime = now
				}
			case err = <-watcher.Errors:
				runtime.Errors <- err
			}
		}
	}()
	for _, file := range runtime.WatchFiles {
		err = watcher.Add(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (runtime *Runtime) ArchiveUploadFiles(archiveFile string, isDeployFromJavaWar bool, ignoreFilePath string) error {
	if runtime.Name == "java" && isDeployFromJavaWar {
		warFile, err := getDefaultWarFile(runtime.ProjectPath)
		if err != nil {
			return err
		}
		spinner := chrysanthemum.New("压缩 war 文件:" + warFile).Start()
		err = Archive(archiveFile, warFile, "ROOT.war")
		if err != nil {
			spinner.Failed()
			return err
		}
		spinner.Successed()
	} else {
		err := runtime.defaultArchive(archiveFile, ignoreFilePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func getDefaultWarFile(projectPath string) (string, error) {
	files, err := ioutil.ReadDir(filepath.Join(projectPath, "target"))
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".war") && !file.IsDir() {
			return filepath.Join(projectPath, "target", file.Name()), nil
		}
	}
	return "", errors.New("在 ./target 目录没有找到 war 文件。")
}

// Archive will archive a file to .zip package
func Archive(archiveFile string, file string, name string) error {
	targetFile, err := os.Create(archiveFile)
	if err != nil {
		return err
	}
	writer := zip.NewWriter(targetFile)
	defer writer.Close()
	zippedFile, err := writer.Create(name)
	if err != nil {
		return err
	}
	fromFile, err := os.Open(file)
	if err != nil {
		return err
	}
	fileReader := bufio.NewReader(fromFile)
	blockSize := 512 * 1024 // 512kb
	bytes := make([]byte, blockSize)
	for {
		readedBytes, err := fileReader.Read(bytes)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			if err.Error() != "EOF" {
				return err
			}
		}
		if readedBytes >= blockSize {
			zippedFile.Write(bytes)
			continue
		}
		zippedFile.Write(bytes[:readedBytes])
	}
	return nil
}

func (runtime *Runtime) defaultArchive(archiveFile string, ignoreFilePath string) error {
	matcher, err := runtime.readIgnore(ignoreFilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("指定的 ignore 文件 '%s' 不存在", ignoreFilePath)
	} else if err != nil {
		return err
	}

	files := []string{}
	err = symwalk.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// convert DOS's '\' path seprater to UNIX style
		path = filepath.ToSlash(path)
		decision, err := matcher.Match(path, info)
		if err != nil {
			return err
		}

		if info.IsDir() {
			if decision == parseignore.Exclude {
				return filepath.SkipDir
			}
			return nil
		}

		if decision != parseignore.Exclude {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return err
	}
	spinner := chrysanthemum.New("压缩项目文件").Start()
	zip := new(archivex.ZipFile)
	func() {
		defer zip.Close()
		zip.Create(archiveFile)
		for _, f := range files {
			err := zip.AddFile(filepath.ToSlash(f))
			if err != nil {
				panic(err)
			}
		}
	}()
	spinner.Successed()
	return nil
}

// DetectRuntime returns the project's runtime
func DetectRuntime(projectPath string) (*Runtime, error) {
	bar := chrysanthemum.New("正在检测运行时").Start()
	// order is important
	if utils.IsFileExists(filepath.Join(projectPath, "cloud", "main.js")) {
		chrysanthemum.Printf("检测到 cloudcode 运行时\r\n")
		bar.Successed()
		return &Runtime{
			Name: "cloudcode",
		}, nil
	}
	if utils.IsFileExists(filepath.Join(projectPath, "server.js")) && utils.IsFileExists(filepath.Join(projectPath, "package.json")) {
		bar.Successed()
		chrysanthemum.Printf("检测到 node.js 运行时\r\n")
		return newNodeRuntime(projectPath)
	}
	if utils.IsFileExists(filepath.Join(projectPath, "package.json")) {
		data, err := ioutil.ReadFile(filepath.Join(projectPath, "package.json"))
		if err == nil {
			var result struct {
				Scripts struct {
					Start string `json:"start"`
				} `json:"scripts"`
			}
			if err = json.Unmarshal(data, &result); err == nil {
				if result.Scripts.Start != "" {
					bar.Successed()
					chrysanthemum.Printf("检测到 node.js 运行时\r\n")
					return newNodeRuntime(projectPath)
				}
			}
		}
	}
	if utils.IsFileExists(filepath.Join(projectPath, "requirements.txt")) && utils.IsFileExists(filepath.Join(projectPath, "wsgi.py")) {
		bar.Successed()
		chrysanthemum.Printf("检测到 Python 运行时\r\n")
		return newPythonRuntime(projectPath)
	}
	if utils.IsFileExists(filepath.Join(projectPath, "composer.json")) && utils.IsFileExists(filepath.Join("public", "index.php")) {
		bar.Successed()
		chrysanthemum.Printf("检测到 PHP 运行时\r\n")
		return newPhpRuntime(projectPath)
	}
	if utils.IsFileExists(filepath.Join(projectPath, "pom.xml")) {
		bar.Successed()
		chrysanthemum.Printf("检测到 Java 运行时\r\n")
		return newJavaRuntime(projectPath)
	}
	return nil, ErrInvalidRuntime
}

func lookupBin(fallbacks []string) (string, error) {
	for i, bin := range fallbacks {
		binPath, err := exec.LookPath(bin)
		if err == nil { // found
			if i == 0 {
				chrysanthemum.Printf("找到运行文件 `%s`\r\n", binPath)
			} else {
				chrysanthemum.Printf("没有找到命令 `%s`，使用 `%s` 代替 \r\n", fallbacks[i-1], fallbacks[i])
			}
			return bin, nil
		}
	}

	return "", fmt.Errorf("`%s` not found", fallbacks[0])
}

func newPythonRuntime(projectPath string) (*Runtime, error) {
	execName := "python2.7"

	if content, err := ioutil.ReadFile(filepath.Join(projectPath, "runtime.txt")); err == nil {
		if strings.HasPrefix(string(content), "python-2.7") {
			execName, err = lookupBin([]string{"python2.7", "python2", "python"})
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(string(content), "python-3.5") {
			execName, err = lookupBin([]string{"python3.5", "python3", "python"})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("invalid python runtime.txt format, only `python-2.7` and `python-3.5` were allowed")
		}
	}

	return &Runtime{
		ProjectPath: projectPath,
		Name:        "python",
		Exec:        execName,
		Args:        []string{"wsgi.py"},
		WatchFiles:  []string{"."},
		Envs:        os.Environ(),
		Errors:      make(chan error),
	}, nil
}

func newNodeRuntime(projectPath string) (*Runtime, error) {
	execName := "node"
	args := []string{"server.js"}
	pkgFile := filepath.Join(projectPath, "package.json")
	if content, err := ioutil.ReadFile(pkgFile); err == nil {
		pkg := new(struct {
			Scripts struct {
				Start string `json:"start"`
				Dev   string `json:"dev"`
			} `json:"scripts"`
		})
		err = json.Unmarshal(content, pkg)
		if err != nil {
			return nil, err
		}

		if pkg.Scripts.Dev != "" {
			execName = "npm"
			args = []string{"run", "dev"}
		} else if pkg.Scripts.Start != "" {
			execName = "npm"
			args = []string{"start"}
		}
	}

	return &Runtime{
		ProjectPath: projectPath,
		Name:        "node.js",
		Exec:        execName,
		Args:        args,
		WatchFiles:  []string{"."},
		Envs:        os.Environ(),
		Errors:      make(chan error),
	}, nil
}

func newJavaRuntime(projectPath string) (*Runtime, error) {
	return &Runtime{
		ProjectPath: projectPath,
		Name:        "java",
		Exec:        "mvn",
		Args:        []string{"jetty:run"},
		WatchFiles:  []string{"."},
		Envs:        os.Environ(),
		Errors:      make(chan error),
	}, nil
}

func newPhpRuntime(projectPath string) (*Runtime, error) {
	return &Runtime{
		ProjectPath: projectPath,
		Name:        "php",
		Exec:        "php",
		Args:        []string{"-S", "127.0.0.1:3000", "-t", "public"},
		WatchFiles:  []string{"."},
		Envs:        os.Environ(),
		Errors:      make(chan error),
	}, nil
}
