package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	pb "watcher/filetransfer" // Replace with the actual path to the generated

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	directory = "content"
	chunkSize = 64 * 1024
) // Chunk size for streaming files.

type server struct {
	pb.UnimplementedFileServiceServer
	clients map[string]pb.FileService_StreamFilesServer // Track connected clients
	mu      sync.Mutex                                  // Protect access to clients map
}

// StreamFiles: Bidirectional streaming that handles both uploads and downloads.
func (s *server) StreamFiles(stream pb.FileService_StreamFilesServer) error {
	clientID := uuid.NewString() // Generate a unique ID for each client connection

	s.mu.Lock()
	s.clients[clientID] = stream // Register client stream
	log.Println("clientID: ", clientID)
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientID) // Remove client when done
		s.mu.Unlock()
	}()

	// var filename string

	go func() {
		for {
			req, err := stream.Recv() // Receive file data from client (for upload)
			if err == io.EOF {
				log.Println("File upload completed.")
				return // End of stream; close connection gracefully.
			}
			if err != nil {
				log.Printf("Error receiving data from client: %v", err)
				return
			}
			if req.Content != nil { // Client is uploading a file
				s.handleFileUpload(req)
			}
		}
	}()
	return nil
}

// handleFileUpload: Save uploaded file chunks from the client.
func (s *server) handleFileUpload(fileData *pb.FileData) error {
	// filepath := "/path/to/upload/" + fileData.Filename
	filePath := filepath.Join(directory, fileData.Location)
	// print(fileData.Filename)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open local file %s: %v", filePath, err)
		return err
	}
	defer file.Close()

	_, err = file.Write(fileData.Content)
	if err != nil {
		log.Printf("Failed writing data to local file %s: %v", filePath, err)
		return err
	}

	log.Printf("Successfully saved data to %s", filePath)
	return nil
}

// Push updated file content to all connected clients.
func (s *server) pushFileUpdate(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	buffer := make([]byte, chunkSize)

	for {
		n, readErr := file.Read(buffer)
		if readErr == io.EOF && n == 0 {
			break
		}
		if readErr != nil && readErr != io.EOF {
			log.Printf("Failed reading file: %v", readErr)
			return
		}
		fileData := &pb.FileData{
			Location:  filepath.Base(filePath),
			Content:   buffer[:n],
			Offset:    int64(n),
			TotalSize: int64(n),
		}

		// s.mu.Lock()
		for _, clientStream := range s.clients {
			err := clientStream.Send(fileData)
			if err != nil {
				log.Printf("Failed sending update to client: %v", err)
				continue
			}
		}
		// s.mu.Unlock()
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	srv := &server{
		clients: make(map[string]pb.FileService_StreamFilesServer),
	}

	path, _ := filepath.Abs(fmt.Sprintf("./%s", directory))
	go srv.watchFiles(path)

	server := grpc.NewServer()
	pb.RegisterFileServiceServer(server, srv)

	log.Println("gRPC server listening on :50051")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
