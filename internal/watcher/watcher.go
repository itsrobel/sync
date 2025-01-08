package watcher

import (
	"context"
	"crypto/tls"
	"database/sql"
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
	"github.com/itsrobel/sync/internal/types"
	"golang.org/x/net/http2"
)

type FileWatcher struct {
	watcher       *fsnotify.Watcher
	db            *sql.DB
	wait          sync.WaitGroup
	done          chan struct{}
	client        filetransferconnect.FileServiceClient
	sessionID     string
	controlStream *connect.BidiStreamForClient[ft.ControlMessage, ft.ControlMessage]
	isConnected   bool
	mu            sync.RWMutex
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
		watcher: watcher,
		db:      db,
		// client:    filetransferconnect.NewFileServiceClient(http.DefaultClient, "http://localhost:50051"),
		client:    client,
		sessionID: fmt.Sprintf("session_%d", time.Now().UnixNano()),
		done:      make(chan struct{}),
	}

	// Start the connection ticker
	go fw.connectionTicker()

	// Process initial files regardless of connection status
	if err := fw.processInitialFiles(); err != nil {
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

		switch msg.Type {
		case ft.ControlMessage_READY:
			fw.setConnected(true)
			log.Printf("Server connection established for session: %s", fw.sessionID)
		case ft.ControlMessage_NEW_FILE_AVAILABLE:
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

func (fw *FileWatcher) file_upload(file_path string) error {
	if err := fw.sendControlMessage(&ft.ControlMessage{
		SessionId: fw.sessionID,
		Type:      ft.ControlMessage_START_TRANSFER,
		Filename:  filepath.Base(file_path),
	}); err != nil {
		return err
	}

	file, err := os.Open(file_path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stream := fw.client.SendFileToServer(context.Background())
	buffer := make([]byte, types.ChunkSize)

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
		}

		if err := stream.Send(&ft.FileData{
			Id:       fw.sessionID,
			Location: filepath.Base(file_path),
			Content:  buffer[:n],
			Offset:   int64(n),
		}); err != nil {
			return fmt.Errorf("error sending file data: %v", err)
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

func (fw *FileWatcher) processInitialFiles() error {
	return filepath.Walk("./content", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path != "./content" {
			isFile, _ := sql_manager.FindFileByLocation(fw.db, path)
			var fileID string

			if isFile == nil {
				var err error
				fileID, err = sql_manager.CreateFile(fw.db, path)
				if err != nil {
					return fmt.Errorf("failed to create file record: %w", err)
				}
				log.Printf("Created new file record: %s", path)
			} else {
				fileID = isFile.ID
			}

			if err := fw.processFileContent(path, fileID); err != nil {
				return err
			}

			if fw.IsConnected() {
				if err := fw.file_upload(path); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (fw *FileWatcher) processFileContent(path, fileID string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var content strings.Builder
	buffer := make([]byte, 8192)

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		content.Write(buffer[:n])
	}

	if err := sql_manager.CreateFileVersion(fw.db, fileID, content.String()); err != nil {
		return fmt.Errorf("failed to create file version: %w", err)
	}

	return nil
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
	if !sql_manager.ValidFileExtension(event.Name) {
		return nil
	}

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		isFile, _ := sql_manager.FindFileByLocation(fw.db, event.Name)
		if isFile == nil {
			fileID, err := sql_manager.CreateFile(fw.db, event.Name)
			if err != nil {
				return fmt.Errorf("failed to create file record: %w", err)
			}
			log.Printf("Created new file: %s", event.Name)
			return fw.processFileContent(event.Name, fileID)
		}

	case event.Op&fsnotify.Write == fsnotify.Write:
		isFile, _ := sql_manager.FindFileByLocation(fw.db, event.Name)
		if isFile == nil {
			return fmt.Errorf("file not found in database: %s", event.Name)
		}
		if err := fw.processFileContent(event.Name, isFile.ID); err != nil {
			return err
		}
		return fw.file_upload(event.Name)
	}

	return nil
}

func (fw *FileWatcher) Stop() {
	if fw.done != nil {
		close(fw.done)
		fw.wait.Wait()
	}
	if fw.watcher != nil {
		fw.watcher.Close()
	}
	if fw.db != nil {
		fw.db.Close()
	}
}
