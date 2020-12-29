# Folder Watcher

## Overview
FolderWatcher is a Go library for tracking file events in specified folders. If you are looking for a command 
line application for tracking file changes, see https://github.com/mikerapa/GoFileWatcher. 

Supported file events
- Add - creation of a new file in a watched folder
- Remove - deletion of a file from a watched folder
- Write - update of a file in a watched folder
- Move - change a file path from one watched folder to another 


## Getting started

### Import Folder Watcher 
`import ("github.com/mikerapa/FolderWatcher")`

### Create a FolderWatcher instance 
Use the New() function to get a new instance of the FolderWatcher. 

`	watcher := folderWatcher.New()`
   
Use the AddFolder function to provide a new folder path. If the path is invalid an error will be returned.
   
   ` err := watcher.AddFolder("../testFolder", true, false)`
   
### Collect FileEvents from the FileChanged channel
When the FolderWatcher is running, it sends data through following channels:
1. Stopped Channel - FolderWatcher will send a `true` to this channel when the WatcherState changes to `Stopped`.  
2. FileChanged - FolderWatcher passes file events through this channel. 


```	
// collect events from the watcher
go func () {
    for {
        select{
            case <- watcher.Stopped:
                println("Got the stopped message")
                return
            case fe:= <- watcher.FileChanged:
                println(fe.FileChange.ToString(), fe.FilePath)
        }
    }
}()
```
	
### Starting and Stopping the FolderWatcher
Use the `watcher.Start()` function to start an instance of the watcher. You must have something to receive data on the 
channels before starting the watcher, otherwise you'll likely get an error related to deadlocks. 

Call `watcher.Stop()` to stop the watcher. Note that stopping the watcher does not remove the list of folders currently 
being watched. 

## Examples 
See sample/main.go for a basic working example. 

There is a full application built with this library here: https://github.com/mikerapa/GoFileWatcher.

## FolderWatcher Struct

#### Interval (int)
This is the rate at which the FolderWatcher polls the file system for changes. The FolderWatcher is setting this value 
automatically based on the number of files currently being watched. The value should be an integer between 500 and 5000
and this value is the number of milliseconds the watcher will wait before starting another cycle. 

#### Stopped Channel (chan bool)
FolderWatcher will send a `true` to this channel when the WatcherState changes to `Stopped`. 

#### FileChanged (chan bool)
The FolderWatcher passes file events through this channel. 

#### WatcherState (int)

Value indicating the status of the watcher

1. NotStarted WatcherState = 1
2. Running WatcherState = 2
3. Stopped WatcherState =3 

#### RequestedWatches (map[string]WatchRequest)
Map containing all watched folders. The key is the folder path. 

#### New

`func New() Watcher `

Creates and returns a new instance of a FolderWatcher with default settings.

Input Parameters: None

Return Values

| Type | Description | 
| ------------| ------- |
| FolderWatcher| An instance of a folder watcher|  

#### AddFolder
`func (w *Watcher) AddFolder(path string, recursive bool, showHidden bool) (err error)`

Add a folder path to the list of folders to watch. 

Input Parameters 

| Parameter | Type | Description |
| ----------- | ----------- | ----------- |
| path | string | path of the folder to watch |
| recursive | boolean | determines if subfolders will also be watched |
| showHidden | boolean | determines if hidden files should be watched |

Return Values

| Type | Description |
| ----------- | ----------- | 
| error | An error is returned if the AddFolder function failed for any reason. If an error is returned, the caller should assume that the folder was not added. If a nil is returned, the folder was added. |

#### RemoveFolder

`func (w *Watcher) RemoveFolder(path string, returnErrorIfNotFound bool) ( err error){`

Input Parameters 

| Parameter | Type | Description |
| ----------- | ----------- | ----------- |
| path | string | path of the folder to stop watching |
| returnErrorIfNotFound | boolean | determines if this function should return an error if the path does not match a folder currently being watched. |


Return Values

| Type | Description |
| ----------- | ----------- | 
| error | An error is returned if the RemoveFolder function failed for any reason. | 


#### Stop
`func (w *Watcher) Stop()`

Input Parameters: None

Return Values: None 

Sets the watcher WatcherState to `Stopped` and discontinues polling the file system for changes. 


#### Start

`func (w *Watcher) Start()`

Sets the WatcherState to `Running` and begins polling the file system for changes

Input Parameters: None

Return Values: None 


## FileEvent Struct 

#### FileChange (int32)

A value indicating which file change occurred 

	0. Add FileChange = 0
	1. Remove FileChange = 1
	2. Write FileChange =2
	3. Move FileChange =3
	
#### FilePath (string)

Full path of the file

#### PreviousPath (string)

The previous path, if the file path has changed. This field will only have a value when the FileChanged is 3, indicating a file move. 

#### Description (string)

A string representing the file change. See examples below.

> ../testFolder/subFolder/file1.txt was added
>
> File was moved from ../testFolder/subFolder/file1.txt to ../testFolder/subFolder/file2.txt
>
>../testFolder/subFolder/file2.txt was updated
>
>../testFolder/subFolder/file2.txt was removed
 

## Known limitations
1. Files moved from a watched folder to an unwatched folder will be recorded as a Remove event. 
2. Changes to the file metadata, such as chmod, may be not be captured as a Write event. Depending
 on your operating system the datetime stamp may not be updated.  
3. Extremely rapid file events may be missed. This library uses a polling technique for detecting changes
in the file system. If there are successive modifications in a time span less than the polling interval, 
it's possible that FolderWatcher make not report all of the events. 
4. On Windows, some file Move events will be recorded as an Add followed by a Remove. 
 
 
 ## Future feature development 
 - [ ] Pause - ability to temporarily discontinue file events, but not stop the watcher
 - [ ] Clear - remove all watched folders and watched files
 - [ ] Override Interval - the cycle interval (rate at which the file system is polled for changes) is automatically set based 
 on the number of files being watched. There is no feature to override that value.
   
## Version History
### v1.0
 - Initial version 
### v1.1 
 - Corrected concurrency error. See Issue #2. 
 - Made changes to improve the performance of each iteration 
 - Changed iteration processing to re-calculate the interval time with each iteration. 
