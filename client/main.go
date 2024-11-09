package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	pb "watcher/filetransfer"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	directory = "content"
	chunkSize = 64 * 1024
)

func runClient(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Client %d failed to connect: %v\n", id, err)
		return
	}
	defer conn.Close()
	client := pb.NewFileServiceClient(conn)
	stream, err := client.TransferFile(context.Background())
	filePath := filepath.Join(directory, "t.md")
	uploadFile(stream, filePath)
}

func main() {
	// numClients := 3
	var wg sync.WaitGroup
	// // Start multiple clients
	// for i := 0; i < numClients; i++ {
	wg.Add(1)
	// 	go runClient(i, &wg)
	// 	time.Sleep(time.Millisecond * 100) // Stagger client starts
	// }
	go runClient(1, &wg)
	wg.Wait()
}

// Upload a local file using bidirectional streaming.
func uploadFile(stream pb.FileService_TransferFileClient, filePath string) {
	file, openErr := os.Open(filePath)
	id := 1
	if openErr != nil {
		log.Fatalf("Failed to open local file: %v", openErr)
		return
	}
	defer file.Close()
	buf := make([]byte, chunkSize) // Define your buffer size

	for {

		log.Printf("Trying to upload...")
		n, readErr := file.Read(buf) // Read from file into buffer

		if n > 0 { // Only send if there's data to send
			fileData := &pb.FileData{
				Id:       fmt.Sprintf("%d", id),
				Location: filepath.Base(filePath), // Use actual filename here
				Content:  buf[:n],                 // Send only n bytes
				Offset:   int64(n),
				// TotalSize: int64,
			}

			if err := stream.Send(fileData); err != nil {
				log.Printf("Client %d error sending file data: %v\n", id, err)
				return
			}
			response, err := stream.Recv()
			if err != nil {
				log.Printf("Client %d error receiving response: %v\n", id, err)
				return
			}
			log.Printf("Client %d received acknowledgment for chunk: ID=%s, Offset=%d\n",
				id, response.Id, response.Offset)
			log.Printf("Sent %d bytes", n)
		}
		if err := stream.CloseSend(); err != nil {
			log.Printf("Client %d error closing stream: %v\n", id, err)
		}
		if readErr == io.EOF {
			log.Println("Reached end of file")
			break
		}

		if readErr != nil {
			log.Fatalf("Error reading local file: %v", readErr)
			return
		}

	}
}

func saveToFile(filename string, data []byte) error {
	path, _ := filepath.Abs(fmt.Sprintf("./%s", directory))
	filePath := filepath.Join(path, filename)

	file, openErr := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		log.Printf("Failed to open local file %s: %v", filename, openErr)
		return openErr
	}
	defer file.Close()

	if _, writeErr := file.Write(data); writeErr != nil {
		log.Printf("Failed writing data to local file %s: %v", filename, writeErr)
		return writeErr
	}

	log.Printf("Successfully saved data to %s", filename)
	return nil
}
