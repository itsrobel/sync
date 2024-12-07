package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
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

// func (s *FileTransferServer) SendFileToClient(ctx context.Context, stream *connect.Request[ft.FileRequest]) *connect.ServerStream[ft.FileData] {
// 	return nil
// }

func main() {
	filetransfer := &FileTransferServer{}
	mux := http.NewServeMux()
	path, handler := filetransferconnect.NewFileServiceHandler(filetransfer)
	mux.Handle(path, handler)
	http.ListenAndServe("localhost:50051", h2c.NewHandler(mux, &http2.Server{}))
}
