package utils

import (
	"os"
	"path/filepath"
)

func FindRootPath() string {
	dirPath, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dirPath, "go.mod")); err == nil {
			return dirPath
		}
		parentDirPath := filepath.Dir(dirPath)
		if parentDirPath == dirPath {
			break
		}
		dirPath = parentDirPath
	}
	return ""
}
