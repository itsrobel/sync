package watcher

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

type WatcherNotification interface {
	HandleChange(path string, eventType string) error
}

type FileWatcher struct {
	watcher  *fsnotify.Watcher
	handlers []WatcherNotification
}

func NewFileWatcher() (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatcher{
		watcher:  watcher,
		handlers: make([]WatcherNotification, 0),
	}, nil
}

func (fw *FileWatcher) Watch(path string) error {
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-fw.watcher.Events:
				for _, handler := range fw.handlers {
					eventType := getEventType(event.Op)
					handler.HandleChange(event.Name, eventType)
				}
			case err := <-fw.watcher.Errors:
				log.Printf("Error: %v", err)
			}
		}
	}()

	err := fw.watcher.Add(path)
	if err != nil {
		return err
	}
	<-done
	return nil
}

func getEventType(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "CREATE"
	case op&fsnotify.Write == fsnotify.Write:
		return "WRITE"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "REMOVE"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "RENAME"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "CHMOD"
	default:
		return "UNKNOWN"
	}
}

func (fw *FileWatcher) AddHandler(handler WatcherNotification) {
	fw.handlers = append(fw.handlers, handler)
}

// Watch files and notify all connected clients when a file changes.
// if the client cant connect to the server, it should watch the files in the directory and send the files to the server when the server is back online
// func WatchFiles(path string) {
// 	client, _ := datacontroller.ConnectMongo()
// 	watcher, err := fsnotify.NewWatcher()
// 	collection := client.Database("sync").Collection("server")
// 	documents, _ := datacontroller.GetAllDocuments(collection)
//
// 	if err != nil {
// 		return
// 	}
//
// 	// deletedCount, _ := deleteAllDocuments(collection)
// 	// fmt.Printf("Deleted %d documents\n", deletedCount)
// 	for _, doc := range documents {
// 		fmt.Println("documents:", doc)
// 	}
//
// 	defer watcher.Close()
//
// 	go func() {
// 		for {
// 			select {
// 			case event, ok := <-watcher.Events:
// 				if !ok {
// 					return
// 				}
//
// 				if (event.Op == fsnotify.Create) && datacontroller.ValidFileExtension(event.Name) {
// 					log.Println("valid event location:", event)
// 					fileID := datacontroller.CreateFile(collection, event.Name)
//
// 					datacontroller.CreateFileVersion(collection, fileID)
//
// 				}
//
// 				if event.Op&fsnotify.Write == fsnotify.Write {
// 					log.Println("Modified or created file:", event.Name)
// 				}
//
// 			case err, ok := <-watcher.Errors:
// 				if !ok {
// 					return
// 				}
// 				log.Println("Error:", err)
// 			}
// 		}
// 	}()
//
// 	err = watcher.Add(path)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("Watching directory: %s\n", path)
// 	<-make(chan struct{})
// }
