package main

import (
	"io"
	"log"
	"net"
	"os"
	"watcher/filetransfer"

	"google.golang.org/grpc"
)

type server struct {
	filetransfer.UnimplementedFileServiceServer
}

// NOTE: the file that is uploaded depends on where in the file directory the program is run from.
func (s *server) SendFile(req *filetransfer.FileRequest, stream filetransfer.FileService_SendFileServer) error {
	// Open the file for reading
	file, err := os.Open(req.Filename)
	if err != nil {
		return err
	}
	defer file.Close()
	// Read the file and stream the content
	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// Send the file content in chunks
		if err := stream.Send(&filetransfer.FileResponse{Data: buffer[:n]}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	filetransfer.RegisterFileServiceServer(s, &server{})
	log.Println("Server listening at", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
