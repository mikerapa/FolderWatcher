package folderWatcher

import (
	"math"
	"testing"
	"time"
)



// Make sure the New funciton returns a valid watcher
func TestNew(t *testing.T) {
	newWatcher := New()

	if &newWatcher== nil {
		t.Error("folderWatcher.New() returned nil")
	}

	if len(newWatcher.RequestedWatches) != 0 {
		t.Error("folderWatcher.New() should created an empty RequestedWatches map")
	}

	// check the default interval
	if newWatcher.Interval != 500 {
		t.Errorf("folderWatch.New() is the incorrect default interval. got %d, wanted %d", newWatcher.Interval, MinimumIntervalTime)
	}

}

func TestWatcher_AddFolder(t *testing.T) {
	tests := []struct {
		name    string
		path      string
		recursive bool
		showHidden bool
		wantAdd bool
		wantErr bool
	}{
		{name: "add a bad path", path: "whehrkh", recursive: false, showHidden:false, wantAdd: false, wantErr: true},
		{name: "add a valid path", path: "testFolder", recursive: false, showHidden:false, wantAdd: true, wantErr: false},
	}
	// populate the test folder with test files
	setupTestFiles()
	defer tearDownTestFiles()
	for _, tt := range tests {
		watcher := New()
		t.Run(tt.name, func(t *testing.T) {
			if err := watcher.AddFolder(tt.path, tt.recursive, tt.showHidden); (err != nil) != tt.wantErr {
				t.Errorf("%s AddFolder() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			// check if the path was added to the list
			if (len(watcher.RequestedWatches)==0) == tt.wantAdd{
				t.Errorf("%s AddFolder() wantAdd = %v, got %v", tt.name, tt.wantAdd, len(watcher.RequestedWatches)==0)
			}

			if tt.wantAdd && len(watcher.watchedFiles)==0{
				t.Errorf("%s AddFolder() files should have been added to the watchedFiles map. 0 were found.", tt.name)
			}
		})
	}
}

func TestWatcher_RemoveFolder(t *testing.T) {
	tests := []struct {
		name    string
		path                  string
		returnErrorIfNotFound bool
		shouldRemoveFolder bool
	}{
		{name:"folder not in the list and do not return error", path:"dummypath", returnErrorIfNotFound: false, shouldRemoveFolder: false},
		{name: "folder not in the list and return error", path: "dummypath", returnErrorIfNotFound: true, shouldRemoveFolder: false},
		{name: "folder path in the list", path:"testFolder", returnErrorIfNotFound: false, shouldRemoveFolder: true},
	}
	// populate the test folder with test files
	setupTestFiles()
	defer tearDownTestFiles()
	for _, tt := range tests {
		watcher := New()
		watcher.AddFolder("testFolder", true, false)

		t.Run(tt.name, func(t *testing.T) {
			// make sure an error is returned if it should
			if err := watcher.RemoveFolder(tt.path, tt.returnErrorIfNotFound); (err != nil) != tt.returnErrorIfNotFound {
				t.Errorf("RemoveFolder() error = %v, wantErr %v", err, tt.returnErrorIfNotFound)
			}

			// test that the folder was removed
			if tt.shouldRemoveFolder && len(watcher.RequestedWatches) != 0 {
				t.Errorf("the RequestedWatches list should be empty")
			}

			// make sure the files are removed from the watchedFiles list
			if tt.shouldRemoveFolder && len(watcher.watchedFiles)!=0{
				t.Errorf("%s RemoveFolder() after removing path, there should be 0 files watched.", tt.name)
			}

		})
	}
}

func TestAddFileEvent(t *testing.T) {
	watcher := New()
	watcher.AddFolder("testFolder", false, false)
	var receivedFileEvent FileEvent
	watcher.Start()
	// collect events from the watcher
	go func () {
		for {
			select{
			case <- watcher.Stopped:
				return
			case receivedFileEvent= <- watcher.FileChanged:
			}
		}
	}()

	writeToFile(normalFilePath, "nothing")
	defer deleteFile(normalFilePath)

	time.Sleep(1 * time.Second)
	watcher.Stop<-true

	// make sure the current type of FileEvent was received
	if receivedFileEvent.FileChange != Add {
		t.Errorf("Watcher did not send Add FileEvent to the FileChanged channel")
	}

	// make sure the path was included in the FileEvent
	if receivedFileEvent.FilePath != normalFilePath{
		t.Errorf("Filepath was not included in the FileEvent")
	}

	// make sure the new file is in the watchedFiles
	_, fileFound := watcher.watchedFiles[normalFilePath]
	if !fileFound{
		t.Errorf("%s should have been added to the watched files", normalFilePath)
	}

}


func TestWriteFileEvent(t *testing.T) {
	// set up the watcher and a file
	writeToFile(normalFilePath, "nothing")
	watcher := New()
	watcher.AddFolder("testFolder", false, false)

	// get the mod time for the newly created file. This will be used to for a comparison.
	testFile,_ :=watcher.watchedFiles[normalFilePath]
	initialModTime := testFile.ModTime()

	defer deleteFile(normalFilePath)

	// start the watcher and collect events
	var receivedFileEvent FileEvent
	watcher.Start()

	go func () {
		for {
			select{
			case <- watcher.Stopped:
				return
			case receivedFileEvent= <- watcher.FileChanged:
			}
		}
	}()

	// update the file and give some time to get the FileEvent on the channel
	time.Sleep(1 * time.Second)
	writeToFile(normalFilePath, "updated file content")
	time.Sleep(2 * time.Second)



	// make sure the correct FileEvent was received
	if receivedFileEvent.FileChange != Write {
		t.Errorf("Watcher did not send Write FileEvent to the FileChanged channel")
	}

	// make sure the path is included in the FileEvent
	if receivedFileEvent.FilePath != normalFilePath{
		t.Errorf("Filepath was not included in the FileEvent")
	}

	// make sure the updated file remains in the watched files list
	newWatchedFile, fileFound := watcher.watchedFiles[normalFilePath]
	if !fileFound{
		t.Errorf("file %s should have remained in the list of watched files", normalFilePath)
	}

	// compare the mod time of the file in the watchedFiles list. It should be different that the orignial value
	if newWatchedFile.ModTime() == initialModTime{
		t.Errorf("the file in watchedFiles was not updated to one with a different modtime.")
	}
	watcher.Stop<-true
}



func TestRemoveFileEvent(t *testing.T) {
	// set up the watcher and a file
	watcher := New()
	watcher.AddFolder("testFolder", false, false)
	writeToFile(normalFilePath, "nothing")
	defer deleteFile(normalFilePath)

	// start the watcher and collect events
	var receivedFileEvent FileEvent
	watcher.Start()

	go func () {
		for {
			select{
			case <- watcher.Stopped:
				return
			case receivedFileEvent= <- watcher.FileChanged:
			}
		}
	}()

	// delete the file and give some time to get the FileEvent on the channel
	time.Sleep(1 * time.Second)
	deleteFile(normalFilePath)
	time.Sleep(1 * time.Second)

	watcher.Stop<-true

	// make sure the correct FileEvent was received
	if receivedFileEvent.FileChange != Remove {
		t.Errorf("Watcher did not send Remove FileEvent to the FileChanged channel")
	}

	// make sure the path is included in the FileEvent
	if receivedFileEvent.FilePath != normalFilePath{
		t.Errorf("Filepath was not included in the FileEvent")
	}

	// make sure the removed file is no longer being watched
	_, fileFound := watcher.watchedFiles[normalFilePath]
	if fileFound{
		t.Errorf("file was not removed with the Remove FileEvent path=%s", normalFilePath)
	}

}


func TestCalculateInterval(t *testing.T) {
	tests := []struct {
		name string
		watchedFileCount int
	}{
		{name: "smallest number", watchedFileCount: 1},
		{name: "largest number", watchedFileCount: math.MaxInt32},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculatedValue :=  CalculateInterval(tt.watchedFileCount)
			if calculatedValue< MinimumIntervalTime || calculatedValue > MaximumIntervalTime{
				t.Errorf("CalculateInterval() = %v is out of range. Min:%v, Max:%v", calculatedValue, MinimumIntervalTime, MaximumIntervalTime)
			}
		})
	}
}