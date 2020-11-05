package folderWatcher

import (
	"errors"
	"fmt"
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
}

func New() Watcher {
	newWatcher := &Watcher{
		RequestedWatches: make(map[string]WatchRequest),
		Interval: MinimumIntervalTime,
	}

	return *newWatcher
}

func (w *Watcher) AddFolder(path string, recursive bool) (err error){
	// check that the path is valid, return error if it's not
	if !IsValidPath(path){
		err = errors.New(fmt.Sprintf("%s is not a valid path", path))
		return
	}
	w.RequestedWatches[path] = WatchRequest{Path: path, Recursive: recursive}
	return
}

func (w *Watcher) RemoveFolder(path string, returnErrorIfNotFound bool) (error){
	if _, found := w.RequestedWatches[path]; !found{
		// the path was not in the collection
		if (returnErrorIfNotFound) {
			return errors.New(fmt.Sprintf("Cannot remove %s from RequestedWatches list because it is not a member of the list", path))
		}
	}
	delete(w.RequestedWatches, path)
	return nil
}