package folderWatcher

import "fmt"

//
type FileEvent struct {
	FileChange
	FilePath string
	PreviousPath string
	Description string

}

// FileChange enumeration
type FileChange int32
const (
	Add FileChange = 0
	Remove FileChange = 1
	Write FileChange =2
	Move FileChange =3
)

func (fc FileChange) String() string {
	fileChangeStrings:= [...]string{"Add", "Remove", "Write", "Move"}
	return fileChangeStrings[fc]
}

// return a string representation of the event value
func (fc *FileChange) ToString() string {return fmt.Sprintf("%s", fc)}

