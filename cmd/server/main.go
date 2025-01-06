package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"connectrpc.com/connect"
	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type FileTransferServer struct {
	filetransferconnect.UnimplementedFileServiceHandler
}

type server struct {
	transferserver FileTransferServer
	mongoclient    *mongo.Client
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

// TODO: write the file to mongodb saving
// first step is to try and locate the file using a get method from mongodb then create a new version if found, else create the file and then version

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

// func (s *FileTransferServer) SendFileToClient(ctx context.Context, stream *connect.Request[ft.FileRequest]) *connect.ServerStream[ft.FileData] {
// 	return nil
// }

func main() {
	filetransfer := &FileTransferServer{}
	mux := http.NewServeMux()
	path, handler := filetransferconnect.NewFileServiceHandler(filetransfer)
	mux.Handle(path, handler)
	log.Println("Server started on port 50051")
	http.ListenAndServe("localhost:50051", h2c.NewHandler(mux, &http2.Server{}))
}

type WatcherServer struct {
	grpcServer FileTransferServer
	watcher    *watcher.FileWatcher
}

func (c *WatcherServer) HandleChange(path string, eventType string) error {
	// Send gRPC notification to server
	// _, err := c.grpcClient.NotifyFileChange(context.Background(), &FileChangeRequest{
	//     Path:      path,
	//     EventType: eventType,
	// })
	// return err
	return nil
}

func NewWatcherServer(grpcServer FileTransferServer) *WatcherServer {
	watcher, _ := watcher.NewFileWatcher()
	server := &WatcherServer{
		grpcServer: grpcServer,
		watcher:    watcher,
	}
	server.watcher.AddHandler(server)
	return server
}

func main() {
	Server()
}
