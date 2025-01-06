package watcher

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"github.com/itsrobel/sync/internal/services/filetransfer/filetransferconnect"
)

type ServerMonitor struct {
	client      filetransferconnect.FileServiceClient
	isConnected atomic.Bool
	ticker      *time.Ticker
	done        chan bool
	interval    time.Duration
}

func NewServerMonitor(client filetransferconnect.FileServiceClient, interval time.Duration) *ServerMonitor {
	return &ServerMonitor{
		client:   client,
		done:     make(chan bool),
		interval: interval,
	}
}

func (sm *ServerMonitor) validateServer() bool {
	req := connect.NewRequest(&ft.ActionResponse{})
	_, err := sm.client.ValidateServer(context.Background(), req)
	if err != nil {
		sm.isConnected.Store(false)
		log.Print("Monitor: Server not connected")
		return false
	}
	sm.isConnected.Store(true)
	log.Print("Monitor: Server is connected")
	return true
}

func (sm *ServerMonitor) Start() {
	// Run immediately on start
	sm.validateServer()

	// Then start the ticker
	sm.ticker = time.NewTicker(sm.interval)
	go func() {
		for {
			select {
			case <-sm.ticker.C:
				if sm.validateServer() {
					sm.ticker.Stop()
					return

				}
			case <-sm.done:
				sm.ticker.Stop()
				return
			}
		}
	}()
}

func (sm *ServerMonitor) Stop() {
	sm.done <- true
}

func (sm *ServerMonitor) IsConnected() bool {
	return sm.isConnected.Load()
}
