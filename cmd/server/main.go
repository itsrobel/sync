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
	ct "github.com/itsrobel/sync/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type FileTransferServer struct {
	filetransferconnect.UnimplementedFileServiceHandler
	sessions map[string]*sessionState
	mu       sync.RWMutex
	db       *mongo.Client
}

type sessionState struct {
	controlStream *connect.BidiStream[ft.ControlMessage, ft.ControlMessage]
	isPaused      bool
}

func NewFileTransferServer(db *mongo.Client) *FileTransferServer {
	return &FileTransferServer{
		sessions: make(map[string]*sessionState),
		db:       db,
	}
}

// func (s *FileTransferServer) ControlStream(
// 	ctx context.Context,
// 	stream *connect.BidiStream[ft.ControlMessage, ft.ControlMessage],
// ) error {
// 	// First message should contain session initialization
// 	msg, err := stream.Receive()
// 	if err != nil {
// 		return err
// 	}
//
// 	sessionID := msg.SessionId
// 	s.mu.Lock()
// 	s.sessions[sessionID] = &sessionState{
// 		controlStream: stream,
// 		isPaused:      false,
// 	}
// 	s.mu.Unlock()
//
// 	// Send acknowledgment
// 	if err := stream.Send(&ft.ControlMessage{
// 		SessionId: sessionID,
// 		Type:      ft.ControlMessage_READY,
// 	}); err != nil {
// 		return err
// 	}
//
// 	// Handle session cleanup
// 	defer func() {
// 		s.mu.Lock()
// 		delete(s.sessions, sessionID)
// 		s.mu.Unlock()
// 	}()
//
// 	// Handle incoming control messages
// 	for {
// 		msg, err := stream.Receive()
// 		if err != nil {
// 			return err
// 		}
//
// 		switch msg.Type {
// 		case ft.ControlMessage_READY:
// 			log.Printf("Client %s ready for transfer", sessionID)
// 		case ft.ControlMessage_START_TRANSFER:
// 			log.Printf("Starting transfer for client %s: %s", sessionID, msg.Filename)
// 		case ft.ControlMessage_PAUSE:
// 			s.setPauseState(sessionID, true)
// 		case ft.ControlMessage_RESUME:
// 			s.setPauseState(sessionID, false)
// 		}
// 	}
// }

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
func (s *FileTransferServer) SendFileToServer(ctx context.Context, stream *connect.ClientStream[ft.FileData]) (*connect.Response[ft.ActionResponse], error) {
	success := true
	message := "OK"

	log.Println("Request headers:", stream.RequestHeader())
	for stream.Receive() {
		fmt.Printf("File data: %s", stream.Msg().Content)
	}

	if err := stream.Err(); err != nil {
		log.Println("Stream error:", err)
		success = false
	}
	res := connect.NewResponse(&ft.ActionResponse{Success: success, Message: message})
	res.Header().Set("Transfer-Version", "v1")

	return res, nil
}

func (s *FileTransferServer) ValidateServer(ctx context.Context, req *connect.Request[ft.ActionResponse]) (*connect.Response[ft.ActionResponse], error) {
	return connect.NewResponse(&ft.ActionResponse{
		Success: true,
		Message: "OK",
	}), nil
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
	// The connection context only lasts as long as specified in the timemout, since
	// We are running these commands not on a time frame we should be able to use contex.TODO although that is likely
	// not best practice
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
	lastSync, err := s.getLastSyncTime(sessionID)
	if err != nil {
		return err
	}
	log.Printf("Last sync time: %s, client: %s", lastSync, sessionID)

	// Get files modified since last sync
	//NOTE: I don't have to create another message for controlling file emits since the
	//"socket" already reconnects for latest files anyway
	//all I have to do I send files to the client
	collection := s.db.Database("sync").Collection("files")
	filter := bson.M{
		"timestamp": bson.M{"$gt": lastSync},
		"active":    true,
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	// Send modified files to client
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

	// Update client's last sync time
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
