// package main
//
// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	pb "watcher/filetransfer"
//
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/credentials/insecure"
// )
//
// const (
// 	directory = "content"
// 	chunkSize = 64 * 1024
// ) // Chunk size for streaming files.
//
// func main() {
// 	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("failed to connect: %v", err)
// 	}
// 	defer conn.Close()
//
// 	client := pb.NewFileServiceClient(conn)
// 	stream, err := client.TransferFile(context.Background())
// 	if err != nil {
// 		log.Fatalf("error creating stream: %v", err)
// 	}
//
// 	// Simulate sending file chunks
// 	for i := 0; i < 5; i++ {
// 		fileData := &pb.FileData{
// 			Id:        "file123",
// 			Content:   []byte("sample content"),
// 			Location:  "/path/to/file",
// 			Offset:    int64(i * 1024),
// 			TotalSize: 5 * 1024,
// 		}
//
// 		if err := stream.Send(fileData); err != nil {
// 			log.Fatalf("error sending file data: %v", err)
// 		}
//
// 		response, err := stream.Recv()
// 		if err != nil {
// 			log.Fatalf("error receiving response: %v", err)
// 		}
//
// 		log.Printf("Received acknowledgment for chunk: ID=%s, Offset=%d\n",
// 			response.Id, response.Offset)
//
// 		// time.Sleep(time.Second) // Simulate processing time
// 	}
//
// 	if err := stream.CloseSend(); err != nil {
// 		log.Fatalf("error closing stream: %v", err)
// 	}
// }

// func main() {
// 	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("Failed to connect: %v", err)
// 	}
// 	defer conn.Close()
//
// 	ctx := context.Background()
// 	client := pb.NewFileServiceClient(conn)
// 	stream, err := client.StreamFiles(ctx)
// 	// stream, err := client.StreamFiles(context.Background())
// 	if err != nil {
// 		log.Fatalf("Failed to start stream: %v", err)
// 	}
//
// 	// get all the files in the directory
// 	path, _ := filepath.Abs(fmt.Sprintf("./%s", directory))
// 	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		log.Println(path)
// 		return nil
// 	})
//
// 	filePath := filepath.Join(path, "t.md")
// 	uploadFile(stream, filePath)
// 	// log.Printf("Finished uploading file")
//
// 	// NOTE: I cant run go func and the watch files in the same runtime
// 	// path, _ := filepath.Abs(fmt.Sprintf("./%s", directory))
// 	// watchFiles(path)
//
// 	// go func() { // Listen for incoming updates from server.
// 	log.Println("Connected to server")
// 	for {
// 		in, recvErr := stream.Recv()
// 		if recvErr == io.EOF {
// 			break
// 		}
// 		if recvErr != nil {
// 			log.Fatalf("Error receiving data from server: %v", recvErr)
// 		}
//
// 		log.Printf("Received updated file: %s (%d bytes)", in.Location, len(in.Content))
// 		saveToFile(in.Location, in.Content)
// 	}
// 	// }()
// 	// select {} // Keep running indefinitely.
// }

// Upload a local file using bidirectional streaming.
// func uploadFile(stream pb.FileService_StreamFilesClient, filePath string) error {
// 	log.Println(filePath)
// 	file, openErr := os.Open(filePath)
// 	if openErr != nil {
// 		log.Fatalf("Failed to open local file: %v", openErr)
// 		return openErr
// 	}
// 	defer file.Close()
// 	buf := make([]byte, chunkSize) // Define your buffer size
//
// 	for {
// 		log.Printf("Trying to upload...")
//
// 		n, readErr := file.Read(buf) // Read from file into buffer
//
// 		if n > 0 { // Only send if there's data to send
// 			sendErr := stream.Send(&pb.FileData{
// 				Location: filepath.Base(filePath), // Use actual filename here
// 				Content:  buf[:n],                 // Send only n bytes
// 				Offset:   int64(n),
// 			})
// 			if sendErr != nil {
// 				log.Fatalf("Failed sending chunk data to server: %v", sendErr)
// 				return sendErr
// 			}
// 			log.Printf("Sent %d bytes", n)
// 		}
// 		closeErr := stream.CloseSend()
// 		if closeErr != nil {
// 			log.Printf("Failed to close send stream: %v", closeErr)
// 		}
//
// 		if readErr == io.EOF {
// 			log.Println("Reached end of file")
// 			break
// 		}
//
// 		if readErr != nil {
// 			log.Fatalf("Error reading local file: %v", readErr)
// 			return readErr
// 		}
// 	}
//
// 	return nil
// }
//
// // Save received data into a local file.
// func saveToFile(filename string, data []byte) error {
// 	path, _ := filepath.Abs(fmt.Sprintf("./%s", directory))
// 	filePath := filepath.Join(path, filename)
//
// 	file, openErr := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if openErr != nil {
// 		log.Printf("Failed to open local file %s: %v", filename, openErr)
// 		return openErr
// 	}
// 	defer file.Close()
//
// 	if _, writeErr := file.Write(data); writeErr != nil {
// 		log.Printf("Failed writing data to local file %s: %v", filename, writeErr)
// 		return writeErr
// 	}
//
// 	log.Printf("Successfully saved data to %s", filename)
// 	return nil
// }

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	pb "watcher/filetransfer"

	"google.golang.org/grpc"
)

func runClient(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Printf("Client %d failed to connect: %v\n", id, err)
		return
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)
	stream, err := client.TransferFile(context.Background())
	if err != nil {
		log.Printf("Client %d error creating stream: %v\n", id, err)
		return
	}

	// Simulate sending file chunks
	for i := 0; i < 5; i++ {
		fileData := &pb.FileData{
			Id:        fmt.Sprintf("file_%d_%d", id, i),
			Content:   []byte("sample content"),
			Location:  fmt.Sprintf("/path/to/file_%d", id),
			Offset:    int64(i * 1024),
			TotalSize: 5 * 1024,
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

		time.Sleep(time.Second) // Simulate processing time
	}

	if err := stream.CloseSend(); err != nil {
		log.Printf("Client %d error closing stream: %v\n", id, err)
	}
}

func main() {
	numClients := 3
	var wg sync.WaitGroup

	// Start multiple clients
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go runClient(i, &wg)
		time.Sleep(time.Millisecond * 100) // Stagger client starts
	}

	wg.Wait()
}
