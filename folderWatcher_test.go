package folderWatcher

import (
	"math"
	"testing"
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