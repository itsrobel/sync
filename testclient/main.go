package main

import (
	"context"
	"io"
	"log"
	"os"
	"watcher/filetransfer" // Replace with the actual path to the generated Go package

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := filetransfer.NewFileServiceClient(conn)
	// Request the file
	stream, err := c.SendFile(context.Background(), &filetransfer.FileRequest{Filename: "example.txt"})
	if err != nil {
		log.Fatalf("could not request file: %v", err)
	}

	outFile, err := os.Create("received_example.txt")
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer outFile.Close()
	// Receive the file content and write to the file
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			// End of file
			break
		}
		if err != nil {
			log.Fatalf("error while receiving file: %v", err)
		}
		// Write the received chunk to the file
		if _, err := outFile.Write(res.Data); err != nil {
			log.Fatalf("failed to write to file: %v", err)
		}
	}
	log.Println("File received and saved to received_example.txt")
}
