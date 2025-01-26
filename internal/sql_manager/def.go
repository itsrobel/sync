package sql_manager

import "time"

type File struct {
	ID       string `gorm:"primaryKey;type:text;default:lower(hex(randomblob(16)))"`
	Location string `gorm:"unique"`
	Content  string
	Active   bool
}

type FileVersion struct {
	ID        string `gorm:"primaryKey;type:text;default:lower(hex(randomblob(16)))"`
	Timestamp time.Time
	Location  string
	Content   string
	FileID    string
}
