package db_controller

import (
	"database/sql"
	"fmt"
	ct "github.com/itsrobel/sync/internal/types"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func initializeTables(db *sql.DB) error {
	createFileTable := `
	    CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		location TEXT UNIQUE,
		contents TEXT,
		active BOOLEAN
	    );
	`

	createFileVersionTable := `
	    CREATE TABLE IF NOT EXISTS file_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME,
		location TEXT,
		contents TEXT,
		file_id INTEGER,
		FOREIGN KEY(file_id) REFERENCES files(id)
	    );
	`

	createFileChangeTable := `
	    CREATE TABLE IF NOT EXISTS file_changes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME,
		content_change TEXT,
		location TEXT,
		version_id INTEGER,
		FOREIGN KEY(version_id) REFERENCES file_versions(id)
	    );
	`

	_, err := db.Exec(createFileTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(createFileVersionTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(createFileChangeTable)
	return err
}
func ConnectSQLite(dbPath string) (*sql.DB, error) {
	// Set default database path if none provided
	if dbPath == "" {
		dbPath = "./sync.db"
	}

	// Create database directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection (SQLite will create the file if it doesn't exist)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize tables
	if err = initializeTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	log.Printf("Connected to SQLite database at %s", dbPath)
	return db, nil
}

func CreateFile(db *sql.DB, location string) (string, error) {
	stmt, err := db.Prepare("INSERT INTO files(location, active, contents) VALUES(?, ?, ?) RETURNING id")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var id string
	err = stmt.QueryRow(location, true, "").Scan(&id)
	if err != nil {
		return "", err
	}

	log.Printf("Inserted file with ID: %s at %s", id, location)
	return id, nil
}

func CreateFileVersion(db *sql.DB, fileID string) error {
	file, err := findFile(db, fileID)
	if err != nil {
		return err
	}

	stmt, err := db.Prepare(`
        INSERT INTO file_versions(timestamp, location, contents, file_id) 
        VALUES(?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), file.Location, file.Contents, fileID)
	return err
}

func findFile(db *sql.DB, id string) (*ct.File, error) {
	var file ct.File
	err := db.QueryRow("SELECT id, location, contents, active FROM files WHERE id = ?", id).
		Scan(&file.ID, &file.Location, &file.Contents, &file.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no file found with id: %s", id)
		}
		return nil, err
	}
	return &file, nil
}

func ValidFileExtension(location string) bool {
	extensions := []string{".md", ".pdf"}
	for _, ext := range extensions {
		if strings.HasSuffix(location, ext) {
			return true
		}
	}
	return false
}

func GetAllDocuments(db *sql.DB) ([]ct.File, error) {
	rows, err := db.Query("SELECT id, location, contents, active FROM files")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []ct.File
	for rows.Next() {
		var file ct.File
		if err := rows.Scan(&file.ID, &file.Location, &file.Contents, &file.Active); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

func deleteAllDocuments(db *sql.DB) (int64, error) {
	result, err := db.Exec("DELETE FROM files")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
