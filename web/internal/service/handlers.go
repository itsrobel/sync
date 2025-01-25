package service

import (
	// "baby-backend/internal/services/apiv1"

	"context"
	"fmt"
	"net/http"

	"github.com/itsrobel/sync/web/internal/services/web/filetransfer"
	"github.com/itsrobel/sync/web/internal/services/web/filetransfer/filetransferconnect"
	"github.com/itsrobel/sync/web/internal/templates"

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
