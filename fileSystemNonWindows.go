// +build !windows

package folderWatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func isHiddenFile(filePath string) bool {
	return strings.HasPrefix(filepath.Base(filePath), ".")
}

func hideFile(filepath string) (err error){
	if !isHiddenFile(filepath){
		err = os.Rename(filepath, fmt.Sprint(".%s", filepath))
	}
	return
}
