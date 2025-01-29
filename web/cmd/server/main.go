package main

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"github.com/itsrobel/sync/web/internal/service"
	"github.com/itsrobel/sync/web/internal/services/web/filetransfer/filetransferconnect"

	"github.com/rs/cors"
	"golang.org/x/net/http2"
)

func main() {
	mux := http.NewServeMux()

	client := filetransferconnect.NewFileServiceClient(
		&http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		},
		"http://localhost:50051",
	)

	// Initialize handlers
	handlers := service.NewHandlers(client)

	// Routes
	mux.HandleFunc("/", handlers.Index)
	mux.HandleFunc("/edit", handlers.HandleEditor)
	mux.HandleFunc("/greet", handlers.HandleGreet)

	// Serve static files
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Configure CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Connect-Protocol-Version"},
	})

	wrappedHandler := corsHandler.Handler(mux)

	log.Println("Server starting on :3000")
	if err := http.ListenAndServe(":3000", wrappedHandler); err != nil {
		log.Fatal(err)
	}
}
