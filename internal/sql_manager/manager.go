package sql_manager

//
// import (
// 	"database/sql"
// 	"fmt"
//
// 	// "github.com/huandu/go-sqlbuilder"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"reflect"
// 	"strings"
// 	"time"
//
// 	ct "github.com/itsrobel/sync/internal/types"
// 	_ "github.com/mattn/go-sqlite3"
// )
//
// func BuildCreateTableStmt(structType interface{}, tableName string) string {
// 	t := reflect.TypeOf(structType)
// 	if t.Kind() == reflect.Ptr {
// 		t = t.Elem()
// 	}
//
// 	var columns []string
// 	for i := 0; i < t.NumField(); i++ {
// 		field := t.Field(i)
// 		columnName := field.Name
// 		tag := field.Tag.Get("bson")
// 		if tag != "" {
// 			columnName = tag
// 		}
//
// 		sqlType := getSQLType(field.Type.Kind())
// 		column := fmt.Sprintf("%s %s", strings.ToLower(columnName), sqlType)
//
// 		if field.Name == "ID" {
// 			column += " PRIMARY KEY DEFAULT (lower(hex(randomblob(16))))"
// 		}
//
// 		columns = append(columns, column)
// 	}
// 	log.Print("Prepared Table: ", tableName)
//
// 	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\t%s\n);",
// 		tableName,
// 		strings.Join(columns, ",\n\t"))
// }
//
// func getSQLType(k reflect.Kind) string {
// 	switch k {
// 	case reflect.String:
// 		return "TEXT"
// 	case reflect.Bool:
// 		return "BOOLEAN"
// 	case reflect.Int, reflect.Int32, reflect.Int64:
// 		return "INTEGER"
// 	case reflect.Float32, reflect.Float64:
// 		return "REAL"
// 	default:
// 		return "TEXT"
// 	}
// }
//
// func initializeTables(db *sql.DB) error {
// 	createFileTable := `
// 	  CREATE TABLE IF NOT EXISTS files (
// 		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
// 		location TEXT UNIQUE,
// 		content TEXT,
// 		active BOOLEAN
// 	  );
// 	`
//
// 	createFileVersionTable := `
// 	  CREATE TABLE IF NOT EXISTS file_versions (
// 		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
// 		timestamp DATETIME,
// 		location TEXT,
// 		content TEXT,
// 		file_id TEXT,
// 		FOREIGN KEY(file_id) REFERENCES files(id)
// 	    );
// 	`
//
// 	log.Println("Create File Table")
// 	_, err := db.Exec(createFileTable)
// 	if err != nil {
// 		return err
// 	}
//
// 	log.Println("Create FileVersion Table")
// 	_, err = db.Exec(createFileVersionTable)
// 	if err != nil {
// 		return err
// 	}
// 	return err
// }
//
// func ConnectSQLite(dbPath string) (*sql.DB, error) {
// 	// Set default database path if none provided
// 	if dbPath == "" {
// 		dbPath = "./sync.db"
// 	}
//
// 	// Create database directory if it doesn't exist
// 	dbDir := filepath.Dir(dbPath)
// 	if err := os.MkdirAll(dbDir, 0755); err != nil {
// 		return nil, fmt.Errorf("failed to create database directory: %w", err)
// 	}
//
// 	// Open database connection (SQLite will create the file if it doesn't exist)
// 	db, err := sql.Open("sqlite3", dbPath)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open database: %w", err)
// 	}
//
// 	// Verify connection
// 	if err = db.Ping(); err != nil {
// 		db.Close()
// 		return nil, fmt.Errorf("failed to ping database: %w", err)
// 	}
//
// 	if err = initializeTables(db); err != nil {
// 		db.Close()
// 		return nil, fmt.Errorf("failed to initialize tables: %w", err)
// 	}
//
// 	log.Printf("Connected to SQLite database at %s", dbPath)
// 	return db, nil
// }
//
// // return file object
// func CreateFileInitial(db *sql.DB, location string) (*ct.File, error) {
// 	file := ct.File{Location: location, Active: true, Content: ""}
// 	stmt, err := db.Prepare(`
//     INSERT INTO files(id, location, active, content)
//     VALUES(lower(hex(randomblob(16))),?, ?, ?) RETURNING id
//     `)
// 	if err != nil {
// 		return &file, err
// 	}
// 	defer stmt.Close()
//
// 	var id string
// 	err = stmt.QueryRow(file.Location, file.Active, file.Content).Scan(&id)
// 	if err != nil {
// 		return &file, err
// 	}
//
// 	log.Printf("Inserted file with ID: %s at %s", id, location)
// 	file.Id = id
//
// 	return &file, nil
// }
//
// func CreateFileVersion(db *sql.DB, file *ct.File, newContent string) (*ct.FileVersion, error) {
// 	fileVersion := ct.FileVersion{Timestamp: time.Now(), Location: file.Location, Content: newContent, FileId: file.Id}
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return &fileVersion, err
// 	}
// 	defer tx.Rollback()
// 	stmt, err := tx.Prepare(`
//         INSERT INTO file_versions(id, timestamp, location, content, file_id)
//         VALUES(lower(hex(randomblob(16))), ?, ?, ?, ?)
//         RETURNING id
//     `)
// 	if err != nil {
// 		return &fileVersion, err
// 	}
// 	defer stmt.Close()
//
// 	var versionID string
//
// 	err = stmt.QueryRow(fileVersion.Timestamp, file.Location, newContent, file.Id).Scan(&versionID)
// 	if err != nil {
// 		return &fileVersion, err
// 	}
// 	tx.Exec("UPDATE files SET content = ? WHERE id = ?", newContent, file.Id)
// 	tx.Commit()
// 	fileVersion.Id = versionID
// 	return &fileVersion, nil
// }
//
// func getLatestVersion(db *sql.DB, fileID string) (*ct.FileVersion, error) {
// 	var version ct.FileVersion
// 	err := db.QueryRow(`
//         SELECT id, timestamp, location, content, file_id
//         FROM file_versions
//         WHERE file_id = ?
//         ORDER BY timestamp DESC
//         LIMIT 1
//     `, fileID).Scan(&version.FileId, &version.Timestamp, &version.Location, &version.Content, &version.FileId)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Get changes associated with this version
// 	// changes, err := getVersionChanges(db, version.FileID)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// version.Changes = changes
//
// 	return &version, nil
// }
//
// func getVersionChanges(db *sql.DB, versionID string) ([]ct.FileChange, error) {
// 	rows, err := db.Query(`
//         SELECT type, content, position
//         FROM file_changes
//         WHERE version_id = ?
//         ORDER BY position
//     `, versionID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var changes []ct.FileChange
// 	for rows.Next() {
// 		var change ct.FileChange
// 		if err := rows.Scan(&change.Type, &change.Content, &change.Position); err != nil {
// 			return nil, err
// 		}
// 		changes = append(changes, change)
// 	}
// 	return changes, rows.Err()
// }
//
// func getInitialChanges(content string) []ct.FileChange {
// 	lines := strings.Split(content, "\n")
// 	changes := make([]ct.FileChange, len(lines))
// 	for i, line := range lines {
// 		changes[i] = ct.FileChange{
// 			Type:     "add",
// 			Content:  line,
// 			Position: i,
// 		}
// 	}
// 	return changes
// }
//
// func getChangesBetweenVersions(oldContent, newContent string) []ct.FileChange {
// 	oldLines := strings.Split(oldContent, "\n")
// 	newLines := strings.Split(newContent, "\n")
//
// 	changes := []ct.FileChange{}
// 	position := 0
//
// 	for _, line := range newLines {
// 		if position >= len(oldLines) {
// 			changes = append(changes, ct.FileChange{
// 				Type:     "add",
// 				Content:  line,
// 				Position: position,
// 			})
// 		} else if line != oldLines[position] {
// 			if containsLine(oldLines[position:], line) {
// 				changes = append(changes, ct.FileChange{
// 					Type:     "add",
// 					Content:  line,
// 					Position: position,
// 				})
// 			} else if containsLine(newLines[position:], oldLines[position]) {
// 				changes = append(changes, ct.FileChange{
// 					Type:     "remove",
// 					Content:  oldLines[position],
// 					Position: position,
// 				})
// 				position--
// 			}
// 		}
// 		position++
// 	}
//
// 	for i := position; i < len(oldLines); i++ {
// 		changes = append(changes, ct.FileChange{
// 			Type:     "remove",
// 			Content:  oldLines[i],
// 			Position: i,
// 		})
// 	}
//
// 	return changes
// }
//
// func containsLine(lines []string, line string) bool {
// 	for _, l := range lines {
// 		if l == line {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// func findFileById(db *sql.DB, id string) (*ct.File, error) {
// 	var file ct.File
// 	err := db.QueryRow("SELECT id, location, content, active FROM files WHERE id = ?", id).
// 		Scan(&file.Id, &file.Location, &file.Content, &file.Active)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("no file found with id: %s", id)
// 		}
// 		return nil, err
// 	}
// 	return &file, nil
// }
//
// func FindFileByLocation(db *sql.DB, location string) (*ct.File, error) {
// 	var file ct.File
// 	err := db.QueryRow("SELECT id, location, content, active FROM files WHERE location = ?", location).
// 		Scan(&file.Id, &file.Location, &file.Content, &file.Active)
// 	log.Println(file.Id, file.Location)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("no file found with location: %s", location)
// 		}
// 		return nil, err
// 	}
// 	return &file, nil
// }
//
// func FindFileByParam(db *sql.DB, param, param_value string) (*ct.File, error) {
// 	var file ct.File
// 	err := db.QueryRow("SELECT id, location, content, active FROM files WHERE ? = ?", param, param_value).
// 		Scan(&file.Id, &file.Location, &file.Content, &file.Active)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, fmt.Errorf("no file found with %s: %s", param, param_value)
// 		}
// 		return nil, err
// 	}
// 	return &file, nil
// }
//
// func ValidFileExtension(location string) bool {
// 	extensions := []string{".md", ".pdf"}
// 	for _, ext := range extensions {
// 		if strings.HasSuffix(location, ext) {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// func GetAllFiles(db *sql.DB) ([]ct.File, error) {
// 	rows, err := db.Query("SELECT id, location, content, active FROM files")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var files []ct.File
// 	for rows.Next() {
// 		var file ct.File
// 		if err := rows.Scan(&file.Id, &file.Location, &file.Content, &file.Active); err != nil {
// 			return nil, err
// 		}
// 		files = append(files, file)
// 	}
// 	return files, rows.Err()
// }
//
// func GetAllFileVersions(db *sql.DB, fileID string) ([]ct.FileVersion, error) {
// 	log.Println("file versions for file: ", fileID)
// 	rows, err := db.Query("SELECT  content, file_id FROM file_versions where file_id = ?", fileID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
//
// 	var files []ct.FileVersion
// 	for rows.Next() {
// 		var file ct.FileVersion
// 		if err := rows.Scan(&file.Content, &file.FileId); err != nil {
// 			return nil, err
// 		}
// 		files = append(files, file)
// 	}
// 	return files, rows.Err()
// }
//
// func deleteAllDocuments(db *sql.DB) (int64, error) {
// 	result, err := db.Exec("DELETE FROM files")
// 	if err != nil {
// 		return 0, err
// 	}
// 	return result.RowsAffected()
// }
