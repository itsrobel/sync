package sql_manager

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectSQLite(dbPath string) (*gorm.DB, error) {
	if dbPath == "" {
		dbPath = "./sync-orm.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Auto Migrate the schema
	err = db.AutoMigrate(&File{}, &FileVersion{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func ConnectPostgres() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=postgres dbname=myapp port=5432 sslmode=disable"

	// Open connection with retry logic
	var db *gorm.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info), // Add logging
		})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database, retrying in 5 seconds... (attempt %d/5)", i+1)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after 5 attempts: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return nil, fmt.Errorf("failed to create uuid extension: %v", err)
	}

	// Drop existing tables
	if err := db.Migrator().DropTable(&FileVersion{}); err != nil {
		return nil, fmt.Errorf("failed to drop tables: %v", err)
	}

	// if err := db.Migrator().DropIndex(&File{}, "idx_files_location"); err != nil {
	// 	return nil, fmt.Errorf("failed to drop tables: %v", err)
	// }

	// Migrate all models at once
	if err := db.AutoMigrate(
		&ClientSession{},
		&File{},
		&FileVersion{},
		// Add other models here
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}
