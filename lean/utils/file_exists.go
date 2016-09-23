package utils

import "os"

// IsFileExists returns whether the file exists
func IsFileExists(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return false
	}
	if fileInfo.IsDir() {
		return false
	}
	return true
}
