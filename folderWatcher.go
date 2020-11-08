package folderWatcher

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"
)

const (
	MinimumIntervalTime = 500
	MaximumIntervalTime = 5000
)


type WatchRequest struct {
	Path string
	Recursive bool
	ShowHidden bool
}



type Watcher struct {
	RequestedWatches map[string]WatchRequest
	Interval int
	watchedFiles map[string]os.FileInfo
	Stop chan bool
	Stopped chan bool
	FileChanged chan FileEvent
}

func New() Watcher {
	newWatcher := &Watcher{
		RequestedWatches: make(map[string]WatchRequest),
		Interval: MinimumIntervalTime,
		watchedFiles: make(map[string]os.FileInfo),
		Stop : make(chan bool),
		Stopped: make(chan bool),
		FileChanged: make(chan FileEvent),
	}

	return *newWatcher
}

func CalculateInterval(watchedFileCount int) int{
	weightedInt := float64(watchedFileCount) * .5
	// enforce min and max
	return int(math.Min(math.Max(MinimumIntervalTime, weightedInt), MaximumIntervalTime))
}

func (w *Watcher) UpdateInterval(){
	w.Interval = CalculateInterval(len(w.watchedFiles))
}

func (w *Watcher) AddFolder(path string, recursive bool, showHidden bool) (err error){
	// check that the path is valid, return error if it's not
	if !IsValidPath(path){
		err = errors.New(fmt.Sprintf("%s is not a valid path", path))
		return
	}
	// add the path to the list of watched folders
	w.RequestedWatches[path] = WatchRequest{Path: path, Recursive: recursive, ShowHidden: showHidden}

	// Add the new set of files to watch
	newFilesToWatch, err := GetFileList(path, recursive, showHidden)
	if err!=nil {
		return
	}

	// add files to the watchFiles list
	for p, file := range newFilesToWatch{
		w.watchedFiles[p]=file
	}
	w.UpdateInterval()
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

	// get a list of files to remove. Including recursive and hidden files, even though they may not have been included
	// when the folder was added.
	watchedFilesToRemove, err := GetFileList(path, true, true)
	if err!=nil{
		return
	}

	// remove the files from the watchedFiles map
	for p, _ := range watchedFilesToRemove{
		delete(w.watchedFiles, p)
	}
	delete(w.RequestedWatches, path)
	w.UpdateInterval()
	return
}

func (w *Watcher) scanForFileEvents(){
	for _, requestedWatch := range w.RequestedWatches{
		newFileList, err := GetFileList(requestedWatch.Path, requestedWatch.Recursive, requestedWatch.ShowHidden)
		if err != nil {
			fmt.Println(err.Error())
		}

		// look for added files
		go func(){
			for newFilePath, newFile := range newFileList {
				existingFile, isExistingFile := w.watchedFiles[newFilePath]
				if !isExistingFile {
					// new file was found
					w.watchedFiles[newFilePath] = newFile
					w.FileChanged<- FileEvent{FileChange:Add, FilePath: newFilePath}
				} else {
					// check for file writes
					if newFile.ModTime() != existingFile.ModTime(){
						w.FileChanged<-FileEvent{FileChange:Write, FilePath: newFilePath}
						w.watchedFiles[newFilePath] = newFile
					}
				}
			}
		}()

		// look for removed files
		go func() {
			for path, _:= range w.watchedFiles{
				_,stillExists := newFileList[path]
				if !stillExists{
					w.FileChanged<- FileEvent{FileChange: Remove, FilePath: path}
					delete(w.watchedFiles, path)
				}
			}
		}()
	}
}

func (w *Watcher) StopWatch(){
	// TODO Change the state
	fmt.Println("StopWatch()")
	w.Stopped<-true
}

func (w *Watcher) Start(){
	intervalChan := make(chan bool)
	runServiceLoop := true
	// Service loop
	go func(){
		for {
			time.Sleep(time.Duration(w.Interval) * time.Millisecond )

			// exit service loop
			if !runServiceLoop{
				break
			}

			intervalChan<-true
		}

	}()

	// event loop
	go func(){
		for {
			select {
				case <- w.Stop:
					runServiceLoop=false
					w.StopWatch()
				case <- intervalChan:
					//fmt.Printf("Watching %d files\n", len(w.watchedFiles))
					w.scanForFileEvents()

			}

		}

	}()


}