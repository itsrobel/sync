package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// TODO: rewrite the web server in gin
// TODO: make the client be cli or something?
// TODO: turn the file listen into a folder listen
const (
	// Time allowed to write the file to the client.
	writeWait = 3 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Poll file for changes with this period.
	filePeriod = 10 * time.Second
)

var (
	addr = flag.String("addr", ":8080", "http service address")
	// homeTempl = template.Must(template.New("").Parse(homeHTML))
	filename string
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
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

func watcherServer() {
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("filename not specified")
	}
	filename = flag.Args()[0]
	r := gin.Default()

	// NOTE: the current watcher structer works by having a client port
	// at the home that client then connects as a socket to /ws where
	// it then takes in the read and rewrites
	r.GET("/", serveWs)
	if err := r.Run(*addr); err != nil {
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

	var lastMod time.Time
	if n, err := strconv.ParseInt(c.Query("lastMod"), 16, 64); err == nil {
		lastMod = time.Unix(0, n)
	}

	go writer(ws, lastMod)
	reader(ws)
}
