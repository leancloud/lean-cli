package utils

import (
	"archive/zip"
	"bufio"
	"os"
	"path/filepath"
)

// ArchiveFiles will make a zip file archive to targetPath
func ArchiveFiles(targetPath string, files []struct{ Name, Path string }) error {
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	writer := zip.NewWriter(targetFile)
	// close order is important
	defer targetFile.Close()
	defer writer.Close()

	for _, file := range files {
		zippedFile, err := writer.Create(filepath.ToSlash(file.Name))
		if err != nil {
			return err
		}
		fromFile, err := os.Open(file.Path)
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
	}
	return nil
}
