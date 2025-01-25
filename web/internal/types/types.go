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
	Id       string `bson:"id"`
	Location string `bson:"location"`
	Content  string `bson:"contents"`
	Active   bool   `bson:"active"` // this can decide whether or not to sync
}

// Every Hour if changes have been made create a new Version
// Shouldn't the file just point to the latest version?
type FileVersion struct {
	Id        string    `bson:"id"`
	Client    string    `bson:"client"`
	Timestamp time.Time `bson:"timestamp"` // Time when this version was created
	Location  string    `bson:"location"`  // File location
	Content   string    `bson:"contents"`  // Full contents of the file at this version
	FileId    string    `bson:"file_id"`   // Unique ID for the file
}
