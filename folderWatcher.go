package folderWatcher

import (
	"errors"
	"fmt"
	"os"
)

const (
	MinimumIntervalTime = 500
	MaximumIntervalTime = 5000
)

type WatchRequest struct {
	Path string
	Recursive bool
}

type Watcher struct {
	RequestedWatches map[string]WatchRequest
	Interval int
	watchedFiles map[string]os.FileInfo
}

func New() Watcher {
	newWatcher := &Watcher{
		RequestedWatches: make(map[string]WatchRequest),
		Interval: MinimumIntervalTime,
		watchedFiles: make(map[string]os.FileInfo),
	}

	return *newWatcher
}

func (w *Watcher) AddFolder(path string, recursive bool, showHidden bool) (err error){
	// check that the path is valid, return error if it's not
	if !IsValidPath(path){
		err = errors.New(fmt.Sprintf("%s is not a valid path", path))
		return
	}
	// add the path to the list of watched folders
	w.RequestedWatches[path] = WatchRequest{Path: path, Recursive: recursive}

	// Add the new set of files to watch
	newFilesToWatch, err := GetFileList(path, recursive, showHidden)
	if err!=nil {
		return
	}

	// add files to the watchFiles list
	for p, file := range newFilesToWatch{
		w.watchedFiles[p]=file
	}

	return
}

func (w *Watcher) RemoveFolder(path string, returnErrorIfNotFound bool) ( err error){
	if _, found := w.RequestedWatches[path]; !found{
		// the path was not in the collection
		if (returnErrorIfNotFound) {
			err = errors.New(fmt.Sprintf("Cannot remove %s from RequestedWatches list because it is not a member of the list", path))
		}
		// even if an error should not be returned, do not proceed with the rest of the function because the
		// requested path does not exist in the map.
		return
	}

	// get a list of files to remove
	watchedFilesToRemove, err := GetFileList(path, true, true)
	if err!=nil{
		return
	}

	// remove the files from the watchedFiles map
	for p, _ := range watchedFilesToRemove{
		delete(w.watchedFiles, p)
	}
	delete(w.RequestedWatches, path)
	return
}