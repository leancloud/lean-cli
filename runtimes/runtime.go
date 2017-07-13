package runtimes

import (
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
	"github.com/leancloud/lean-cli/utils"
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
	Remote      string
	Port        string
	// DeployFiles is the patterns for source code to deploy to the remote server
	DeployFiles filesPattern
	// Errors is the channel that receives the command's error result
	Errors chan error
}

// Run the project, and watch file changes
func (runtime *Runtime) Run() {
	go func() {
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
		if err != nil {
			runtime.Errors <- err
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

func (runtime *Runtime) ArchiveUploadFiles(archiveFile string, ignoreFilePath string) error {
	return runtime.defaultArchive(archiveFile, ignoreFilePath)
}

func (runtime *Runtime) defaultArchive(archiveFile string, ignoreFilePath string) error {
	matcher, err := runtime.readIgnore(ignoreFilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("指定的 ignore 文件 '%s' 不存在", ignoreFilePath)
	} else if err != nil {
		return err
	}

	files := []struct{ Name, Path string }{}
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
			files = append(files, struct{ Name, Path string }{
				Name: path,
				Path: path,
			})
		}
		return nil
	})

	if err != nil {
		return err
	}
	spinner := chrysanthemum.New("压缩项目文件").Start()

	err = utils.ArchiveFiles(archiveFile, files)
	if err != nil {
		spinner.Failed()
	}
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
	packageFilePath := filepath.Join(projectPath, "package.json")
	if utils.IsFileExists(filepath.Join(projectPath, "server.js")) && utils.IsFileExists(packageFilePath) {
		bar.Successed()
		chrysanthemum.Printf("检测到 node.js 运行时\r\n")
		return newNodeRuntime(projectPath)
	}
	if utils.IsFileExists(packageFilePath) {
		data, err := ioutil.ReadFile(packageFilePath)
		if err == nil {
			data = utils.StripUTF8BOM(data)
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
	bar.Failed()
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

	content, err := ioutil.ReadFile(filepath.Join(projectPath, ".python-version"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		execName := "python2.7"
		content, err = ioutil.ReadFile(filepath.Join(projectPath, "runtime.txt"))
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			// the default content
			content = []byte("python-2.7")
		}
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
	pythonVersion := string(content)
	if !(strings.HasPrefix(pythonVersion, "2.") || strings.HasPrefix(pythonVersion, "3.")) {
		return nil, errors.New("错误的 pyenv 版本，目前云引擎只支持 CPython，请检查 .python-version 文件确认")
	}
	chrysanthemum.Println("检测到项目使用 pyenv，请确保当前环境 pyenv 已正确设置")

	return &Runtime{
		ProjectPath: projectPath,
		Name:        "python",
		Exec:        "python",
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
		content = utils.StripUTF8BOM(content)
		pkg := new(struct {
			Scripts struct {
				Start string `json:"start"`
				Dev   string `json:"dev"`
			} `json:"scripts"`
			Dependencies map[string]string `json:"dependencies"`
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

		if sdkVersion, ok := pkg.Dependencies["leanengine"]; ok {
			if strings.HasPrefix(sdkVersion, "0.") ||
				strings.HasPrefix(sdkVersion, "~0.") ||
				strings.HasPrefix(sdkVersion, "^0.") {
				s := "当前使用 leanengine SDK 版本过低，本地云函数调试功能将会不能正常启用。建议参考 http://url.leanapp.cn/Og1cVia 尽快升级。"
				fmt.Fprintf(os.Stderr, " %s [WARNING] %s\r\n", chrysanthemum.Fail, s)
			}
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
	entryScript, err := getPHPEntryScriptPath()
	if err != nil {
		return nil, err
	}
	return &Runtime{
		ProjectPath: projectPath,
		Name:        "php",
		Exec:        "php",
		Args:        []string{"-S", "127.0.0.1:3000", "-t", "public", entryScript},
		WatchFiles:  []string{"."},
		Envs:        os.Environ(),
		Errors:      make(chan error),
	}, nil
}
