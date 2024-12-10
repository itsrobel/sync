package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	"github.com/itsrobel/sync/internal/watcher"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type FileTransferServer struct {
	filetransferconnect.UnimplementedFileServiceHandler
}

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

//	func (s *FileTransferServer) SendFileToClient(ctx context.Context, stream *connect.Request[ft.FileRequest]) *connect.ServerStream[ft.FileData] {
//		return nil
//	}
func Server() {
	filetransfer := &FileTransferServer{}
	mux := http.NewServeMux()
	path, handler := filetransferconnect.NewFileServiceHandler(filetransfer)
	mux.Handle(path, handler)
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
