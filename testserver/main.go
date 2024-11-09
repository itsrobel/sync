package main

import (
	// "fmt"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	pb "watcher/filetransfer" // Replace with the actual path to the generated

	"google.golang.org/grpc"
)

const (
	directory = "content"
	chunkSize = 64 * 1024
) // Chunk size for streaming files.

type server struct {
	pb.UnimplementedFileServiceServer
	clients map[string]*clientSession
	mu      sync.RWMutex
}

type clientSession struct {
	stream     pb.FileService_TransferFileServer
	id         string
	active     bool
	lastOffset int64
}

func newServer() *server {
	return &server{
		clients: make(map[string]*clientSession),
	}
}

func (s *server) registerClient(clientID string, stream pb.FileService_TransferFileServer) *clientSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	session := &clientSession{
		id:     clientID,
		active: true,
		stream: stream,
	}
	s.clients[clientID] = session
	log.Printf("Client registered: %s\n", clientID)
	return session
}

func (s *server) unregisterClient(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, clientID)
	log.Printf("Client unregistered: %s\n", clientID)
}

func (s *server) TransferFile(stream pb.FileService_TransferFileServer) error {
	clientID := generateClientID()
	session := s.registerClient(clientID, stream)
	defer s.unregisterClient(clientID)

	// Create channels for coordination
	errChan := make(chan error, 1)
	done := make(chan bool)
	defer close(done)

	// Start receiving goroutine
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				fileData, err := stream.Recv()
				if err == io.EOF {
					errChan <- nil
					return
				}
				if err != nil {
					errChan <- err
					return
				}

				s.mu.Lock()
				if client, exists := s.clients[clientID]; exists {
					client.lastOffset = fileData.Offset
				}
				s.mu.Unlock()

				log.Printf("Client %s: Received file chunk: ID=%s, Location=%s, Content=%s, Offset=%d, Size=%d\n",
					clientID, fileData.Id, fileData.Location, fileData.Content, fileData.Offset, fileData.TotalSize)

				// Send acknowledgment
				response := &pb.FileData{
					Id:        fileData.Id,
					Location:  fileData.Location,
					Offset:    fileData.Offset,
					TotalSize: fileData.TotalSize,
				}

				if err := session.stream.Send(response); err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	// Wait for error or completion
	return <-errChan
}

func (s *server) GetActiveClients() []*clientSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeClients := make([]*clientSession, 0, len(s.clients))
	for _, client := range s.clients {
		if client.active {
			activeClients = append(activeClients, client)
		}
	}
	return activeClients
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	fileServer := newServer()
	pb.RegisterFileServiceServer(s, fileServer)

	// Monitor active clients
	go func() {
		for {
			time.Sleep(5 * time.Second)
			activeClients := fileServer.GetActiveClients()
			log.Printf("Active clients: %d\n", len(activeClients))
			for _, client := range activeClients {
				log.Printf("Client ID: %s, Last Offset: %d\n", client.id, client.lastOffset)
			}
		}
	}()

	log.Println("Server started on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}
