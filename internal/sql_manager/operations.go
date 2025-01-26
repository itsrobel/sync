package sql_manager

import (
	"time"

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
