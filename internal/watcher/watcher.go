package watcher

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/itsrobel/sync/internal/sql_controller"
)

type FileWatcher struct {
	watcher *fsnotify.Watcher
	db      *sql.DB
}

func InitFileWatcher(dbPath, watchPath string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}
	return &FileWatcher{
		watcher:  watcher,
		handlers: make([]WatcherNotification, 0),
	}, nil
}

	db, err := db_controller.ConnectSQLite(dbPath)
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	fw := &FileWatcher{
		watcher: watcher,
		db:      db,
	}

	if err := fw.startWatching(watchPath); err != nil {
		watcher.Close()
		db.Close()
		return nil, err
	}

	return fw, nil
}

func (fw *FileWatcher) startWatching(path string) error {
	if err := fw.watcher.Add(path); err != nil {
		return fmt.Errorf("failed to add path to watcher: %w", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}
				if err := fw.handleEvent(event); err != nil {
					log.Printf("Error handling event: %v", err)
				}
			case err, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()

	fmt.Printf("Watching directory: %s\n", path)
	<-make(chan struct{})
	return nil
}

func (fw *FileWatcher) Close() error {
	if err := fw.watcher.Close(); err != nil {
		return fmt.Errorf("failed to close watcher: %w", err)
	}
	if err := fw.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event) error {
	if event.Op == fsnotify.Create && db_controller.ValidFileExtension(event.Name) {
		log.Printf("New file created: %s", event.Name)
		fileID, err := db_controller.CreateFile(fw.db, event.Name)
		if err != nil {
			return fmt.Errorf("failed to create file record: %w", err)
		}

		if err := db_controller.CreateFileVersion(fw.db, fileID); err != nil {
			return fmt.Errorf("failed to create file version: %w", err)
		}
	}

	if event.Op&fsnotify.Write == fsnotify.Write {
		log.Printf("Modified file: %s", event.Name)
	}
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
