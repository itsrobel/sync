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
	test()
}

func test() {
	// Set up paths
	dbPath := "./sync-test.db"
	watchPath := "./content-test"
	clientName := "test-2"

	// Ensure content directory exists
	if err := os.MkdirAll(watchPath, 0755); err != nil {
		log.Fatalf("Failed to create watch directory: %v", err)
	}

	// Initialize file watcher
	//
	fw, err := watcher.InitFileWatcher(dbPath, watchPath, clientName)
	if err != nil {
		log.Fatalf("Failed to initialize file watcher: %v", err)
	}
	defer fw.Stop()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("File watcher started. Watching directory: %s", watchPath)
	<-sigChan
	log.Println("Shutting down...")
}
