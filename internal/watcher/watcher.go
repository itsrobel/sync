package watcher

import (
	"database/sql"
	"fmt"
	"log"
	// "time"
	"io"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/itsrobel/sync/internal/sql_controller"
	"os"
	"strings"
)

type FileWatcher struct {
	watcher *fsnotify.Watcher
	db      *sql.DB
	done    chan struct{}
	wg      sync.WaitGroup
}

func InitFileWatcher(dbPath, watchPath string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}
	db, err := db_controller.ConnectSQLite(dbPath)
	files, _ := db_controller.GetAllFiles(db)
	log.Println(files)
	if files != nil {
		log.Println(db_controller.GetAllFileVersions(db, files[0].ID))
	}
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

	fw.done = make(chan struct{})
	fw.wg.Add(1)

	go func() {
		defer fw.wg.Done()
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
			case <-fw.done:
				return
			}
		}
	}()

	fmt.Printf("Watching directory: %s\n", path)
	// Block until done channel is closed
	<-fw.done
	return nil
}
func (fw *FileWatcher) handleEvent(event fsnotify.Event) error {
	// log.Printf("File Event: %s, File Location %s", event.Op, event.Name)
	if event.Op&fsnotify.Create == fsnotify.Create && db_controller.ValidFileExtension(event.Name) {
		log.Printf("New file create Event: %s", event.Name)
		isFile, _ := db_controller.FindFileByLocation(fw.db, event.Name)
		var fileID string
		if isFile == nil {
			log.Printf("Create new file at: %s", event.Name)
			fileID, _ = db_controller.CreateFile(fw.db, event.Name)
			// if err != nil {
			// 	return fmt.Errorf("failed to create file record: %w", err)
			// }
		} else {
			log.Printf("File exists at: %s", event.Name)
			fileID = isFile.ID
		}

		if err := db_controller.CreateFileVersion(fw.db, fileID, event.Name); err != nil {
			return fmt.Errorf("failed to create file version: %w", err)
		}
	}
	if event.Op&fsnotify.Write == fsnotify.Write && db_controller.ValidFileExtension(event.Name) {
		isFile, _ := db_controller.FindFileByLocation(fw.db, event.Name)
		file, err := os.Open(event.Name) // Note: using event.Name instead of "filename.txt"
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		var content strings.Builder
		buf := make([]byte, 8192) // Using 8KB buffer size
		for {
			n, err := file.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			content.Write(buf[:n])
		}

		if err := db_controller.CreateFileVersion(fw.db, isFile.ID, content.String()); err != nil {
			return fmt.Errorf("failed to create file version: %w", err)
		}
	}
	return nil
}

func (fw *FileWatcher) Stop() {
	if fw.done != nil {
		close(fw.done)
		fw.wg.Wait()
	}
}
