package folderWatcher

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
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

// constants to represent the state of the watcher
const (NotStarted WatcherState = 1
	Running WatcherState = 2
	Stopped WatcherState =3 )
type WatcherState int

func (ws WatcherState) String() string {
	stateStrings:= [...]string{"Not Started", "Started", "Stopped"}
	return stateStrings[ws]
}

func (ws *WatcherState) ToString() string {
	return fmt.Sprintf("%v", ws)
}


type Watcher struct {
	RequestedWatches map[string]WatchRequest
	Interval int
	watchedFiles map[string]os.FileInfo
	Stopped chan bool
	FileChanged chan FileEvent
	watchedFileMutex *sync.RWMutex
	State WatcherState
}

func New() Watcher {
	newWatcher := &Watcher{
		RequestedWatches: make(map[string]WatchRequest),
		Interval: MinimumIntervalTime,
		watchedFiles: make(map[string]os.FileInfo),
		Stopped: make(chan bool),
		FileChanged: make(chan FileEvent),
		watchedFileMutex: &sync.RWMutex{},
		State: NotStarted,
	}

	return *newWatcher
}

// Remove a file from the watchedFiles list protecting the map with a mutex
func (w *Watcher) removeWatchedFile(filepath string){
	w.watchedFileMutex.Lock()
	delete(w.watchedFiles, filepath)
	w.watchedFileMutex.Unlock()
}

// Add an item to the watchedFiles list while protecting it with a mutex
func (w *Watcher) addUpdateWatchedFile(filepath string, file os.FileInfo){
	w.watchedFileMutex.Lock()
	w.watchedFiles[filepath] = file
	w.watchedFileMutex.Unlock()
}


func calculateInterval(watchedFileCount int) int{
	weightedInt := float64(watchedFileCount) * .5
	// enforce min and max
	return int(math.Min(math.Max(MinimumIntervalTime, weightedInt), MaximumIntervalTime))
}


func (w *Watcher) updateInterval(){
	w.Interval = calculateInterval(len(w.watchedFiles))
}

func (w *Watcher) AddFolder(path string, recursive bool, showHidden bool) (err error){
	path, err  = filepath.Abs(path)
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
		w.addUpdateWatchedFile(p, file)
	}
	return
}

func (w *Watcher) RemoveFolder(path string, returnErrorIfNotFound bool) ( err error){
	path, err = filepath.Abs(path)
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
		w.removeWatchedFile(p)
	}
	delete(w.RequestedWatches, path)
	return
}

func findMatchingFile(fileToMatch os.FileInfo, fileList map[string]os.FileInfo) (matchFound bool, matchedFilePath string){
	for path,watchedFile:= range fileList {
		if os.SameFile(watchedFile, fileToMatch) {
			//  SameFile check does not work the same on Windows
			matchedFilePath = path
			matchFound = true
			break
		}
	}
	return
}

func (w *Watcher) scanForFileEvents() {
	// get a refreshed list of all the files in the watched folders
	newFileList := make(map[string]os.FileInfo)
	var newFileChan = make(chan map[string]os.FileInfo, 100)

	go func() {

		for _, requestedWatch := range w.RequestedWatches {
			// start processing by going through all of the flies in the new list
			fl, err := GetFileList(requestedWatch.Path, requestedWatch.Recursive, requestedWatch.ShowHidden)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				newFileChan <- fl
			}
		}
		close(newFileChan)

	}()

	movedFiles := make(map[string]string) // oldPath[newPath]
	// collect the lists of files
	for fl := range newFileChan{
		for newFilePath, newFile:= range fl{
			newFileList[newFilePath] = newFile
			//fmt.Printf("adding %s\n", newFilePath)
			existingFile, isExistingFile := w.watchedFiles[newFilePath]
			if isExistingFile {
				// The new list and pre-existing list have a matching path.
				// Check to see if the file has been updated.
				if newFile.ModTime() != existingFile.ModTime(){
					w.FileChanged<-FileEvent{FileChange:Write, FilePath: newFilePath,
						Description: fmt.Sprintf("%s updated", newFilePath)}
				}
			} else {
				// a file in the new list of files was not found in the watchedFiles map. It could be a new file, or
				// it could be a file which has moved.
				matchFound, matchPath := findMatchingFile(newFile, w.watchedFiles)
				if matchFound{
					movedFiles[matchPath] = newFilePath
					w.FileChanged <- FileEvent{FileChange: Move,
						FilePath: newFilePath,
						PreviousPath: matchPath,
						Description: fmt.Sprintf("%s move to %s", matchPath, newFilePath)}
				} else {
					// The file is in the new list of files, but not the watchedFiles list and the file was not moved.
					// Process this as a new file.
					w.FileChanged<- FileEvent{FileChange:Add, FilePath: newFilePath,
							Description: fmt.Sprintf("%s created", newFilePath)}
				}
			}

		}
	}


	// find deleted files
	for path:= range w.watchedFiles{
		_,isInNewFilesList := newFileList[path]
		_,isMovedFile := movedFiles[path]
		if !isInNewFilesList && ! isMovedFile{
			w.FileChanged<- FileEvent{FileChange: Remove, FilePath: path,
				Description: fmt.Sprintf("%s deleted", path)}
		}
	}
	// replace the watch list with the newly created map
	w.watchedFiles = newFileList
	w.updateInterval()

}

// Stop the watcher
func (w *Watcher) Stop(){
	w.State = Stopped
	w.Stopped<-true
}

func (w *Watcher) Start(){
	if w.State == Running{
		// If STart is called when the service is already running, do not start another service loop
		return
	}

	w.State = Running
	// Service loop
	go func(){
		for {
			time.Sleep(time.Duration(w.Interval) * time.Millisecond )

			// exit service loop
			if w.State != Running{
				break
			}

			w.scanForFileEvents()
		}

	}()




}