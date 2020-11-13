package main

import (
	"FolderWatcher"
	"sync"
	"time"
)

func main() {
	wg := sync.WaitGroup{}
	watcher := folderWatcher.New()
	err := watcher.AddFolder("../testFolder", true, false)
	if err != nil {
		panic(err.Error())
	}

	err = watcher.AddFolder("../testFolder/subFolder2/", true, false)
	if err != nil {
		panic(err.Error())
	}


	// collect events from the watcher
	go func () {
		for {
			select{
				case <- watcher.Stopped:
					println("Got the stopped message")
					wg.Done()
					return
				case fe:= <- watcher.FileChanged:
					println(fe.FileChange.ToString(), fe.FilePath)
			}
		}
	}()

	println("Calling start")
	wg.Add(1)
	watcher.Start()
	go func() {
		time.Sleep(300 * time.Second)
		println("Sending the stop message")
		watcher.Stop()
	}()

	wg.Wait()
}
