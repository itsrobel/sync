package db_controller

import (
	"database/sql"
	"fmt"
	ct "github.com/itsrobel/sync/internal/types"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func initializeTables(db *sql.DB) error {
	createFileTable := `
	    CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		location TEXT UNIQUE,
		contents TEXT,
		active BOOLEAN
	    );
	`

	createFileVersionTable := `
	    CREATE TABLE IF NOT EXISTS file_versions (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		timestamp DATETIME,
		location TEXT,
		contents TEXT,
		file_id TEXT,
		FOREIGN KEY(file_id) REFERENCES files(id)
	    );
	`

	createFileChangeTable := `
	    CREATE TABLE IF NOT EXISTS file_changes (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		timestamp DATETIME,
		content_change TEXT,
		location TEXT,
		version_id INTEGER,
		FOREIGN KEY(version_id) REFERENCES file_versions(id)
	    );
	`

	log.Println("Create File Table")
	_, err := db.Exec(createFileTable)
	if err != nil {
		return err
	}

	log.Println("Create FileVersion Table")
	_, err = db.Exec(createFileVersionTable)
	if err != nil {
		return err
	}

	log.Println("Create FileChange Table")
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

// func CreateFileVersion(db *sql.DB, fileID, newContent string) error {
// 	file, err := findFileById(db, fileID)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Get current and new content as lines
// 	currentLines := strings.Split(file.Contents, "\n")
// 	newLines := strings.Split(newContent, "\n")
//
// 	// Calculate diff using unified diff format
// 	diff := difflib.UnifiedDiff{
// 		A:        currentLines,
// 		B:        newLines,
// 		FromFile: "previous",
// 		ToFile:   "current",
// 		Context:  3,
// 	}
//
// 	diffText, err := difflib.GetUnifiedDiffString(diff)
// 	if err != nil {
// 		return fmt.Errorf("failed to generate diff: %w", err)
// 	}
//
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()
//
// 	// Update the current file contents
// 	_, err = tx.Exec("UPDATE files SET contents = ? WHERE id = ?", newContent, fileID)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Store the diff in the version
// 	stmt, err := tx.Prepare(`
//         INSERT INTO file_versions(timestamp, location, contents, file_id)
//         VALUES(?, ?, ?, ?)
// 	`)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()
//
// 	_, err = stmt.Exec(time.Now(), file.Location, diffText, fileID)
// 	if err != nil {
// 		return err
// 	}
//
// 	return tx.Commit()
// }

func CreateFileVersion(db *sql.DB, fileID, newContent string) error {
	file, err := findFileById(db, fileID)
	if err != nil {
		return err
	}

	// Get changes between versions
	changes := getChangesBetweenVersions(file.Contents, newContent)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update current file content
	_, err = tx.Exec("UPDATE files SET contents = ? WHERE id = ?", newContent, fileID)
	if err != nil {
		return err
	}

	// Store each change
	for _, change := range changes {
		err = storeChange(tx, fileID, change)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func getChangesBetweenVersions(oldContent, newContent string) []ct.FileChange {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	changes := []ct.FileChange{}
	position := 0

	for _, line := range newLines {
		if position >= len(oldLines) {
			// New line added at the end
			changes = append(changes, ct.FileChange{
				Type:     "add",
				Content:  line,
				Position: position,
			})
		} else if line != oldLines[position] {
			if containsLine(oldLines[position:], line) {
				// Line was added
				changes = append(changes, ct.FileChange{
					Type:     "add",
					Content:  line,
					Position: position,
				})
			} else if containsLine(newLines[position:], oldLines[position]) {
				// Line was removed
				changes = append(changes, ct.FileChange{
					Type:     "remove",
					Content:  oldLines[position],
					Position: position,
				})
				position--
			}
		}
		position++
	}

	// Check for remaining lines that were removed
	for i := position; i < len(oldLines); i++ {
		changes = append(changes, ct.FileChange{
			Type:     "remove",
			Content:  oldLines[i],
			Position: i,
		})
	}

	return changes
}

func applyChanges(content string, changes []ct.FileChange) string {
	lines := strings.Split(content, "\n")

	// Sort changes by position in reverse order to avoid position shifts
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Position > changes[j].Position
	})

	for _, change := range changes {
		if change.Type == "add" {
			// Insert new line
			if change.Position >= len(lines) {
				lines = append(lines, change.Content)
			} else {
				lines = append(lines[:change.Position], append([]string{change.Content}, lines[change.Position:]...)...)
			}
		} else if change.Type == "remove" {
			// Remove line
			if change.Position < len(lines) {
				lines = append(lines[:change.Position], lines[change.Position+1:]...)
			}
		}
	}

	return strings.Join(lines, "\n")
}

func storeChange(tx *sql.Tx, fileID string, change ct.FileChange) error {
	stmt, err := tx.Prepare(`
        INSERT INTO file_changes(type, content, position, file_id)
        VALUES(?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(change.Type, change.Content, change.Position, fileID)
	return err
}

func containsLine(lines []string, line string) bool {
	for _, l := range lines {
		if l == line {
			return true
		}
	}
	return false
}

func findFileById(db *sql.DB, id string) (*ct.File, error) {
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

func FindFileByLocation(db *sql.DB, location string) (*ct.File, error) {
	var file ct.File
	err := db.QueryRow("SELECT id, location, contents, active FROM files WHERE location = ?", location).
		Scan(&file.ID, &file.Location, &file.Contents, &file.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no file found with location: %s", location)
		}
		return nil, err
	}
	log.Printf("Found File By Location: %s", location)
	return &file, nil
}

func FindFileByParam(db *sql.DB, param, param_value string) (*ct.File, error) {
	var file ct.File
	err := db.QueryRow("SELECT id, location, contents, active FROM files WHERE ? = ?", param, param_value).
		Scan(&file.ID, &file.Location, &file.Contents, &file.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no file found with %s: %s", param, param_value)
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

func GetAllFiles(db *sql.DB) ([]ct.File, error) {
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

func GetAllFileVersions(db *sql.DB, fileID string) ([]ct.FileVersion, error) {
	log.Println("file versions for file: ", fileID)
	rows, err := db.Query("SELECT  contents, file_id FROM file_versions where file_id = ?", fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []ct.FileVersion
	for rows.Next() {
		var file ct.FileVersion
		if err := rows.Scan(&file.Contents, &file.FileID); err != nil {
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
