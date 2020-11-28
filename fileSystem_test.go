package folderWatcher

import (
	"os"
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
	validFolderPath := "testFolder"
	isValid = IsValidPath(validFolderPath)
	if !isValid {
		t.Errorf("%s is a valid folder path", validFolderPath)
	}

	newFilePath := randomizedFilePath("testFolder/tempfile#.txt")
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

func Test_GetFileList(t *testing.T) {

	// set up the test folder
	setupTestFiles()
	defer tearDownTestFiles()

	// Try finding files recursive, without hidden files
	fileListRecursiveNotHidden, err := GetFileList("testFolder",true, false )
	if err != nil {
		t.Errorf("GetFileList: %s", err.Error())
	}
	if len(fileListRecursiveNotHidden) != 2{
		t.Errorf("GetFileList() should have found 2 files, found %d", len(fileListRecursiveNotHidden))
	}

	// make sure the correct files are in the collection
	assertFileInMap(t, fileListRecursiveNotHidden, normalFilePath2, normalFilePath)

	// Try finding files not recursive, without hidden files
	fileListNonRecursiveNoHidden, err := GetFileList("testFolder",false, false )
	if err != nil {
		t.Errorf("GetFileList: %s", err.Error())
	}
	if len(fileListNonRecursiveNoHidden) != 1{
		t.Errorf("GetFileList() should have found 1 file, found %d", len(fileListNonRecursiveNoHidden))
	}
	// make sure the correct files are in the collection
	assertFileInMap(t, fileListNonRecursiveNoHidden, normalFilePath)

	// Try finding files recursive with hidden files
	fileListRecursiveHidden, err :=  GetFileList("testFolder", true, true)
	if len(fileListRecursiveHidden)!=4{
		t.Errorf("GetFileList() should have 4 files, found %d", len(fileListRecursiveHidden))
	}
	// make sure the correct files are in the map
	assertFileInMap(t, fileListRecursiveHidden, normalFilePath, normalFilePath2, hiddenFilePath, hiddenFilePath2)

	// TRy finding files non-recursive with hidden files
	fileListNonRecursiveHidden, err :=  GetFileList("testFolder", false, true)
	if len(fileListNonRecursiveHidden)!=2{
		t.Errorf("GetFileList() should have 2 files, found %d", len(fileListNonRecursiveHidden))
	}
	// make sure the correct files are in the map
	assertFileInMap(t, fileListNonRecursiveHidden, normalFilePath, hiddenFilePath)

}