package folderWatcher

import (
	"os"
	"path/filepath"
	"testing"
)





func TestIsValidPath(t *testing.T) {
	// Try an invalid folder path
	invalidFolderPath:= "whrwekhr"
	isValid := IsValidPath(invalidFolderPath)
	if isValid{
		t.Errorf("%s is not a valid path", invalidFolderPath)
	}

	// try with the test folder
	validFolderPath := testFolderPath
	isValid = IsValidPath(validFolderPath)
	if !isValid {
		t.Errorf("%s is a valid folder path", validFolderPath)
	}

	newFilePath := randomizedFilePath(filepath.Join(testFolderPath, "tempfile#.txt"))
	// the new file path should not be valid yet.
	if isValid=IsValidPath(newFilePath); isValid{
		t.Errorf("%s is not a valid path", newFilePath)
	}

	writeToFile(newFilePath, "empty file")

	// the new file path should be valid now
	if isValid=IsValidPath(newFilePath); !isValid{
		t.Errorf("%s is a valid path", newFilePath)
	}

	// clean up the temp file
	removeFiles(true, newFilePath)
}

func Test_isValidDirPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"empty path", "", false},
		{"just a slash", "/", true},
		{"fake path", "/thing", false},
		{"dot", ".", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidDirPath(tt.path); got != tt.want {
				t.Errorf("isValidDirPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func assertFileInMap(t *testing.T, fileList map[string]os.FileInfo, filenames ...string){
	for _,filename:= range filenames{
		_, found := fileList[filename]
		if !found{
			t.Errorf("%s should be in the fileList map, but it is not.", filename)
		}
	}
}

// Tests to make sure that the GetFileList function finds the correct files
func TestGetFileList(t *testing.T) {
	tests := []struct {
		name         string
		recursive  bool
		showHidden bool
		wantFileCount int
		folderPath string
		wantErr      bool
	}{
		{name: "invalid path", recursive: false, showHidden: false, wantFileCount: 0, folderPath: "fakepath", wantErr: true},
		{name: "root", recursive: false, showHidden: false, wantFileCount: 2, folderPath: testFolderPath, wantErr: false},
		{name: "just sub folder", recursive: false, showHidden: false, wantFileCount: 2, folderPath: testSubFolder, wantErr: false},
		{name: "just sub folder with hidden", recursive: false, showHidden: true, wantFileCount: 3, folderPath: testSubFolder, wantErr: false},
		{name: "root recursive", recursive: true, showHidden: false, wantFileCount: 4, folderPath: testFolderPath, wantErr: false},
		{name: "root recursive with hidden", recursive: true, showHidden: true, wantFileCount: 6, folderPath: testFolderPath, wantErr: false},
	}
	// prepare test files
	rootFilePaths := createTestFiles(testFolderPath, 2)
	subFilePaths := createTestFiles(testSubFolder, 2)
	rootHiddenFilePaths, _ := createHiddenTestFiles(testFolderPath, 1)
	subHiddenFilePaths, _ := createHiddenTestFiles(testSubFolder, 1)
	// clean up files
	defer removeFiles(false, rootFilePaths...)
	defer removeFiles(false, subFilePaths...)
	defer removeFiles(false, rootHiddenFilePaths...)
	defer removeFiles(false, subHiddenFilePaths...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFileList, err := GetFileList(tt.folderPath, tt.recursive, tt.showHidden)
			// test error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// test the file count
			gotFileCount:= len(gotFileList)
			if gotFileCount != tt.wantFileCount{
				t.Errorf("GetFileList() want %d files, got %d", tt.wantFileCount, gotFileCount)
			}
		})
	}
}