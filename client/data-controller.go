package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

func databaseConnection() {
	db, err := sql.Open("sqlite3", "./sync.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}

// 	// Open the SQLite database file
// 	db, err := sql.Open("sqlite3", "./sync.db")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()
// 	// Create a table
// 	sqlStmt := `
//     CREATE TABLE IF NOT EXISTS userinfo (
//         id INTEGER PRIMARY KEY AUTOINCREMENT,
//         username TEXT NOT NULL,
//         created_at DATETIME
//     );
//     `
// 	_, err = db.Exec(sqlStmt)
// 	if err != nil {
// 		log.Printf("%q: %s\n", err, sqlStmt)
// 		return
// 	}
// 	// Insert a new user
// 	_, err = db.Exec("INSERT INTO userinfo(username, created_at) VALUES(?, ?)", "johndoe", "2023-04-01")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// Query the database
// 	rows, err := db.Query("SELECT id, username, created_at FROM userinfo")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var id int
// 		var username string
// 		var createdAt string
// 		err = rows.Scan(&id, &username, &createdAt)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Println(id, username, createdAt)
// 	}
// 	err = rows.Err()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
