package types

import "time"

type ClientSession struct {
	SessionID    string    `bson:"session_id"`
	LastSyncTime time.Time `bson:"last_sync_time"`
	IsActive     bool      `bson:"is_active"`
}
