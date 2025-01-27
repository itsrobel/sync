package sql_manager

import (
	"fmt"
	"log"
	"time"

	ft "github.com/itsrobel/sync/internal/services/filetransfer"
	"gorm.io/gorm"
)

func CreateFileInitial(db *gorm.DB, location string) (*File, error) {
	file := &File{
		Location: location,
		Active:   true,
		Content:  "",
	}

	result := db.Create(file)
	return file, result.Error
}

func CreateFileVersion(db *gorm.DB, file *File, newContent string) (*FileVersion, error) {
	fileVersion := &FileVersion{
		Timestamp: time.Now(),
		Location:  file.Location,
		Content:   newContent,
		FileID:    file.ID,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(fileVersion).Error; err != nil {
			return err
		}
		return tx.Model(file).Update("content", newContent).Error
	})

	return fileVersion, err
}

func CreateFileVersionServer(db *gorm.DB, file *ft.FileVersionData) error {
	fileVersion := FileVersion{
		// FileBase:  FileBase{ID: uuid.NewString()},
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

	log.Printf("Created file version with ID: %s at %s", fileVersion.ID, file.Location)
	return nil
}

func UpdateFileServer(db *gorm.DB, file *File) error {
	var existingFile File

	result := db.Where("id = ?", file.ID).First(&existingFile)
	if result.Error == gorm.ErrRecordNotFound {
		// Create new file if not found
		newFile := File{
			Location: file.Location,
			Content:  file.Content,
			Active:   file.Active,
		}
		if err := db.Create(&newFile).Error; err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		log.Printf("Created new file with ID: %s at %s", file.ID, file.Location)
		return nil
	} else if result.Error != nil {
		return fmt.Errorf("failed to query file: %v", result.Error)
	}

	// Update existing file
	result = db.Model(&existingFile).Updates(map[string]interface{}{
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

func GetLatestVersion(db *gorm.DB, fileID string) (*FileVersion, error) {
	var version FileVersion
	err := db.Where("file_id = ?", fileID).
		Order("timestamp desc").
		First(&version).Error
	return &version, err
}

func FindFileById(db *gorm.DB, id string) (*File, error) {
	var file File
	err := db.First(&file, "id = ?", id).Error
	return &file, err
}

// TODO: make sure to create before trying to find by location
func FindFileByLocation(db *gorm.DB, location string) (*File, error) {
	var file File
	err := db.First(&file, "location = ?", location).Error
	return &file, err
}

func GetAllFiles(db *gorm.DB) ([]File, error) {
	var files []File
	err := db.Find(&files).Error
	return files, err
}

func GetAllFileVersions(db *gorm.DB, fileID string) ([]FileVersion, error) {
	var versions []FileVersion
	err := db.Where("file_id = ?", fileID).Find(&versions).Error
	return versions, err
}

func DeleteAllDocuments(db *gorm.DB) error {
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&File{}).Error
}
