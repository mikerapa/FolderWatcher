// +build windows

package folderWatcher

import "syscall"

func isHiddenFile(filePath string) bool {
	prt, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return false
	}

	attributes, err := syscall.GetFileAttributes(prt)
	if err != nil {
		return false
	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}

func hideFile(filepath string) (err error){
	winFilepath, err := syscall.UTF16PtrFromString(filepath)
	if err !=nil {
		return
	}

	// set the file attribute
	err = syscall.SetFileAttributes(winFilepath, syscall.FILE_ATTRIBUTE_HIDDEN)
	return
}
