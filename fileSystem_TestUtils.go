package folderWatcher

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	)



const (
	testFolderPath = "testFolder")

var(
	testSubFolder2 = filepath.Join("testFolder","subFolder2")
	testSubFolder = filepath.Join("testFolder","subFolder")
)



func createTestFiles(folderPath string, count int) (fileList []string){
	filePathTemplate := filepath.Join(folderPath, "testfile#.txt")
	for i:=1; i<=count; i++{
		newFilePath := randomizedFilePath(filePathTemplate)
		rndNum := rand.Intn(1000)
		writeToFile(newFilePath, fmt.Sprintf("%d", rndNum) )
		fileList = append(fileList, newFilePath)
	}

	return
}

// Creates test files without any randomization
// this method should be used for testing only.
func createSequentialTestFiles(folderPath string, count int) (fileList []string){
	filePathTemplate := filepath.Join(folderPath, "seqtestfile%d.txt")
	for i:=1; i<=count; i++{
		newFilePath := fmt.Sprintf(filePathTemplate, i)
		writeToFile(newFilePath, "file from createSequentialTestFiles" )
		fileList = append(fileList, newFilePath)
	}

	return
}


func createHiddenTestFiles(folderPath string, count int ) (fileList []string, err error ){
	var filePathTemplate string
	// if this is running on a non-Windows OS, put the dot in front of the name to make the file
	// hidden.
	if runtime.GOOS == "windows"{
		filePathTemplate = filepath.Join(folderPath, "testfile#.txt")
	} else {
		filePathTemplate = filepath.Join(folderPath, ".testfile#.txt")
	}

	// create test files with random names and content
	for i:=1; i<=count; i++{
		newFilePath := randomizedFilePath(filePathTemplate)
		writeToFile(newFilePath, fmt.Sprintf("%d", rand.Intn(1000)) )
		err = hideFile(newFilePath)
		if err != nil{ break}
		fileList = append(fileList, newFilePath)
	}

	return
}

func randomizedFilePath(pathTemplate string) string {
	pathTemplate = strings.Replace(pathTemplate, "#", "%d", 1)
	return fmt.Sprintf(pathTemplate, rand.Intn(1000))
}


func closeFiles(files ...*os.File){
	for _, file:= range files{
		_ = file.Close()
	}
}

func writeToFile(path string, fileContent string){
	f1, _ := os.Create(path)
	defer closeFiles(f1)
	_, err := f1.WriteString(fileContent)

	if err!=nil{
		// if there is any error while writing to a file, all tests should stop.
		panic(err.Error())
	}
}

func moveFile(currentPath string, newPath string){
	err := os.Rename(currentPath, newPath)
	if err!=nil{
		// if there's an error while trying to move the file, the test should not continue
		panic(err.Error())
	}
}

func removeFiles(panicOnFailure bool, filePaths ... string){
	for _, thisFilePath := range filePaths{
		// as a safety measure, only delete a file with a txt extension
		if path.Ext(thisFilePath) == ".txt" {
			err := os.Remove(thisFilePath)
			if err!=nil && panicOnFailure{
				// if there is an error when trying to delete a file, the test should not continue
				panic(err.Error())
			}
		} else {
			// should not have reached this point. Cannot delete a file with an extension other than txt.
			fmt.Printf("attempted to delete %s (extension %s) but it was prevented", thisFilePath, path.Ext(thisFilePath))
		}
	}
}

