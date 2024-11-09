package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

// Watch files and notify all connected clients when a file changes.
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
					createFile(collection, event.Name)
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
