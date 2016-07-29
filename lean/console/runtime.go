package console

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/leancloud/lean-cli/lean/utils"
)

type filesPattern struct {
	Includes []string
	Excludes []string
}

// Runtime stands for a language runtime
type Runtime struct {
	Name       string
	Exec       string
	Args       []string
	WatchFiles []string
	Envs       []string
	// DeployFiles is the patterns for source code to deploy to the remote server
	DeployFiles filesPattern
}

// Run the project, and watch file changes
func (runtime *Runtime) Run() error {
	command := exec.Command(runtime.Exec, runtime.Args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	for _, env := range runtime.Envs {
		command.Env = append(command.Env, env)
	}

	return command.Run()
	// if err != nil {
	// 	return err
	// }

	// // watch file changes
	// watcher, err := fsnotify.NewWatcher()
	// if err != nil {
	// 	return err
	// }
	// defer watcher.Close()
	// go func() {
	// 	for {
	// 		select {
	// 		case event := <-watcher.Events:
	// 			fmt.Println("event:", event)
	// 			command.Process.Kill()
	// 		case err := <-watcher.Errors:
	// 			fmt.Println("error:", err)
	// 		}
	// 	}
	// }()
	// for _, file := range runtime.WatchFiles {
	// 	err = watcher.Add(file)
	// }

	// // start the command line
	// return command.Wait()
}

// DetectRuntime returns the project's runtime
func DetectRuntime(projectPath string) (*Runtime, error) {
	// order is importand
	if utils.IsFileExists(filepath.Join("cloud", "main.js")) {
		log.Println("cloudcode runtime detected.")
		return nil, nil
	}
	if utils.IsFileExists("server.js") && utils.IsFileExists("package.json") {
		log.Println("node.js runtime detected.")
		return newNodeRuntime(projectPath)
	}
	if utils.IsFileExists("requirements.txt") && utils.IsFileExists("wsgi.py") {
		log.Println("python runtime detected.")
		return newPythonRuntime(projectPath)
	}
	if utils.IsFileExists("composer.json") && utils.IsFileExists(filepath.Join("public", "index.php")) {
		log.Println("php runtime detected.")
		return newPhpRuntime(projectPath)
	}
	return nil, errors.New("invalid runtime")
}

func newPythonRuntime(projectPath string) (*Runtime, error) {
	execName := "python2.7"

	if content, err := ioutil.ReadFile(filepath.Join(projectPath, "runtime.txt")); err == nil {
		if strings.HasPrefix(string(content), "python-2.7") {
			execName = "python2.7"
		} else if strings.HasPrefix(string(content), "python-3.5") {
			execName = "python3.5"
		} else {
			return nil, errors.New("invalid python runtime.txt format, only `python-2.7` and `python-3.5` were allowed")
		}
	}

	// for windows don't have a pythonx.x symbol link
	if _, err := exec.LookPath(execName); err != nil {
		log.Printf("`%s` command not found, fallback to `python`", execName)
	}

	return &Runtime{
		Name:       "python",
		Exec:       execName,
		Args:       []string{"wsgi.py"},
		WatchFiles: []string{"."},
		Envs:       os.Environ(),
		DeployFiles: filesPattern{
			Includes: []string{"**"},
			Excludes: []string{
				".git/**",
				".avoscloud/**",
				".leancloud/**",
				"*.pyc",
			},
		},
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
	}, nil
}

func newPhpRuntime(projectPath string) (*Runtime, error) {
	return &Runtime{
		Name:       "php",
		Exec:       "php",
		Args:       []string{"-S", "127.0.0.1:3000", "-t", "public"},
		WatchFiles: []string{"."},
		Envs:       os.Environ(),
	}, nil
}
