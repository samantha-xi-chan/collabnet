package filems

import (
	"github.com/pkg/errors"
	"os"
	"strings"
)

func IsLinuxPath(path string) bool {
	// 确保路径不为空
	if path == "" {
		return false
	}

	// 确保路径以斜杠开始
	if !strings.HasPrefix(path, "/") {
		return false
	}

	// 如果所有条件都通过，认为是Linux路径
	return true
}

func CreateFolderIfNotExists(folderPath string) error {
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func CheckFileReady(absPath string) (e error) {
	_, err := os.Stat(absPath)
	if err == nil {
		return nil
	} else if os.IsNotExist(err) {
		return errors.Wrap(err, "os.IsNotExist: ")
	} else {
		return errors.Wrap(err, "other: ")
	}
}
