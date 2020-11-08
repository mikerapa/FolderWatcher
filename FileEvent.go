package folderWatcher

//
type FileEvent struct {
	FileChange
	FilePath string

}

// return a string representation of the event value
func (fe *FileEvent) EventName() string {
	switch fe.FileChange {
	case Add:
		return "Add"
	case Remove:
		return "Remove"
	default:
		return "Unknown"

	}
}


// FileChange enumeration
type FileChange int32
const (
	Add FileChange = iota
	Remove
)

