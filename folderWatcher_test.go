package folderWatcher

import (
	"math"
	"strings"
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

	// make sure the watcher is created in the correct state
	if newWatcher.State != NotStarted{
		t.Errorf("folderWatcher.New() watcher should have been created in the NotStarted state, got %s",  newWatcher.State.ToString())
	}

}

func TestWatcher_Stop(t *testing.T) {
	gotStop:= false
	watcher := New()
	go func () {
		for {
			select{
			case <- watcher.Stopped:
				gotStop=true
				return
			}
		}
	}()
	watcher.Start()

	// Make sure the watcher is in the running state
	if watcher.State!=Running{
		t.Errorf("Watcher should be in the 'Running state. State=%s", watcher.State.ToString())
	}

	watcher.Stop()
	time.Sleep(1 * time.Second)

	if !gotStop{
		t.Errorf("Calling Stop() did not send stopped message")
	}

	if watcher.State != Stopped{
		t.Errorf("Watcher should be in the Stopped state after the Stop() function is called. State=%s", watcher.State.ToString())
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

// TODO does the watcher work when the channel listening is before the Start func is called?


func TestRestartWatcher(t *testing.T) {
	watcher := New()
	watcher.AddFolder("testFolder/subFolder", false, false)

	var testFiles []string
	recievedEvents := make(map[FileChange]int)


	// set up the test file paths
	testFiles = append(testFiles,
		randomizedFilePath("testFolder/subFolder/file#.txt"),
		randomizedFilePath("testFolder/subFolder/file#.txt"))


	// create and update each of the files
	for _,fp := range testFiles {


		// collect events from the watcher
		go func () {
			for {
				select{
				case <- watcher.Stopped:
					return
				case evnt:= <- watcher.FileChanged:
					recievedEvents[evnt.FileChange] = recievedEvents[evnt.FileChange]+1
				}
			}
		}()
		watcher.Start()

		time.Sleep(1 * time.Second)
		writeToFile(fp, "stuff")
		time.Sleep(1 * time.Second)
		writeToFile(fp, "updated stuff")
		time.Sleep(1 * time.Second)
		watcher.Stop()
		time.Sleep(1 * time.Second)
	}

	// make sure the correct number of events were received
	var eventType FileChange
	for _, eventType= range []FileChange{Add, Write}{
		if recievedEvents[eventType]!=2{
			t.Errorf("should have received 2 %s events, got %d", FileChange(eventType), recievedEvents[eventType] )
		}
	}

	removeFiles(false, testFiles...)
}

func TestMultipleWatchRequests(t *testing.T){
	watcher := New()
	watcher.AddFolder("testFolder/subFolder", false, false)
	watcher.AddFolder("testFolder/subFolder2", false, false)

	var testFiles []string
	recievedEvents := make(map[FileChange]int)

	watcher.Start()
	// collect events from the watcher
	go func () {
		for {
			select{
			case <- watcher.Stopped:
				return
			case evnt:= <- watcher.FileChanged:
				recievedEvents[evnt.FileChange] = recievedEvents[evnt.FileChange]+1
			}
		}
	}()

	time.Sleep(1 * time.Second)
	testFiles = append(testFiles,
		randomizedFilePath("testFolder/subFolder/file#.txt"),
		randomizedFilePath("testFolder/subFolder2/file#.txt"))

	// create, update and delete each of the files
	for _,fp := range testFiles {
		writeToFile(fp, "stuff")
		time.Sleep(1 * time.Second)
		writeToFile(fp, "updated stuff")
		time.Sleep(1 * time.Second)
	}

	removeFiles(false, testFiles...)

	time.Sleep(1 * time.Second)
	// make sure the correct number of events were received
	var eventType FileChange
	for _, eventType= range []FileChange{Add, Remove, Write}{
		if recievedEvents[eventType]!=2{
			t.Errorf("should have received 2 %s events, got %d", FileChange(eventType), recievedEvents[eventType] )
		}
	}

	watcher.Stop()

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
	defer removeFiles(true, normalFilePath)

	time.Sleep(1 * time.Second)
	watcher.Stop()

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

	// Make sure the description is set
	if !strings.Contains(receivedFileEvent.Description, receivedFileEvent.FilePath){
		t.Errorf("FileEvent description should be set and contain the filepath")
	}
}


func TestMoveFileEvent(t *testing.T) {
	// set up the watcher and a file
	writeToFile(normalFilePath, "nothing")
	watcher := New()
	watcher.AddFolder("testFolder", true, false)

	// clean up the files that were used in this test
	defer removeFiles(false, normalFilePath)
	defer removeFiles(true, normalFilePath2)

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

	// Move the file and give some time to get the FileEvent on the channel
	time.Sleep(1 * time.Second)
	moveFile(normalFilePath, normalFilePath2)
	time.Sleep(2 * time.Second)

	// make sure the correct FileEvent was received
	if receivedFileEvent.FileChange != Move {
		t.Errorf("Watcher did not send Move FileEvent to the FileChanged channel. Wanted %v, got %v", Move, receivedFileEvent.FileChange)
	}

	// make sure the path and previous path are included in the FileEvent
	if receivedFileEvent.FilePath != normalFilePath2{
		t.Errorf("Filepath is not correct in FileEvent. Wanted %s, got %s", normalFilePath2, receivedFileEvent.FilePath)
	}
	if receivedFileEvent.PreviousPath!=normalFilePath{
		t.Errorf("FileEvent should have PreviousPath value set for a Move event. Wanted '%s', got '%s'", normalFilePath2, receivedFileEvent.PreviousPath)
	}

	// Make sure the description is set
	if !strings.Contains(receivedFileEvent.Description, receivedFileEvent.FilePath){
		t.Errorf("FileEvent description should be set and contain the filepath")
	}

	// Make sure the new path is in the watchedFiles map
	_, fileFound := watcher.watchedFiles[normalFilePath2]
	if !fileFound{
		t.Errorf("file %s should be in the list of watched files", normalFilePath)
	}

	watcher.Stop()
}


func TestWriteFileEvent(t *testing.T) {
	// set up the watcher and a file
	writeToFile(normalFilePath, "nothing")
	watcher := New()
	watcher.AddFolder("testFolder", false, false)

	// get the mod time for the newly created file. This will be used to for a comparison.
	testFile,_ :=watcher.watchedFiles[normalFilePath]
	initialModTime := testFile.ModTime()

	defer removeFiles(true, normalFilePath)

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
	watcher.Stop()


	// Make sure the description is set
	if !strings.Contains(receivedFileEvent.Description, receivedFileEvent.FilePath){
		t.Errorf("FileEvent description should be set and contain the filepath")
	}
}



func TestRemoveFileEvent(t *testing.T) {
	// set up the watcher and a file
	watcher := New()
	watcher.AddFolder("testFolder", false, false)
	writeToFile(normalFilePath, "nothing")
	defer removeFiles(false, normalFilePath)

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
	removeFiles(true, normalFilePath)
	time.Sleep(1 * time.Second)

	watcher.Stop()

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

	// Make sure the description is set
	if !strings.Contains(receivedFileEvent.Description, receivedFileEvent.FilePath){
		t.Errorf("FileEvent description should be set and contain the filepath")
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