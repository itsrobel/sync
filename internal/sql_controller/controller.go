package controller

import (
	"database/sql"
	"fmt"
	ct "github.com/itsrobel/sync/internal/types"
	_ "github.com/mattn/go-sqlite3"
	"log"
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
    );`

	createFileVersionTable := `
    CREATE TABLE IF NOT EXISTS file_versions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        timestamp DATETIME,
        location TEXT,
        contents TEXT,
        file_id INTEGER,
        FOREIGN KEY(file_id) REFERENCES files(id)
    );`

	createFileChangeTable := `
    CREATE TABLE IF NOT EXISTS file_changes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        timestamp DATETIME,
        content_change TEXT,
        location TEXT,
        version_id INTEGER,
        FOREIGN KEY(version_id) REFERENCES file_versions(id)
    );`

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
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	if err = initializeTables(db); err != nil {
		return nil, err
	}

	log.Println("Connected to SQLite database!")
	return db, nil
}

func CreateFile(db *sql.DB, location string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO files(location, active, contents) VALUES(?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(location, true, "")
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("Inserted file with ID: %d at %s", id, location)
	return id, nil
}

func CreateFileVersion(db *sql.DB, fileID int64) error {
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

func ValidFileExtension(location string) bool {
	extensions := []string{".md", ".pdf"}
	for _, ext := range extensions {
		if strings.HasSuffix(location, ext) {
			return true
		}
	}
	return false
}

func findFile(db *sql.DB, id int64) (*ct.File, error) {
	var file ct.File
	err := db.QueryRow("SELECT id, location, contents, active FROM files WHERE id = ?", id).
		Scan(&file.ID, &file.Location, &file.Contents, &file.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no file found with id: %d", id)
		}
		return nil, err
	}
	return &file, nil
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
