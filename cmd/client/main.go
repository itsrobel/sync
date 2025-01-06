package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	// "fmt"

	// ct "github.com/itsrobel/sync/internal/types"
	"github.com/itsrobel/sync/internal/watcher"
	// "io"
	// "path/filepath"
)

func main() {
	test_watcher()
	// test_connect("./content/t.md")
}

func test_watcher() {
	fw, err := watcher.InitFileWatcher("", "./content")
	if err != nil {
		log.Fatal(err)
	}
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fw.Stop()
	}()
}

// func test_connect(file_path string) {
// 	file, openErr := os.Open(file_path)
// 	// id := 1
// 	if openErr != nil {
// 		log.Fatalf("Failed to open local file: %v", openErr)
// 		return
// 	}
// 	defer file.Close()
// 	// buf := make([]byte, ct.ChunkSize) // Define your buffer size
//
// 	client := filetransferconnect.NewFileServiceClient(http.DefaultClient, "http://localhost:50051")
//
// 	// TODO: make a ping request on file service request
// 	// Create a go routinue for the ping request
//
// 	req := connect.NewRequest(&ft.ActionResponse{
// 		Success: true,
// 		Message: "client-1",
// 	})
//
// 	resp, err := client.ValidateServer(context.Background(), req)
// 	if err != nil {
// 		log.Fatal("this is a error", err)
// 	}
//
// 	// Access the response
// 	result := resp.Msg // This will be your ActionResponse
// 	log.Println(result)

// we will be transfering over fileversions mostly

// stream := client.SendFileToServer(context.Background())
// for {
// 	log.Printf("Trying to upload...")
// 	n, readErr := file.Read(buf) // Read from file into buffer
//
// 	if n > 0 { // Only send if there's data to send
// 		fileData := &ft.FileData{
// 			Id:       fmt.Sprintf("%d", id),
// 			Location: filepath.Base(file_path), // Use actual filename here
// 			Content:  buf[:n],                  // Send only n bytes
// 			Offset:   int64(n),
// 			// TotalSize: int64,
// 		}
//
// 		if err := stream.Send(fileData); err != nil {
// 			log.Printf("Client %d error sending file data: %v\n", id, err)
// 			return
// 		}
// 		log.Printf("Sent %d bytes", n)
// 		res, _ := stream.CloseAndReceive()
// 		log.Printf("Server response: %v", res)
// 	}
//
// 	if readErr == io.EOF {
// 		log.Println("Reached end of file")
// 		break
// 	}
//
// 	if readErr != nil {
// 		log.Fatalf("Error reading local file: %v", readErr)
// 		return
// 	}
//
// }
// }
