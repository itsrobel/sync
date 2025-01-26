package manager

import (
	"fmt"
	"log"
	"time"

	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	ct "github.com/itsrobel/sync/internal/types"
	"gorm.io/gorm"
)

// Models
type File struct {
	ID       string `gorm:"primaryKey"`
	Location string `gorm:"uniqueIndex;not null"`
	Content  string
	Active   bool
	// CreatedAt time.Time
	// UpdatedAt time.Time
	Timestamp time.Time
}

type FileVersion struct {
	ID        uint   `gorm:"primaryKey"`
	FileID    string `gorm:"index"`
	Location  string
	Content   string
	Timestamp time.Time
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&File{}, &FileVersion{})
}

func CreateFileVersion(db *gorm.DB, file *ft.FileVersionData) error {
	fileVersion := FileVersion{
		Timestamp: file.Timestamp.AsTime(),
		Location:  file.Location,
		Content:   string(file.Content),
		FileID:    file.FileId,
	}

	result := db.Create(&fileVersion)
	if result.Error != nil {
		log.Printf("Error creating file version: %v", result.Error)
		return result.Error
	}

	log.Printf("Created file version with ID: %d at %s", fileVersion.ID, file.Location)
	return nil
}

func UpdateFile(db *gorm.DB, file *ct.File) error {
	result := db.Model(&File{}).
		Where("id = ?", file.ID).
		Updates(map[string]interface{}{
			"location": file.Location,
			"content":  file.Content,
			"active":   file.Active,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update file: %v", result.Error)
	}

	log.Printf("Updated file with ID: %s at %s", file.ID, file.Location)
	return nil
}

func FindFileByLocation(db *gorm.DB, location string) (*ct.File, error) {
	var file File
	result := db.Where("location = ?", location).First(&file)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no file found with location: %s", location)
		}
		return nil, fmt.Errorf("error finding file: %v", result.Error)
	}

	return &ct.File{
		ID:       file.ID,
		Location: file.Location,
		Content:  file.Content,
		Active:   file.Active,
	}, nil
}

func FindFileById(db *gorm.DB, fileID string) (*ct.File, error) {
	var file File
	result := db.First(&file, "id = ?", fileID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no file found with ID: %s", fileID)
		}
		return nil, fmt.Errorf("error finding file: %v", result.Error)
	}

	return &ct.File{
		ID:       file.ID,
		Location: file.Location,
		Content:  file.Content,
		Active:   file.Active,
	}, nil
}

func GetAllDocuments(db *gorm.DB) ([]ct.File, error) {
	var files []File
	result := db.Find(&files)

	if result.Error != nil {
		return nil, fmt.Errorf("error fetching files: %v", result.Error)
	}

	var documents []ct.File
	for _, file := range files {
		documents = append(documents, ct.File{
			ID:       file.ID,
			Location: file.Location,
			Content:  file.Content,
			Active:   file.Active,
		})
	}

	return documents, nil
}

func DeleteAllDocuments(db *gorm.DB) (int64, error) {
	result := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&File{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete documents: %v", result.Error)
	}

	return result.RowsAffected, nil
}
