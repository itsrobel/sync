package sql_manager

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	Directory = "content"
	ChunkSize = 64 * 1024
)

type FileBase struct {
	ID string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
}

type File struct {
	FileBase
	Active   bool
	Location string
	Content  string
}
type FileVersion struct {
	FileBase
	Timestamp time.Time
	Client    string
	Location  string
	Content   string
	FileID    string `gorm:"type:uuid"`
}

type ClientSession struct {
	SessionID    string `gorm:"primaryKey"`
	LastSyncTime time.Time
	IsActive     bool
}

func (f *FileBase) BeforeCreate(tx *gorm.DB) (err error) {
	f.ID = uuid.New().String()
	return
}
