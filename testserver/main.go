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

	"github.com/fsnotify/fsnotify"
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

	go func() {
		for {
			req, err := stream.Recv() // Receive file data from client (for upload)
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Printf("Error receiving data from client: %v", err)
				return
			}
			if req.Data != nil { // Client is uploading a file
				s.handleFileUpload(req)
			}
		}
	}()
	return nil
}

// handleFileUpload: Save uploaded file chunks from the client.
func (s *server) handleFileUpload(fileData *pb.FileData) error {
	// filepath := "/path/to/upload/" + fileData.Filename
	filePath := filepath.Join(directory, fileData.Filename)
	print(fileData.Filename)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open local file %s: %v", filePath, err)
		return err
	}
	defer file.Close()

	_, err = file.Write(fileData.Data)
	if err != nil {
		log.Printf("Failed writing data to local file %s: %v", filePath, err)
		return err
	}

	log.Printf("Successfully saved data to %s", filePath)
	return nil
}

// Watch files and notify all connected clients when a file changes.
func (s *server) watchFiles(path string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Watching directory: %s\n", path)

	if err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				log.Println("Modified or created file:", event.Name)
				s.pushFileUpdate(event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Println("Error:", err)
		}
	}
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
			Filename:  filepath.Base(filePath),
			Data:      buffer[:n],
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

	s := grpc.NewServer()
	pb.RegisterFileServiceServer(s, srv)

	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
