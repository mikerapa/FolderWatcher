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
	//stopChan := make (chan bool)
	//stoppedChan :=make (chan bool)
	//
	//watcher.Stop = stopChan
	//watcher.Stopped = stoppedChan

	go func () {
		for {
			select{
				case <- watcher.Stopped:
					println("Got the stopped message")
					wg.Done()
					return
			}
		}
	}()

	println("Calling start")
	wg.Add(1)
	watcher.Start()
	go func() {
		time.Sleep(15 * time.Second)
		println("Sending the stop message")
		watcher.Stop<-true
	}()

	wg.Wait()
}
