package watcher

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	"github.com/itsrobel/sync/internal/sql_manager"
	"github.com/itsrobel/sync/internal/types"
)

// TODO: I need to isolate and refactor this file to abract out the FileServiceClient
type FileWatcher struct {
	watcher *fsnotify.Watcher
	db      *sql.DB
	wait    sync.WaitGroup
	monitor *ServerMonitor
	done    chan struct{}
}

func InitFileWatcher(dbPath, watchPath string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}
	db, err := sql_manager.ConnectSQLite(dbPath)
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	client := filetransferconnect.NewFileServiceClient(http.DefaultClient, "http://localhost:50051")

	monitor := NewServerMonitor(client, 10*time.Second)

	fw := &FileWatcher{
		watcher: watcher,
		db:      db,
		monitor: monitor,
	}

	monitor.Start()
	// if monitor.IsConnected() {
	// 	log.Println("Server is connected")
	// } else {
	// 	log.Println("Server not connected")
	// }

	err = filepath.Walk("./content",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path != "./content" {
				fmt.Println(path, info.Size())
				// NOTE: since I am not doning any comparison
				// operations I am just going to create a new
				// version of every file as the default behavior
				// update later to be a cooler function
				isFile, _ := sql_manager.FindFileByLocation(fw.db, path)
				// log.Println(isFile.ID, isFile.Location, isFile.Contents)
				var file_id string
				if isFile == nil {
					log.Printf("Create new file at: %s", path)
					file_id, _ = sql_manager.CreateFile(fw.db, path)
				} else {
					file_id = isFile.ID
				}
				log.Printf("File exists at: %s", path)
				file, err := os.Open(path) // Note: using event.Name instead of "filename.txt"
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
				if err := sql_manager.CreateFileVersion(fw.db, file_id, content.String()); err != nil {
					return fmt.Errorf("failed to create file version: %w", err)
				}
				fw.file_upload(path)

			}
			return nil
		})
	if err != nil {
		log.Fatal(err)
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
	fw.wait.Add(1)

	go func() {
		defer fw.wait.Done()
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
	// TODO: fix the triggered create event
	if event.Op&fsnotify.Create == fsnotify.Create && sql_manager.ValidFileExtension(event.Name) {
		log.Printf("New file create Event: %s", event.Name)
		isFile, _ := sql_manager.FindFileByLocation(fw.db, event.Name)
		// var fileID string
		if isFile == nil {
			log.Printf("Create new file at: %s", event.Name)
			sql_manager.CreateFile(fw.db, event.Name)
			// if err != nil {
			// 	return fmt.Errorf("failed to create file record: %w", err)
			// }
		} else {
			log.Printf("File exists at: %s", event.Name)
		}

	}
	if event.Op&fsnotify.Write == fsnotify.Write && sql_manager.ValidFileExtension(event.Name) {

		isFile, _ := sql_manager.FindFileByLocation(fw.db, event.Name)
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

		if err := sql_manager.CreateFileVersion(fw.db, isFile.ID, content.String()); err != nil {
			return fmt.Errorf("failed to create file version: %w", err)
		}
	}
	return nil
}

func (fw *FileWatcher) file_upload(file_path string) {
	if fw.monitor.IsConnected() {

		file, openErr := os.Open(file_path)
		id := 1
		if openErr != nil {
			log.Fatalf("Failed to open local file: %v", openErr)
			return
		}
		defer file.Close()
		buf := make([]byte, types.ChunkSize) // Define your buffer size

		// client := filetransferconnect.NewFileServiceClient(http.DefaultClient, "http://localhost:50051")

		// we will be transfering over fileversions mostly
		stream := fw.monitor.client.SendFileToServer(context.Background())
		for {
			log.Printf("Trying to upload...")
			n, readErr := file.Read(buf) // Read from file into buffer

			if n > 0 { // Only send if there's data to send
				fileData := &filetransfer.FileData{
					Id:       fmt.Sprintf("%d", id),
					Location: filepath.Base(file_path), // Use actual filename here
					Content:  buf[:n],                  // Send only n bytes
					Offset:   int64(n),
					// TotalSize: int64,
				}

				if err := stream.Send(fileData); err != nil {
					log.Printf("Client %d error sending file data: %v\n", id, err)
					return
				}
				log.Printf("Sent %d bytes", n)
				res, _ := stream.CloseAndReceive()
				log.Printf("Server response: %v", res)
			}

			if readErr == io.EOF {
				log.Println("Reached end of file")
				break
			}

			if readErr != nil {
				log.Fatalf("Error reading local file: %v", readErr)
				return
			}
		}
	} else {
		log.Println("cannot upload file, not connected to server")
	}
}

func (fw *FileWatcher) Stop() {
	if fw.done != nil {
		close(fw.done)
		fw.wait.Wait()
	}
}
