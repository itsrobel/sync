package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	manager "github.com/itsrobel/sync/internal/mongo_manager"
	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	ct "github.com/itsrobel/sync/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type FileTransferServer struct {
	filetransferconnect.UnimplementedFileServiceHandler
	sessions map[string]*SessionState
	mu       sync.RWMutex
	db       *mongo.Client
}

type SessionState struct {
	controlStream *connect.BidiStream[ft.ControlMessage, ft.ControlMessage]
	isPaused      bool
}

func NewFileTransferServer(db *mongo.Client) *FileTransferServer {
	return &FileTransferServer{
		sessions: make(map[string]*SessionState),
		db:       db,
	}
}

func (s *FileTransferServer) setPauseState(sessionID string, isPaused bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session, exists := s.sessions[sessionID]; exists {
		session.isPaused = isPaused
	}
}

func (s *FileTransferServer) isSessionPaused(sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if session, exists := s.sessions[sessionID]; exists {
		return session.isPaused
	}
	return false
}

// Existing methods remain unchanged
func (s *FileTransferServer) SendFileToServer(ctx context.Context, stream *connect.ClientStream[ft.FileVersionData]) (*connect.Response[ft.ActionResponse], error) {
	success := true
	message := "OK"
	var fileData *ft.FileVersionData
	log.Println("Request headers:", stream.RequestHeader())
	for stream.Receive() {
		// Store the latest message
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

	// TODO: find by file if file location exits create version else create file
	// manager.CreateFile()
	res := connect.NewResponse(&ft.ActionResponse{Success: success, Message: message})
	res.Header().Set("Transfer-Version", "v1")
	collection := s.db.Database("sync").Collection("files")

	manager.CreateFileVersion(collection, fileData)

	return res, nil
}

func ConnectMongo() (*mongo.Client, context.Context) {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB!")
	return client, ctx
}

func main() {
	mongoClient, _ := ConnectMongo()
	filetransfer := NewFileTransferServer(mongoClient)

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

func (s *FileTransferServer) updateClientTimestamp(sessionID string) error {
	collection := s.db.Database("sync").Collection("client_sessions")
	filter := bson.M{"session_id": sessionID}
	update := bson.M{
		"$set": bson.M{
			"last_sync_time": time.Now(),
			"is_active":      true,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

func (s *FileTransferServer) getLastSyncTime(sessionID string) (time.Time, error) {
	collection := s.db.Database("sync").Collection("client_sessions")
	var session ct.ClientSession
	err := collection.FindOne(
		context.Background(),
		bson.M{"session_id": sessionID},
	).Decode(&session)

	if err == mongo.ErrNoDocuments {
		return time.Time{}, nil
	}
	return session.LastSyncTime, err
}

// func (s *FileTransferServer) Greet()

func (s *FileTransferServer) Greet(
	ctx context.Context,
	req *connect.Request[ft.GreetRequest],
) (*connect.Response[ft.GreetResponse], error) {
	fmt.Println("response message: ", req.Msg.Name)
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
	log.Println(msg)
	if err != nil {
		return err
	}

	sessionID := msg.SessionId

	// Send READY response first
	if err := stream.Send(&ft.ControlMessage{
		SessionId: sessionID,
		Type:      ft.ControlMessage_READY,
	}); err != nil {
		return err
	}
	lastSync, err := s.getLastSyncTime(sessionID)
	if err != nil {
		return err
	}
	log.Printf("Last sync time: %s, client: %s", lastSync, sessionID)
	collection := s.db.Database("sync").Collection("files")

	// manager.DeleteAllDocuments(collection)
	manager.GetAllDocuments(collection)

	filter := bson.M{
		"timestamp": bson.M{"$gt": lastSync},
		"active":    true,
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var file ct.File
		if err := cursor.Decode(&file); err != nil {
			continue
		}
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
