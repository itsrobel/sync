package handlers

import (
	// "baby-backend/internal/services/apiv1"

	"context"
	"fmt"
	"net/http"

	"github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
	"github.com/itsrobel/sync/internal/templates"

	"connectrpc.com/connect"
)

type Handlers struct {
	greetClient filetransferconnect.FileServiceClient
}

func NewHandlers(greetClient filetransferconnect.FileServiceClient) *Handlers {
	return &Handlers{
		greetClient: greetClient,
	}
}

func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	// TODO: make a request to the backend and then pass that into templates.Index
	component := templates.Index()
	component.Render(context.Background(), w)
}

func (h *Handlers) HandleGreet(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	req := connect.NewRequest(&filetransfer.GreetRequest{Name: name})
	fmt.Println(req)

	resp, err := h.greetClient.Greet(r.Context(), req)
	fmt.Println(resp)

	if err != nil {
		http.Error(w, "Failed to greet", http.StatusInternalServerError)
		return
	}

	component := templates.GreetingResponse(resp.Msg.Greeting)
	component.Render(r.Context(), w)
}

// func (h *Handlers) HandleFiles(w http.ResponseWriter, r *http.Request) {
// }

func (h *Handlers) HandleEditor(w http.ResponseWriter, r *http.Request) {
	req := connect.NewRequest(&filetransfer.ActionRequest{})
	resp, _ := h.greetClient.RetrieveListOfFiles(r.Context(), req)
	fmt.Println(resp)

	component := templates.Editor()

	component.Render(context.Background(), w)
}

func (h *Handlers) HandleDraw(w http.ResponseWriter, r *http.Request) {
	component := templates.Excalidraw()

	// Render the component
	component.Render(context.Background(), w)
}
