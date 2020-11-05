// +build windows

package folderWatcher

// TODO Test this on a Windows machine
func isHiddenFile(filePath string) bool {
	prt, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}

	attributes, err := syscall.GetFileAttributes(prt)
	if err != nil {
		return false, err
	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}
