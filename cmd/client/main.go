package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	ct "github.com/itsrobel/sync/internal/types"
)

func main() {
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
