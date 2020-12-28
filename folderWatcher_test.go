package folderWatcher

import (
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// template func for all test functions
func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	println("In TestMain")
	code := m.Run()
	os.Exit(code)
}

// Make sure the New function returns a valid watcher
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
		{name: "add a valid path", path: testFolderPath, recursive: false, showHidden:false, wantAdd: true, wantErr: false},
	}
	// populate the test folder with test files
	fileList:= createTestFiles(testFolderPath,1)
	defer removeFiles(false, fileList...)

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
		{name: "folder path in the list", path:testFolderPath, returnErrorIfNotFound: false, shouldRemoveFolder: true},
	}

	// populate the test folder with test files
	fileList:= createTestFiles(testFolderPath,1)
	defer removeFiles(false, fileList...)

	for _, tt := range tests {
		watcher := New()
		err:=watcher.AddFolder(testFolderPath, true, false)
		if err!=nil{
			t.Error(err.Error())
		}

		t.Run(tt.name, func(t *testing.T) {
			// make sure an error is returned if it should
			if err = watcher.RemoveFolder(tt.path, tt.returnErrorIfNotFound); (err != nil) != tt.returnErrorIfNotFound {
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

func TestWatcher_MultipleFolderRemove(t *testing.T) {
	// this test should cover situation in which there are multiple folders requested for watch and one is removed.

	// create some files in the subfolders
	testFiles := append(createTestFiles(testSubFolder, 2), createTestFiles(testSubFolder2, 2)...)


	// set up the watcher
	watcher := New()
	_ = watcher.AddFolder(testSubFolder, false, false)
	_ = watcher.AddFolder(testSubFolder2, false, false)

	// get data from channels
	go func() {
		for {
			select {
				case <- watcher.FileChanged:
					// No need to do anything with the event for this test
				case <- watcher.Stopped:
					return
			}
		}
	}()

	// start the watcher and wait for a cycle to complete
	watcher.Start()
	defer watcher.Stop()
	defer removeFiles(false, testFiles...)
	time.Sleep(1 * time.Second)

	// there should be 4 files watched
	if len(watcher.watchedFiles)!=4{
		t.Errorf("Watcher is not watching the correct number of files. Want 4, got %d", len(watcher.watchedFiles))
	}

	// remove a folder and there should be 2 files watched
	err := watcher.RemoveFolder(testSubFolder2, true )
	if err!=nil {
		// if there's an error returned from RemoveFolder, the test should fail
		t.Error(err.Error())
	}

	if len(watcher.watchedFiles) != 2{
		t.Errorf("Watcher is not watching the correct number of files. Want 2, got %d", len(watcher.watchedFiles))
	}


	// Remove the one remaining folder and there should be no files watched
	err = watcher.RemoveFolder(testSubFolder, true )
	if err!=nil {
		// if there's an error returned from RemoveFolder, the test should fail
		t.Error(err.Error())
	}

	if len(watcher.watchedFiles) != 0{
		t.Errorf("Watcher is not watching the correct number of files. Want 0, got %d", len(watcher.watchedFiles))
	}

}

// Test to make sure the watcher cannot be started more than once.
func TestStartTwice(t *testing.T){
	watcher := New()
	err:= watcher.AddFolder(testSubFolder, false, false)
	if err!=nil{
		t.Error(err.Error())
	}
	var receivedEventCount = 0
	const testFileCount = 2

	// collect events from the watcher
	go func () {
		for {
			select{
			case <- watcher.Stopped:
				return
			case newFileEvent := <- watcher.FileChanged:
				if newFileEvent.FileChange == Add{
					receivedEventCount++
				}
			}
		}
	}()
	watcher.Start()
	watcher.Start() // attempt to start the watcher again

	testFiles := createTestFiles(testSubFolder, testFileCount)
	time.Sleep(1 * time.Second)
	// make sure the correct number of adds were received
	if len(testFiles) != receivedEventCount {
		t.Errorf("should have received %d Add events, got %d", len(testFiles), receivedEventCount)
	}
	watcher.Stop()
	removeFiles(false, testFiles...)
}

// Test stopping and restarting the watcher
func TestRestartWatcher(t *testing.T) {
	var testFiles  []string


	const iterations =3
	watcher := New()
	err := watcher.AddFolder(testSubFolder, false, false)
	if err!=nil{
		t.Error(err.Error())
	}

	// collect events from the watcher
	receivedEvents := make(map[FileChange]int)
	go func () {
		for {
			select{
			case <- watcher.Stopped:
			case event:= <- watcher.FileChanged:
				// keep record of each event received
				receivedEvents[event.FileChange] = receivedEvents[event.FileChange]+1
			}
		}
	}()


	// create and update each of the files
	for i:=0; i<iterations; i++{
		watcher.Start()
		time.Sleep(1 * time.Second)
		newFiles := createTestFiles(testSubFolder, 1)
		testFiles = append(testFiles, newFiles[0])
		time.Sleep(1 * time.Second)
		writeToFile(newFiles[0], "new data")
		time.Sleep(1 * time.Second)
		watcher.Stop()
		time.Sleep(1 * time.Second)
	}

	defer removeFiles(false, testFiles...)
	// make sure the correct number of events were received
	var eventType FileChange
	for _, eventType= range []FileChange{Add, Write}{
		if receivedEvents[eventType]!=len(testFiles){
			t.Errorf("should have received %d %s events, got %d", len(testFiles), eventType, receivedEvents[eventType] )
		}
	}

}

// Test the watcher with two folders requested
func TestMultipleWatchRequests(t *testing.T){
	watcher := New()
	_ = watcher.AddFolder(testSubFolder, false, false)
	_ = watcher.AddFolder(testSubFolder2, false, false)

	receivedEvents := make(map[FileChange]int)

	watcher.Start()
	// collect events from the watcher
	go func () {
		for {
			select{
			case <- watcher.Stopped:
				return
			case evnt:= <- watcher.FileChanged:
				receivedEvents[evnt.FileChange] = receivedEvents[evnt.FileChange]+1
			}
		}
	}()

	// create test files, edit them and remove
	time.Sleep(1 * time.Second)
	newFiles1 := createTestFiles(testSubFolder, 1)
	newFiles2 := createTestFiles(testSubFolder2, 1 )
	time.Sleep(1 * time.Second)
	writeToFile(newFiles1[0], "updated")
	writeToFile(newFiles2[0], "updated data")
	time.Sleep(1 * time.Second)
	removeFiles(false, append(newFiles1, newFiles2...)...)


	time.Sleep(2 * time.Second)
	// make sure the correct number of events were received
	var eventType FileChange
	for _, eventType= range []FileChange{Add, Remove, Write}{
		if receivedEvents[eventType]!=2{
			t.Errorf("should have received %d %s events, got %d", 2, eventType, receivedEvents[eventType] )
		}
	}

	watcher.Stop()

}

func TestReCalculateInterval(t *testing.T){
	watcher := New()
	_ = watcher.AddFolder(testFolderPath, false, false)
	newFilePaths := createSequentialTestFiles(testFolderPath, 2000)

	watcher.Start()
	// collect events from the watcher
	go func () {
		for {
			select{
			case <- watcher.Stopped:
				return
			case <- watcher.FileChanged:
			}
		}
	}()

	time.Sleep(2* time.Second)

	if watcher.Interval != 1000{
		t.Errorf("Interval should be 1000, got %d", watcher.Interval )
	}
	time.Sleep(1 * time.Second)
	watcher.Stop()

	removeFiles(false, newFilePaths...)
}

func TestAddFileEvent(t *testing.T) {
	watcher := New()
	_ = watcher.AddFolder(testFolderPath, false, false)
	var receivedFileEvent FileEvent

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

	watcher.Start()
	time.Sleep(1 * time.Second)
	newTestFiles := createTestFiles(testFolderPath, 1)
	defer removeFiles(true, newTestFiles...)

	time.Sleep(1 * time.Second)
	watcher.Stop()

	// make sure the current type of FileEvent was received
	if receivedFileEvent.FileChange != Add {
		t.Errorf("Watcher did not send Add FileEvent to the FileChanged channel")
	}

	// make sure the path was included in the FileEvent
	absTestFilePath, _ := filepath.Abs(newTestFiles[0])
	if receivedFileEvent.FilePath != absTestFilePath {
		t.Errorf("Filepath was not included in the FileEvent")
	}

	// make sure the new file is in the watchedFiles
	assertFileInMap(t, watcher.watchedFiles, absTestFilePath)

	// Make sure the description is set
	if !strings.Contains(receivedFileEvent.Description, receivedFileEvent.FilePath){
		t.Errorf("FileEvent description should be set and contain the filepath")
	}
}


func TestMoveFileEvent(t *testing.T) {
	// create two file paths, a before and after
	firstPath := randomizedFilePath(filepath.Join(testFolderPath, "testfile#.txt"))
	secondPath := randomizedFilePath(filepath.Join(testFolderPath, "testfile#.txt"))
	writeToFile(firstPath, "nothing")

	// set up the watcher
	watcher := New()
	_ = watcher.AddFolder(testFolderPath, true, false)

	// clean up the files that were used in this test
	defer removeFiles(false, firstPath, secondPath)

	// start the watcher and collect events
	receivedEvents := make(map[FileChange]int)
	var receivedMoveEvent FileEvent

	watcher.Start()
	// collect events from the watcher
	go func () {
		for {
			select{
			case <- watcher.Stopped:
				return
			case event := <- watcher.FileChanged:
				// keep a count of the file events
				receivedEvents[event.FileChange] = receivedEvents[event.FileChange]+1

				// if there's a move event, keep it in a variable to examine
				if event.FileChange == Move{
					receivedMoveEvent = event
				}
			}
		}
	}()

	// Move the file and give some time to get the FileEvent on the channel
	time.Sleep(1 * time.Second)
	moveFile(firstPath, secondPath)
	time.Sleep(2 * time.Second)

	if runtime.GOOS == "windows"{
		// on windows, this file move will be represented as an Add and a Remove
		if receivedEvents[Add]!=1{
			t.Errorf("should have received 1 Add events, got %d", receivedEvents[Add] )
		}
		if receivedEvents[Remove]!=1{
			t.Errorf("should have received 1 Remove events, got %d", receivedEvents[Remove] )
		}
	} else {
		// on non-Windows operating systems, there should be a Move event
		if receivedEvents[Move]!=1{
			t.Errorf("should have received 1 Move events, got %d", receivedEvents[Move] )
		}

		// make sure the path and previous path are included in the FileEvent
		if receivedMoveEvent.FilePath != AbsPath(secondPath){
			t.Errorf("Filepath is not correct in FileEvent. Wanted %s, got %s", AbsPath(secondPath), receivedMoveEvent.FilePath)
		}
		if receivedMoveEvent.PreviousPath!=AbsPath(firstPath){
			t.Errorf("FileEvent should have PreviousPath value set for a Move event. Wanted '%s', got '%s'", AbsPath(firstPath), receivedMoveEvent.PreviousPath)
		}

		// Make sure the description is set
		if !strings.Contains(receivedMoveEvent.Description, receivedMoveEvent.FilePath){
			t.Errorf("FileEvent description should be set and contain the filepath")
		}

	}

	// Make sure the new path is in the watchedFiles map
	_, fileFound := watcher.watchedFiles[AbsPath(secondPath)]
	if !fileFound{
		t.Errorf("file %s should be in the list of watched files", secondPath)
	}

	watcher.Stop()
}

// This test should cover updates to an existing which is being watched
func TestWriteFileEvent(t *testing.T) {
	// set up the watcher and a file
	testFilePath := createTestFiles(testFolderPath, 1)[0]
	watcher := New()
	_= watcher.AddFolder(testFolderPath, false, false)

	// get the mod time for the newly created file. This will be used to for a comparison.
	absTestFilePath, _ := filepath.Abs(testFilePath)
	testFile,_ :=watcher.watchedFiles[absTestFilePath]
	initialModTime := testFile.ModTime()

	defer removeFiles(true, testFilePath)

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
	writeToFile(testFilePath, "updated file content")
	time.Sleep(2 * time.Second)



	// make sure the correct FileEvent was received
	if receivedFileEvent.FileChange != Write {
		t.Errorf("Watcher did not send Write FileEvent to the FileChanged channel")
	}

	// make sure the path is included in the FileEvent
	if receivedFileEvent.FilePath != absTestFilePath{
		t.Errorf("Filepath was not included in the FileEvent")
	}

	// make sure the updated file remains in the watched files list
	newWatchedFile, fileFound := watcher.watchedFiles[absTestFilePath]
	if !fileFound{
		t.Errorf("file %s should have remained in the list of watched files", testFilePath)
	}

	// compare the mod time of the file in the watchedFiles list. It should be different that the original value
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
	_= watcher.AddFolder(testFolderPath, false, false)
	testFilePath:= createTestFiles(testFolderPath,1)[0]
	defer removeFiles(false, testFilePath)

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
	removeFiles(true, testFilePath)
	time.Sleep(1 * time.Second)

	watcher.Stop()

	// make sure the correct FileEvent was received
	if receivedFileEvent.FileChange != Remove {
		t.Errorf("Watcher did not send Remove FileEvent to the FileChanged channel")
	}

	// make sure the path is included in the FileEvent
	if receivedFileEvent.FilePath != AbsPath(testFilePath){
		t.Errorf("Filepath was not included in the FileEvent")
	}

	// make sure the removed file is no longer being watched
	_, fileFound := watcher.watchedFiles[testFilePath]
	if fileFound{
		t.Errorf("file was not removed with the Remove FileEvent path=%s", testFilePath)
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
			calculatedValue :=  calculateInterval(tt.watchedFileCount)
			if calculatedValue< MinimumIntervalTime || calculatedValue > MaximumIntervalTime{
				t.Errorf("calculateInterval() = %v is out of range. Min:%v, Max:%v", calculatedValue, MinimumIntervalTime, MaximumIntervalTime)
			}
		})
	}
}