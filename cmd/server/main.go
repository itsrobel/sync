package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	sql_manager "github.com/itsrobel/sync/internal/sql_manager"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/gorm"
)

type FileTransferServer struct {
	filetransferconnect.UnimplementedFileServiceHandler
	sessions map[string]*SessionState
	mu       sync.RWMutex
	db       *gorm.DB
}

type SessionState struct {
	controlStream *connect.BidiStream[ft.ControlMessage, ft.ControlMessage]
	isPaused      bool
}

func NewFileTransferServer(db *gorm.DB) *FileTransferServer {
	return &FileTransferServer{
		sessions: make(map[string]*SessionState),
		db:       db,
	}
}

func (s *FileTransferServer) SendFileToServer(ctx context.Context, stream *connect.ClientStream[ft.FileVersionData]) (*connect.Response[ft.ActionResponse], error) {
	success := true
	message := "OK"
	var fileData *ft.FileVersionData

	log.Println("Request headers:", stream.RequestHeader())
	for stream.Receive() {
		fileData = stream.Msg()
		log.Println("Processing chunk for file:", fileData.Id, fileData.Location)
	}

	if err := stream.Err(); err != nil {
		log.Println("Stream error:", err)
		success = false
	}

	if fileData == nil {
		return connect.NewResponse(&ft.ActionResponse{
			Success: false,
			Message: "No data received",
		}), fmt.Errorf("no data received")
	}

	res := connect.NewResponse(&ft.ActionResponse{Success: success, Message: message})
	res.Header().Set("Transfer-Version", "v1")

	if err := sql_manager.CreateFileVersionServer(s.db, fileData); err != nil {
		return connect.NewResponse(&ft.ActionResponse{
			Success: false,
			Message: err.Error(),
		}), err
	}

	if err := sql_manager.UpdateFileServer(s.db, &sql_manager.File{
		FileBase: sql_manager.FileBase{ID: fileData.FileId},
		Location: fileData.Location,
		Content:  string(fileData.Content),
		Active:   true,
	}); err != nil {
		return connect.NewResponse(&ft.ActionResponse{
			Success: false,
			Message: err.Error(),
		}), err
	}

	return res, nil
}

func main() {
	db, err := sql_manager.ConnectPostgres()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// sql_manager.DeleteAllFiles(db)
	// sql_manager.DeleteAllFileVersions(db)
	filetransfer := NewFileTransferServer(db)

	mux := http.NewServeMux()
	path, handler := filetransferconnect.NewFileServiceHandler(filetransfer)
	mux.Handle(path, handler)

	server := &http.Server{
		Addr: "localhost:50051",
		Handler: h2c.NewHandler(mux, &http2.Server{
			MaxConcurrentStreams: 250,
			MaxReadFrameSize:     16384,
			IdleTimeout:          10 * time.Second,
		}),
	}

	log.Println("Server started on port 50051")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// NOTE: if it is the clients first connection there is no sessionID to search for
func (s *FileTransferServer) updateClientTimestamp(sessionID string) error {
	return s.db.Model(&sql_manager.ClientSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"last_sync_time": time.Now(),
			"is_active":      true,
		}).Error
}

func (s *FileTransferServer) getLastSyncTime(sessionID string) (time.Time, error) {
	var session sql_manager.ClientSession
	err := s.db.Where("session_id = ?", sessionID).First(&session).Error
	if err == gorm.ErrRecordNotFound {
		session = sql_manager.ClientSession{
			SessionID:    sessionID,
			LastSyncTime: time.Now(),
			IsActive:     true,
		}
		if err := s.db.Create(&session).Error; err != nil {
			return time.Time{}, fmt.Errorf("failed to create session: %v", err)
		}
		return time.Time{}, nil
	}
	return session.LastSyncTime, err
}

func (s *FileTransferServer) Greet(
	ctx context.Context,
	req *connect.Request[ft.GreetRequest],
) (*connect.Response[ft.GreetResponse], error) {
	fmt.Println("response message: ", req.Msg.Name)

	docs, err := sql_manager.GetAllFiles(s.db)
	if err != nil {
		return nil, err
	}

	// manager.GetAllDocumentsVersions(s.db)
	log.Printf("Found %d documents", len(docs))
	log.Println(docs)

	response := connect.NewResponse(&ft.GreetResponse{
		Greeting: fmt.Sprintf("Hello, %s!", req.Msg.Name),
	})

	return response, nil
}

func (s *FileTransferServer) RetrieveListOfFiles(
	ctx context.Context,
	req *connect.Request[ft.ActionRequest],
) (*connect.Response[ft.FileList], error) {
	docs, err := sql_manager.GetAllFiles(s.db)
	if err != nil {
		return nil, err
	}

	files := make([]*ft.File, len(docs))
	for idx, doc := range docs {
		files[idx] = &ft.File{
			ID:       doc.ID,
			Active:   doc.Active,
			Location: doc.Location,
			Content:  doc.Content,
		}
	}
	log.Println(files)

	return connect.NewResponse(&ft.FileList{
		Files: files,
	}), nil
}

func (s *FileTransferServer) ControlStream(
	ctx context.Context,
	stream *connect.BidiStream[ft.ControlMessage, ft.ControlMessage],
) error {
	msg, err := stream.Receive()
	if err != nil {
		return err
	}

	sessionID := msg.SessionId

	if err := stream.Send(&ft.ControlMessage{
		SessionId: sessionID,
		Type:      ft.ControlMessage_READY,
	}); err != nil {
		return err
	}

	lastSync, _ := s.getLastSyncTime(sessionID)
	// if err != nil {
	// 	return err
	// }
	log.Printf("Last sync time: %s, client: %s", lastSync, sessionID)

	var files []sql_manager.File
	if err := s.db.Where("timestamp > ? AND active = ?", lastSync, true).Find(&files).Error; err != nil {
		return err
	}

	for _, file := range files {
		if err := stream.Send(&ft.ControlMessage{
			SessionId: sessionID,
			Type:      ft.ControlMessage_NEW_FILE,
			Filename:  file.Location,
		}); err != nil {
			return err
		}
	}

	if err := s.updateClientTimestamp(sessionID); err != nil {
		return err
	}

	for {
		msg, err := stream.Receive()
		log.Println(msg)
		if err != nil {
			return err
		}
	}
}
