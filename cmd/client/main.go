package main

import (
	"context"
	"fmt"
	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	ct "github.com/itsrobel/sync/internal/types"
	"github.com/itsrobel/sync/internal/watcher"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// test_watcher()
	test_on_start()
}

func test_on_start() {
	file, err := os.Open("./content")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	names, err := file.Readdirnames(0)
	if err != nil {
		log.Fatal(err)
	}
	for _, name := range names {
		fmt.Println(name)
	}

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
func test_connect() {

	filePath := "example.txt"
	file, openErr := os.Open(filePath)
	id := 1
	if openErr != nil {
		log.Fatalf("Failed to open local file: %v", openErr)
		return
	}
	defer file.Close()
	buf := make([]byte, ct.ChunkSize) // Define your buffer size

	client := filetransferconnect.NewFileServiceClient(http.DefaultClient, "http://localhost:50051")
	stream := client.SendFileToServer(context.Background())

	for {
		log.Printf("Trying to upload...")
		n, readErr := file.Read(buf) // Read from file into buffer

		if n > 0 { // Only send if there's data to send
			fileData := &ft.FileData{
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
			log.Printf("Sent %d bytes", n)
			res, _ := stream.CloseAndReceive()
			log.Printf("Server response: %v", res)
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
