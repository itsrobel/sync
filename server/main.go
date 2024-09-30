package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	// "strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {
	syncServer()
}

var (
	addr = ":8080"

	// homeTempl = template.Must(template.New("").Parse(homeHTML))
	filename string
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func syncServer() {
	r := gin.Default()

	// NOTE: the current watcher structer works by having a client port
	// at the home that client then connects as a socket to /ws where
	// it t hen takes in the read and rewrites
	r.GET("/", serveWs)
	r.POST("/upload", fileUpload)
	r.GET("/emit/:id", fileEmit)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func serveWs(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	// var lastMod time.Time
	// if n, err := strconv.ParseInt(c.Query("lastMod"), 16, 64); err == nil {
	// 	lastMod = time.Unix(0, n)
	// }

	// go writer(ws, lastMod)
	// NOTE: the loop itself it handled by the function
	dirWatcher(ws)
	reader(ws)
}

// TODO: rewrite the web server in gin
// TODO: make the client be cli or something?
// TODO: turn the file listen into a folder listen
const (
	// Time allowed to write the file to the client.
	writeWait = 3 * time.Second
	directory = "content"
	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Poll file for changes with this period.
	filePeriod = 2 * time.Second
)

func readFileIfModified(lastMod time.Time) ([]byte, time.Time, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, lastMod, err
	}
	if !fi.ModTime().After(lastMod) {
		return nil, lastMod, nil
	}
	p, err := os.ReadFile(filename)
	if err != nil {
		return nil, fi.ModTime(), err
	}
	return p, fi.ModTime(), nil
}

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn, lastMod time.Time) {
	lastError := ""
	pingTicker := time.NewTicker(pingPeriod)
	fileTicker := time.NewTicker(filePeriod)
	defer func() {
		pingTicker.Stop()
		fileTicker.Stop()
		ws.Close()
	}()

	for {
		select {
		case <-fileTicker.C:
			var p []byte
			var err error

			p, lastMod, err = readFileIfModified(lastMod)

			if err != nil {
				if s := err.Error(); s != lastError {
					lastError = s
					p = []byte(lastError)
				}
			} else {
				lastError = ""
			}

			if p != nil {
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, p); err != nil {
					return
				}
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))

			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func fileUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "Get form err: %s", err.Error())
		return
	}
	filename := file.Filename
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.String(http.StatusInternalServerError, "Save file err: %s", err.Error())
		return
	}
	c.String(http.StatusOK, "File %s uploaded successfully.", filename)
}

// this function takes in the param of the file id and then returns it to the client
func fileEmit(c *gin.Context) {
	// TODO: make the id fetch from the db that then retrieves the location of the file and the name
	filename := c.Param("id")
	// Construct the full file path
	filePath := filepath.Join(directory, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Set header to force download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")

	// Serve the file
	c.File(filePath)
}
