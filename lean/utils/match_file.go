package utils

import (
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
	debug_ "github.com/tj/go-debug"
)

var debug = debug_.Debug("lean:matchFile")

// MatchFiles will return all files matches the given pattern
// TODO: not elegent
func MatchFiles(dir string, includes []string, excludes []string) ([]string, error) {
	matchedFiles := map[string]bool{}
	// get all files which matches includes patterns
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		for _, include := range includes {
			matched, err := doublestar.PathMatch(include, path)
			if err != nil {
				return err
			}
			if matched {
				matchedFiles[path] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// get all files which matches exludes patterns from matched files
	for path := range matchedFiles {
		for _, exclude := range excludes {
			matched, err := doublestar.PathMatch(exclude, path)
			if err != nil {
				return nil, err
			}
			if matched {
				delete(matchedFiles, path)
				break
			}
		}

	}

	result := []string{}
	for file := range matchedFiles {
		debug(file)
		result = append(result, file)
	}

	return result, nil
}
