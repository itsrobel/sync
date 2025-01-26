package watcher

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// "strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/fsnotify/fsnotify"
	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	"github.com/itsrobel/sync/internal/sql_manager"
	ct "github.com/itsrobel/sync/internal/types"

	// ct "github.com/itsrobel/sync/internal/types"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type FileWatcher struct {
	watcher       *fsnotify.Watcher
	db            *gorm.DB
	wait          sync.WaitGroup
	done          chan struct{}
	client        filetransferconnect.FileServiceClient
	sessionID     string
	controlStream *connect.BidiStreamForClient[ft.ControlMessage, ft.ControlMessage]
	isConnected   bool
	mu            sync.RWMutex
}

func InitFileWatcher(dbPath, watchPath, clientName string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	db, err := sql_manager.ConnectSQLite(dbPath)
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	client := filetransferconnect.NewFileServiceClient(
		&http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		},
		"http://localhost:50051",
	)

	fw := &FileWatcher{
		watcher:   watcher,
		db:        db,
		client:    client,
		sessionID: clientName,
		done:      make(chan struct{}),
	}
	go fw.connectionTicker()

	// Start the connection ticker

	// Process initial files regardless of connection status
	if err := fw.processInitialFiles(watchPath); err != nil {
		return nil, err
	}

	if err := fw.startWatching(watchPath); err != nil {
		return nil, err
	}

	return fw, nil
}

func (fw *FileWatcher) connectionTicker() {
	if err := fw.attemptConnection(); err != nil {
		log.Printf("Failed to connect to server: %v", err)
	}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !fw.IsConnected() {
				if err := fw.attemptConnection(); err != nil {
					log.Printf("Failed to connect to server: %v", err)
				}
			}
		case <-fw.done:
			return
		}
	}
}

func (fw *FileWatcher) attemptConnection() error {
	fw.mu.Lock()
	if fw.controlStream != nil {
		fw.controlStream = nil
	}
	fw.mu.Unlock()

	stream := fw.client.ControlStream(context.Background())

	fw.mu.Lock()
	fw.controlStream = stream
	fw.mu.Unlock()

	if err := stream.Send(&ft.ControlMessage{
		SessionId: fw.sessionID,
		Type:      ft.ControlMessage_READY,
	}); err != nil {
		fw.setConnected(false)
		return fmt.Errorf("failed to send initial message")
	}

	go fw.handleControlStream()
	return nil
}

func (fw *FileWatcher) handleControlStream() {
	defer func() {
		fw.mu.Lock()
		fw.controlStream = nil
		fw.mu.Unlock()
		fw.setConnected(false)
	}()

	for {

		msg, err := fw.controlStream.Receive()
		if err != nil {
			fw.setConnected(false)
			log.Printf("Control stream error: %v", err)
			return
		}

		log.Println("the message is: ", msg)

		switch msg.Type {
		case ft.ControlMessage_READY:
			fw.setConnected(true)
			log.Printf("Server connection established for session: %s", fw.sessionID)
		case ft.ControlMessage_NEW_FILE:
			log.Printf("New file available on server: %s", msg.Filename)
		}
	}
}

func (fw *FileWatcher) startControlStream() error {
	stream := fw.client.ControlStream(context.Background())

	fw.mu.Lock()
	fw.controlStream = stream
	fw.mu.Unlock()

	if err := stream.Send(&ft.ControlMessage{
		SessionId: fw.sessionID,
		Type:      ft.ControlMessage_READY,
	}); err != nil {
		return fmt.Errorf("failed to send initial message: %w", err)
	}

	go fw.handleControlStream()
	return nil
}

// NOTE: this now uploads via the information returned in the database
func (fw *FileWatcher) file_upload(fileVersion *sql_manager.FileVersion) error {
	log.Println("uploading file: ", fileVersion.Location)
	if err := fw.sendControlMessage(&ft.ControlMessage{
		SessionId: fw.sessionID,
		Type:      ft.ControlMessage_START_TRANSFER,
		Filename:  filepath.Base(fileVersion.Location),
	}); err != nil {
		return err
	}

	// file, err := os.Open(fileVersion.)
	// if err != nil {
	// 	return fmt.Errorf("failed to open file: %w", err)
	// }
	// defer file.Close()

	stream := fw.client.SendFileToServer(context.Background())
	buffer := []byte(fileVersion.Content)
	chunkSize := ct.ChunkSize

	for i := 0; i < len(buffer); i += chunkSize {
		end := i + chunkSize
		if end > len(buffer) {
			end = len(buffer)
		}

		chunk := buffer[i:end]
		if err := stream.Send(&ft.FileVersionData{
			Id:        fileVersion.ID,
			Location:  fileVersion.Location, // or any identifier you want to use
			FileId:    fileVersion.FileID,   // or any identifier you want to use
			Timestamp: timestamppb.New(fileVersion.Timestamp),
			Client:    fw.sessionID,
			Content:   chunk,
			Offset:    int64(len(chunk)),
		}); err != nil {
			return fmt.Errorf("error sending string data: %v", err)
		}
	}

	res, err := stream.CloseAndReceive()
	if err != nil {
		return fmt.Errorf("error closing stream: %v", err)
	}

	log.Printf("Upload completed: %v", res)
	return nil
}

func (fw *FileWatcher) sendControlMessage(msg *ft.ControlMessage) error {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	if fw.controlStream == nil {
		return fmt.Errorf("control stream not initialized")
	}
	return fw.controlStream.Send(msg)
}

func (fw *FileWatcher) setConnected(status bool) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.isConnected = status
}

func (fw *FileWatcher) IsConnected() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.isConnected
}

func (fw *FileWatcher) processInitialFiles(watchPath string) error {
	return filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path != watchPath {
			// TODO: find files by location is likely broken
			file, err := sql_manager.FindFileByLocation(fw.db, path)
			if err != nil {
				return err
			}
			if file == nil {
				var err error
				file, err = sql_manager.CreateFileInitial(fw.db, path)
				if err != nil {
					return fmt.Errorf("failed to create file record: %w", err)
				}
				log.Printf("Created new file record: %s", path)
			}

			// TODO: have proccesfilecontent return file to then upload
			fileVersion, err := fw.processFileContent(path, file)
			if err != nil {
				return err
			}

			log.Println("connection status: ", fw.IsConnected())
			if fw.IsConnected() {

				log.Println("connected to server and uploading file: ", fileVersion.Location)
				if err := fw.file_upload(fileVersion); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (fw *FileWatcher) processFileContent(path string, file *sql_manager.File) (*sql_manager.FileVersion, error) {
	raw_file, err := os.Open(path)
	tmpFV := &sql_manager.FileVersion{}
	if err != nil {
		return tmpFV, err
	}
	defer raw_file.Close()

	var content strings.Builder
	buffer := make([]byte, 8192)
	for {
		n, err := raw_file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return tmpFV, err
		}
		content.Write(buffer[:n])
	}

	fileVersion, err := sql_manager.CreateFileVersion(fw.db, file, content.String())
	if err != nil {
		return fileVersion, err
	}
	return fileVersion, nil
}

func (fw *FileWatcher) startWatching(path string) error {
	if err := fw.watcher.Add(path); err != nil {
		return fmt.Errorf("failed to add path to watcher: %w", err)
	}

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

	log.Printf("Started watching directory: %s", path)
	return nil
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event) error {
	if !ValidFileExtension(event.Name) {
		return nil
	}

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		file, _ := sql_manager.FindFileByLocation(fw.db, event.Name)

		if file == nil {
			file, err := sql_manager.CreateFileInitial(fw.db, event.Name)
			if err != nil {
				return fmt.Errorf("failed to create file record: %w", err)
			}
			log.Printf("Created new file: %s", event.Name)
			_, err = fw.processFileContent(event.Name, file)
			return err
		}

	case event.Op&fsnotify.Write == fsnotify.Write:
		isFile, _ := sql_manager.FindFileByLocation(fw.db, event.Name)
		if isFile == nil {
			return fmt.Errorf("file not found in database: %s", event.Name)
		}

		fileVersion, err := fw.processFileContent(event.Name, isFile)
		if err != nil {
			return err
		}
		log.Printf("fileVersion: %v", fileVersion)
		return fw.file_upload(fileVersion)
	}
	return nil
}

func ValidFileExtension(location string) bool {
	extensions := []string{".md", ".pdf"}
	for _, ext := range extensions {
		if strings.HasSuffix(location, ext) {
			return true
		}
	}
	return false
}

func (fw *FileWatcher) Stop() {
	if fw.done != nil {
		close(fw.done)
		fw.wait.Wait()
	}
	if fw.watcher != nil {
		fw.watcher.Close()
	}
	// if fw.db != nil {
	// 	fw.db.Close()
	// }
}
