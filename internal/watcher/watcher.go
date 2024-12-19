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
