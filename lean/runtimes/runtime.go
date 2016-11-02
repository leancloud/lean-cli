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
	"github.com/fsnotify/fsnotify"
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
	command    *exec.Cmd
	Name       string
	Exec       string
	Args       []string
	WatchFiles []string
	Envs       []string
	Port       string
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

			fmt.Printf(" %s 项目已启动，请使用浏览器访问：http://localhost:%s\r\n", chrysanthemum.Success, runtime.Port)
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
					err := runtime.command.Process.Kill()
					if err != nil {
						runtime.Errors <- err
					}
					lastFiredTime = now
				}
			case err := <-watcher.Errors:
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
		Name:       "python",
		Exec:       execName,
		Args:       []string{"wsgi.py"},
		WatchFiles: []string{"."},
		Envs:       os.Environ(),
		Errors:     make(chan error),
	}, nil
}

func newNodeRuntime(projectPath string) (*Runtime, error) {
	execName := "node"
	script := "server.js"
	pkgFile := filepath.Join(projectPath, "package.json")
	if content, err := ioutil.ReadFile(pkgFile); err == nil {
		pkg := new(struct {
			Scripts struct {
				Start string `json:"start"`
			} `json:"scripts"`
		})
		err = json.Unmarshal(content, pkg)
		if err != nil {
			return nil, err
		}
		if pkg.Scripts.Start != "" {
			execName = "npm"
			script = "start"
		}
	}

	return &Runtime{
		Name:       "node.js",
		Exec:       execName,
		Args:       []string{script},
		WatchFiles: []string{"."},
		Envs:       os.Environ(),
		Errors:     make(chan error),
	}, nil
}

func newJavaRuntime(projectPath string) (*Runtime, error) {
	return &Runtime{
		Name:       "java",
		Exec:       "mvn",
		Args:       []string{"jetty:run"},
		WatchFiles: []string{"."},
		Envs:       os.Environ(),
		Errors:     make(chan error),
	}, nil
}

func newPhpRuntime(projectPath string) (*Runtime, error) {
	return &Runtime{
		Name:       "php",
		Exec:       "php",
		Args:       []string{"-S", "127.0.0.1:3000", "-t", "public"},
		WatchFiles: []string{"."},
		Envs:       os.Environ(),
		Errors:     make(chan error),
	}, nil
}
