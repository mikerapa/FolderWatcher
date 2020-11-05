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
	deleteFile(normalFilePath)
	deleteFile(hiddenFilePath)
	deleteFile(normalFilePath2)
	deleteFile(hiddenFilePath2)

}


func writeToFile(path string, fileContent string){
	f1, _ := os.Create(path)
	defer f1.Close()
	f1.WriteString(fileContent)
}

func deleteFile(filePath string){
	// as a safety measure, only delete a file with a txt extension
	if path.Ext(filePath) == ".txt" {
		os.Remove(filePath)
	} else {
		// should not have reached this point. Cannot delete a file with an extension other than txt.
		fmt.Printf("attempted to delete %s (extension %s) but it was prevented", filePath, path.Ext(filePath))
	}
}