package folderWatcher

import (
	"fmt"
	"os"
	"path"
)

const(
	normalFilePath = "testFolder/testFile.txt"
	hiddenFilePath = "testFolder/.hiddenTestFile.txt"
	normalFilePath2 = "testFolder/subFolder/testFile2.txt"
	hiddenFilePath2= "testFolder/subFolder/.hiddenTestFile2.txt"
)

func setupTestFiles(){
	// set up the test folder
	writeToFile(normalFilePath, "nothing")
	writeToFile(hiddenFilePath, "empty file")
	writeToFile(normalFilePath2, "nothing")
	writeToFile(hiddenFilePath2, "empty file")

}

func tearDownTestFiles(){
	// remove test files from the test folder
	removeFiles(true, normalFilePath, hiddenFilePath, normalFilePath2)

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

