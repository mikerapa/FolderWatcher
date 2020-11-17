package folderWatcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsValidDirPath(path string) bool {
	path = strings.Trim(path, " ")
	if len(path) == 0 {
		return false
	}

	fileItem, err := os.Stat(path)
	return !os.IsNotExist(err) && fileItem.Mode().IsDir()
}

// Check if the path is valid. Could be a folder or file path. 
func IsValidPath(path string) bool {
	path = strings.Trim(path, " ")
	if len(path) == 0 {
		return false
	}

	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func AbsPath(path string) (absPath string) {
	absPath, _ = filepath.Abs(path)
	return
}

func GetFileList(folderPath string, recursive bool, showHidden bool) (fileList map[string]os.FileInfo, err error){
	// make sure the path provided is valid
	if !IsValidDirPath(folderPath){
		err = errors.New(fmt.Sprintf("%s is not a valid folder path", folderPath))
		return
	}
	fileList = make(map[string]os.FileInfo)

	err = filepath.Walk(folderPath, func(filePath string, fileInfo os.FileInfo, err error) error{
		if err != nil{
			println(err.Error())
		}

		// check if the file is hidden before adding it
		if !fileInfo.IsDir() && (showHidden || !isHiddenFile(filePath)) && (recursive || filepath.Dir(filePath) == folderPath) {
			fileList[filePath] = fileInfo
		}
		return nil
	})

	return
}
