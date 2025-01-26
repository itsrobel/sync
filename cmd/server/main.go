package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	manager "github.com/itsrobel/sync/internal/postgres_manager"
	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	ct "github.com/itsrobel/sync/internal/types"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/driver/postgres"
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

type ClientSession struct {
	SessionID    string `gorm:"primaryKey"`
	LastSyncTime time.Time
	IsActive     bool
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

	if err := manager.CreateFileVersion(s.db, fileData); err != nil {
		return connect.NewResponse(&ft.ActionResponse{
			Success: false,
			Message: err.Error(),
		}), err
	}

	if err := manager.UpdateFile(s.db, &ct.File{
		ID:       fileData.FileId,
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

func ConnectDatabase() (*gorm.DB, error) {
	// dsn := "host=localhost user=postgres password=yourpassword dbname=sync port=5432 sslmode=disable"
	dsn := "host=localhost user=postgres password=postgres dbname=myapp port=5432 sslmode=disable"
	// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
	// 	os.Getenv("DB_HOST"),
	// 	os.Getenv("DB_USER"),
	// 	os.Getenv("DB_PASSWORD"),
	// 	os.Getenv("DB_NAME"),
	// 	os.Getenv("DB_PORT"),
	// )
	// log.Println(dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the schemas
	if err := manager.AutoMigrate(db); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&ClientSession{}); err != nil {
		return nil, err
	}
	log.Println("Connected to Postgres ")
	return db, nil
}

func main() {
	db, err := ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

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
	return s.db.Model(&ClientSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"last_sync_time": time.Now(),
			"is_active":      true,
		}).Error
}

// func (s *FileTransferServer) getOrCreateClientSession(sessionID string) (*ClientSession, error) {
// 	var session ClientSession
//
// 	result := s.db.Where("session_id = ?", sessionID).First(&session)
// 	if result.Error == gorm.ErrRecordNotFound {
// 		// Create new session if not found
// 		session = ClientSession{
// 			SessionID:    sessionID,
// 			LastSyncTime: time.Now(),
// 			IsActive:     true,
// 		}
// 		if err := s.db.Create(&session).Error; err != nil {
// 			return nil, fmt.Errorf("failed to create session: %v", err)
// 		}
// 	} else if result.Error != nil {
// 		return nil, result.Error
// 	}
//
// 	return &session, nil
// }

func (s *FileTransferServer) getLastSyncTime(sessionID string) (time.Time, error) {
	var session ClientSession
	err := s.db.Where("session_id = ?", sessionID).First(&session).Error
	if err == gorm.ErrRecordNotFound {
		session = ClientSession{
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

	docs, err := manager.GetAllDocuments(s.db)
	if err != nil {
		return nil, err
	}
	log.Printf("Found %d documents", len(docs))

	response := connect.NewResponse(&ft.GreetResponse{
		Greeting: fmt.Sprintf("Hello, %s!", req.Msg.Name),
	})

	return response, nil
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

	var files []ct.File
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
