package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

// Watch files and notify all connected clients when a file changes.
//
// if the client cant connect to the server, it should watch the files in the directory and send the files to the server when the server is back online
func watchFiles(path string) {
	client, _ := connectMongo()
	watcher, err := fsnotify.NewWatcher()
	collection := client.Database("sync").Collection("server")
	documents, _ := getAllDocuments(collection)

	if err != nil {
		return
	}

	// deletedCount, _ := deleteAllDocuments(collection)
	// fmt.Printf("Deleted %d documents\n", deletedCount)
	for _, doc := range documents {
		fmt.Println("documents:", doc)
	}

	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if (event.Op == fsnotify.Create) && validFileExtension(event.Name) {
					log.Println("valid event location:", event)
					fileID := createFile(collection, event.Name)

					createFileVersion(collection, fileID)

				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Modified or created file:", event.Name)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Watching directory: %s\n", path)
	<-make(chan struct{})
}
