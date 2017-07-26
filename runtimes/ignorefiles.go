package runtimes

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/facebookgo/parseignore"
	"github.com/leancloud/lean-cli/logger"
	"github.com/leancloud/lean-cli/utils"
)

// defaultIgnorePatterns returns current runtime's default ignore patterns
func (runtime *Runtime) defaultIgnorePatterns() []string {
	switch runtime.Name {
	case "node.js":
		return []string{
			".git/",
			".DS_Store",
			".avoscloud/",
			".leancloud/",
			"node_modules/",
		}
	case "java":
		return []string{
			".git/",
			".DS_Store",
			".avoscloud/",
			".leancloud/",
			".project",
			".classpath",
			".settings/",
			"target/",
		}
	case "php":
		return []string{
			".git/",
			".DS_Store",
			".avoscloud/",
			".leancloud/",
			"vendor/",
		}
	case "python":
		return []string{
			".git/",
			".DS_Store",
			".avoscloud/",
			".leancloud/",
			"venv",
			"*.pyc",
			"__pycache__/",
		}
	default:
		panic("invalid runtime")
	}
}

func (runtime *Runtime) readIgnore(ignoreFilePath string) (parseignore.Matcher, error) {
	if ignoreFilePath == ".leanignore" && !utils.IsFileExists(filepath.Join(runtime.ProjectPath, ".leanignore")) {
		logger.Warn("没有找到 .leanignore 文件，根据项目文件创建默认的 .leanignore 文件")
		content := strings.Join(runtime.defaultIgnorePatterns(), "\r\n")
		err := ioutil.WriteFile(filepath.Join(runtime.ProjectPath, ".leanignore"), []byte(content), 0644)
		if err != nil {
			return nil, err
		}
	}

	content, err := ioutil.ReadFile(ignoreFilePath)
	if err != nil {
		return nil, err
	}

	matcher, errs := parseignore.CompilePatterns(content)
	if len(errs) != 0 {
		return nil, errs[0]
	}

	return matcher, nil
}
