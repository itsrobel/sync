package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// func main() {
// 	watcher()
// }

// func websocket() {
//
//
// }

// what else do I need the watcher to do?
// the watcher should be a web server that handle client requests to then update the client version of the file
func folderwatcher() {
	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start listening for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// Add a path to watch
	path, _ := filepath.Abs("./content")
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Watching directory: %s\n", path)
	// Block main goroutine forever
	<-make(chan struct{})
}
