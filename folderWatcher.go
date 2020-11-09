package folderWatcher

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
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
		if returnErrorIfNotFound {
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
	for p := range watchedFilesToRemove{
		delete(w.watchedFiles, p)
	}
	delete(w.RequestedWatches, path)
	w.UpdateInterval()
	return
}

func (w *Watcher) scanForFileEvents(){
	// TODO need to put a mutex around the watchedFiles map
	// TODO this probably has a concurrency issue because the list of new files it outside of the go routine
	for _, requestedWatch := range w.RequestedWatches{
		watchedFileMutex := sync.Mutex{}
		newFileList, err := GetFileList(requestedWatch.Path, requestedWatch.Recursive, requestedWatch.ShowHidden)
		if err != nil {
			fmt.Println(err.Error())
		}

		// start processing by going through all of the flies in the new list
		go func(){
			movedFiles := make(map[string]string) // oldPath[newPath]
			for newFilePath, newFile := range newFileList {
				existingFile, isExistingFile := w.watchedFiles[newFilePath]
				if !isExistingFile {
					// a file in the new list of files was not found in the watchedFiles map. It could be a new file, or
					// it could be a file which has moved.
					// Look for existing watchedFile which match
					matchFoundInWatchedFiles:= false
					for path,watchedFile:= range w.watchedFiles {
						if os.SameFile(watchedFile, newFile){
							// a matching file was found in the watched files list. Count this as a moved file.
							matchFoundInWatchedFiles = true
							movedFiles[path] = newFilePath
							w.FileChanged <- FileEvent{FileChange: Move,
								FilePath: newFilePath,
								PreviousPath: path,
								Description: fmt.Sprintf("File was moved from %s to %s", path, newFilePath)}
							watchedFileMutex.Lock()
							delete(w.watchedFiles, path)
							w.watchedFiles[newFilePath] = newFile
							watchedFileMutex.Unlock()
							break
						}
					}
					// The file is in the new list of files, but not the watchedFiles list and the file was not moved.
					// Process this as a new file.
					if !matchFoundInWatchedFiles{
						watchedFileMutex.Lock()
						w.watchedFiles[newFilePath] = newFile
						watchedFileMutex.Unlock()
						w.FileChanged<- FileEvent{FileChange:Add, FilePath: newFilePath,
							Description: fmt.Sprintf("%s was added", newFilePath)}
					}
				} else {
					// The new list and pre-existing list have a matching path.
					// Check for file writes.
					if newFile.ModTime() != existingFile.ModTime(){
						w.FileChanged<-FileEvent{FileChange:Write, FilePath: newFilePath,
						Description: fmt.Sprintf("%s updated", newFilePath)}
						w.watchedFiles[newFilePath] = newFile
					}
				}
			}

			// find deleted files
			for path:= range w.watchedFiles{
				_,isInNewFilesList := newFileList[path]
				_,isMovedFile := movedFiles[path]
				if !isInNewFilesList && ! isMovedFile{
					w.FileChanged<- FileEvent{FileChange: Remove, FilePath: path,
						Description: fmt.Sprintf("%s was removed", path)}
					watchedFileMutex.Lock()
					delete(w.watchedFiles, path)
					watchedFileMutex.Unlock()
				}
			}
		}()

	}
}

func (w *Watcher) StopWatch(){
	// TODO Change the state
	//fmt.Println("StopWatch()")
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