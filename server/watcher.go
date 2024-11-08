package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

// the watcher should be a web server that handle client requests to then update the client version of the file
// NOTE: this only listens to events when a client is connected
// which for obvious reasons is dumb
//
// TODO: there needs to be a function that handles file differences
// when connected
//
// TODO: the dirWatcher has to be running at all times
// And if there is a client connection it can be handled by external calls
// The watcher needs to => write to the database and then emit the changes
// to any clients that are listening
func dirWatcherWS(ws *websocket.Conn) { // Create a new watcher
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

				if err := ws.WriteMessage(websocket.TextMessage, []byte(event.Name)); err != nil {
					return
				}
				// NOTE: this is only checking for write events in the file system
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
	//
	//
	path, _ := filepath.Abs(fmt.Sprintf("./%s", directory))
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Watching directory: %s\n", path)
	// Block main goroutine forever
	<-make(chan struct{})
}

// / mongoDB connection should be passed in
func dirWatcher() {
	// Create a new watcher
	client, _ := connectMongo()
	watcher, err := fsnotify.NewWatcher()
	collection := client.Database("sync").Collection("server")
	documents, _ := getAllDocuments(collection)
	//
	//j

	// deletedCount, _ := deleteAllDocuments(collection)
	// fmt.Printf("Deleted %d documents\n", deletedCount)

	if err != nil {
		log.Fatal(err)
	}

	// Print the documents
	for _, doc := range documents {
		fmt.Println(doc)
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
				// NOTE: when a file is created its not going to have a value

				if (event.Op == fsnotify.Create) && validFileExtension(event.Name) {
					log.Println("valid event location:", event)
					createFile(collection, event.Name)
				}

				// if ()

				// if err := ws.WriteMessage(websocket.TextMessage, []byte(event.Name)); err != nil {
				// 	return
				// NOTE: this is only checking for write events in the file system
				// }

				if (event.Op&fsnotify.Write == fsnotify.Write) && validFileExtension(event.Name) {
					log.Println("modified file:", event.Name)

					file, err := findFile(collection, event.Name)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Printf("Found file: %+v\n", file)
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
	//
	//
	path, _ := filepath.Abs(fmt.Sprintf("./%s", directory))
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Watching directory: %s\n", path)
	// Block main goroutine forever
	<-make(chan struct{})
}
