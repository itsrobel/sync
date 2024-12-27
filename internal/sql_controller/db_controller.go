package db_controller

import (
	"database/sql"
	"fmt"
	// "github.com/huandu/go-sqlbuilder"
	ct "github.com/itsrobel/sync/internal/types"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

func BuildCreateTableStmt(structType interface{}, tableName string) string {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var columns []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		columnName := field.Name
		tag := field.Tag.Get("bson")
		if tag != "" {
			columnName = tag
		}

		sqlType := getSQLType(field.Type.Kind())
		column := fmt.Sprintf("%s %s", strings.ToLower(columnName), sqlType)

		if field.Name == "ID" {
			column += " PRIMARY KEY DEFAULT (lower(hex(randomblob(16))))"
		}

		columns = append(columns, column)
	}
	log.Print("Prepared Table: ", tableName)

	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t%s\n);",
		tableName,
		strings.Join(columns, ",\n\t"))
}

func getSQLType(k reflect.Kind) string {
	switch k {
	case reflect.String:
		return "TEXT"
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Int, reflect.Int32, reflect.Int64:
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		return "REAL"
	default:
		return "TEXT"
	}
}

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
		content TEXT,
		location TEXT,
		type TEXT,
		position INTEGER,
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

	// createTableFiles := BuildCreateTableStmt(ct.File{}, "files")
	// createTableFileVersions := BuildCreateTableStmt(ct.FileVersion{}, "file_versions")
	// createTableFileChanges := BuildCreateTableStmt(ct.FileChange{}, "file_changes")
	// db.Exec(createTableFiles)
	// db.Exec(createTableFileVersions)
	// db.Exec(createTableFileChanges)

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

func CreateFileVersion(db *sql.DB, fileID, newContent string) error {
	file, err := findFileById(db, fileID)
	if err != nil {
		return err
	}

	// Get the latest version
	lastVersion, err := getLatestVersion(db, fileID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Calculate changes from the last version
	var changes []ct.FileChange
	if lastVersion != nil {
		changes = getChangesBetweenVersions(lastVersion.Contents, newContent)
	} else {
		// If this is the first version, treat all lines as additions
		changes = getInitialChanges(newContent)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create new version
	versionID, err := storeFileVersion(tx, file.Location, newContent, fileID)
	if err != nil {
		return err
	}

	// Store changes associated with this version
	for _, change := range changes {
		if err := storeChange(tx, versionID, change); err != nil {
			return err
		}
	}

	// Update current file content
	if err := updateFileContent(tx, fileID, newContent); err != nil {
		return err
	}

	return tx.Commit()
}

func storeFileVersion(tx *sql.Tx, location, contents, fileID string) (string, error) {
	stmt, err := tx.Prepare(`
        INSERT INTO file_versions(id, timestamp, location, contents, file_id)
        VALUES(lower(hex(randomblob(16))), ?, ?, ?, ?)
        RETURNING id
    `)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var versionID string
	err = stmt.QueryRow(time.Now(), location, contents, fileID).Scan(&versionID)
	if err != nil {
		return "", err
	}

	return versionID, nil
}

func storeChange(tx *sql.Tx, versionID string, change ct.FileChange) error {
	stmt, err := tx.Prepare(`
        INSERT INTO file_changes(type, content, position, version_id)
        VALUES(lower(hex(randomblob(16))), ?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(change.Type, change.Content, change.Position, versionID)
	return err
}

func getLatestVersion(db *sql.DB, fileID string) (*ct.FileVersion, error) {
	var version ct.FileVersion
	err := db.QueryRow(`
        SELECT id, timestamp, location, contents, file_id 
        FROM file_versions 
        WHERE file_id = ? 
        ORDER BY timestamp DESC 
        LIMIT 1
    `, fileID).Scan(&version.FileID, &version.Timestamp, &version.Location, &version.Contents, &version.FileID)

	if err != nil {
		return nil, err
	}

	// Get changes associated with this version
	// changes, err := getVersionChanges(db, version.FileID)
	// if err != nil {
	// 	return nil, err
	// }
	// version.Changes = changes

	return &version, nil
}

func getVersionChanges(db *sql.DB, versionID string) ([]ct.FileChange, error) {
	rows, err := db.Query(`
        SELECT type, content, position 
        FROM file_changes 
        WHERE version_id = ? 
        ORDER BY position
    `, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []ct.FileChange
	for rows.Next() {
		var change ct.FileChange
		if err := rows.Scan(&change.Type, &change.Content, &change.Position); err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	return changes, rows.Err()
}

func updateFileContent(tx *sql.Tx, fileID, newContent string) error {
	_, err := tx.Exec("UPDATE files SET contents = ? WHERE id = ?", newContent, fileID)
	return err
}

func getInitialChanges(content string) []ct.FileChange {
	lines := strings.Split(content, "\n")
	changes := make([]ct.FileChange, len(lines))
	for i, line := range lines {
		changes[i] = ct.FileChange{
			Type:     "add",
			Content:  line,
			Position: i,
		}
	}
	return changes
}

func getChangesBetweenVersions(oldContent, newContent string) []ct.FileChange {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	changes := []ct.FileChange{}
	position := 0

	for _, line := range newLines {
		if position >= len(oldLines) {
			changes = append(changes, ct.FileChange{
				Type:     "add",
				Content:  line,
				Position: position,
			})
		} else if line != oldLines[position] {
			if containsLine(oldLines[position:], line) {
				changes = append(changes, ct.FileChange{
					Type:     "add",
					Content:  line,
					Position: position,
				})
			} else if containsLine(newLines[position:], oldLines[position]) {
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

	for i := position; i < len(oldLines); i++ {
		changes = append(changes, ct.FileChange{
			Type:     "remove",
			Content:  oldLines[i],
			Position: i,
		})
	}

	return changes
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
