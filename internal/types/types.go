package types

import (
	"time"
)

const (
	Directory = "content"
	ChunkSize = 64 * 1024
) // Chunk size for streaming files.

// NOTE: ID's are created by default
// time?
type File struct {
	ID       string
	Location string `bson:"location"`
	Contents string `bson:"contents"`
	Active   bool   `bson:"active"` // this can decide whether or not to sync
}

// Every Hour if changes have been made create a new Version
// Shouldn't the file just point to the latest version?
type FileVersion struct {
	ID        string
	Timestamp time.Time `bson:"timestamp"` // Time when this version was created
	Location  string    `bson:"location"`  // File location
	Contents  string    `bson:"contents"`  // Full contents of the file at this version
	FileID    string    `bson:"file_id"`   // Unique ID for the file
}

// TODO: when a file is change it can write a change log and then
// write to the file to update
type FileChange struct {
	Type      string // "add" or "remove"
	Content   string
	Position  int    // Line number where change occurred
	VersionID string `bson:"version_id"` // Unique ID for the file
}
