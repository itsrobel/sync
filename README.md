---
title: { { go-watcher } }
author: [Robel A.E. Schwarz]
date: { { 2024-09-23 } }
tags: []
sources:
  [
    https://medium.com/@abhishekranjandev/building-a-production-grade-websocket-for-notifications-with-golang-and-gin-a-detailed-guide-5b676dcfbd5a,
    https://pkg.go.dev/github.com/gorilla/websocket,
  ]
---

# Introduction

This is less of a Readme and more a design document for the development of this project

# go-watcher

# Issue Identification

I need a way to sync my notes to replace obsidian sync (go)

# Goals

I want to create an application that can be hosted on a web server as a socket server.
I would like to connect to it using my own client-selected folder(s).

The main goal for the website is to host my notes on the cloud and be able to easily download them at any time,
as well as share them using a private link system.

# Constraints

Since the web site is already being built out with gin I have to work around it

# Solution Approach

The go-watcher server will need to be assigned a folder named Alpha that it can monitor.
While the folder is being monitored, it will keep a continuous log of any changes made to
each item within the folder. When other clients connect to the go-watcher server that is
monitoring the folder Alpha, they will download the folder and then continue to monitor it
or add to the server for any future changes.

## A general list of requirements

- The client can select a folder to upload to the server and create a sync instance

  - right now the current client is a web browser instance that listens to the server
  - really the server needs to listen to the client.
  - I need to figure out how to transfer files from the client to the server

- The Server can download the folder from the client and create a "Master" copy of
- New clients that connect to the server can select which folders to then sync to
- New clients download the folder and their changes are uploaded to the server as well

## Proposed Solutions

Each of these will be Solutions on how to handle the file
differences when syncing.

1. Solution 1:
   When a file is changed for the server, re-download that file
   for each of the clients

   - Pros:
     is properly implemented in the easiest way
   - Cons:
     at scale will suck balls and requires a lot of network usage for each iteration

Even with the file transfer system, I need to atleast keep track of the movement of the files/
what their names are so I do not re download the entire file system each time a file change is made

### File transfer over Gin

In Go, using the Gin web framework, you can send a file to the client by using the `Context.File()` method, which sends the specified file as an HTTP response. Below is an example of how you might set up a Gin HTTP server with an endpoint to send a file:

```go
package main
import (
	"github.com/gin-gonic/gin"
	"net/http"
)
func main() {
	// Initialize the Gin router
	router := gin.Default()
	// Define a route that sends a file when accessed
	router.GET("/file", func(c *gin.Context) {
		// Specify the file path
		filePath := "path/to/your/file.txt"
		// Check if the file exists and if it is not a directory before sending
		if _, err := os.Stat(filePath); err == nil {
			c.File(filePath)
		} else {
			// If there's an error (like file not found), return an HTTP 404 status
			c.AbortWithStatus(http.StatusNotFound)
		}
	})
	// Run the server on port 8080
	router.Run(":8080")
}
```

In this example, when a client sends a GET request to `http://localhost:8080/file`, the server will respond by sending the file located at `path/to/your/file.txt`. If the file does not exist, the server will respond with a 404 Not Found status.
Make sure to replace `path/to/your/file.txt` with the actual path to the file you want to serve.
To run this code, you need to have the Gin package installed. If you haven't already installed Gin, you can get it using the following command:

2. Solution 2:
   When a file is changed for the server, for each client send out the difference

   - Pros
     is probably the best and maybe the most "fun" to implement
   - Cons
     requirements are much higher

   What would I need for this solution?

   I need a way for each of the clients to have a master state of machine
   or each of the files and their current "version"?

   How do I file version?

### Versioning in Go

In Go, handling file versions typically involves implementing a system to manage different versions of a file, often by saving each version with a unique identifier, such as a timestamp or version number. Here's a basic approach to creating a file versioning system in Go:

1. **Define a Naming Convention**: Decide on a naming convention for your file versions. For example, you might append a timestamp or an incremental version number to the file's name.
2. **Save New Versions**: When saving a new version of a file, use the naming convention to create a new file rather than overwriting the existing one.
3. **List Versions**: Implement a function to list all versions of a file.
4. **Retrieve a Specific Version**: Implement a function to retrieve a specific version of a file based on its unique identifier.
   Here's an example of how you might implement a simple file versioning system in Go:

```go
package main
import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)
// saveVersion saves a new version of the file with the current timestamp.
func saveVersion(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(filePath)
	name := base[0 : len(base)-len(ext)]
	version := time.Now().Format("20060102T150405") // YYYYMMDDTHHMMSS format
	newFileName := fmt.Sprintf("%s_%s%s", name, version, ext)
	newFilePath := filepath.Join(dir, newFileName)
	return ioutil.WriteFile(newFilePath, data, 0644)
}
// listVersions lists all versions of a file in the directory.
func listVersions(filePath string) ([]string, error) {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(filePath)
	name := base[0 : len(base)-len(ext)]
	pattern := fmt.Sprintf("%s_*%s", name, ext)
	files, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}
	return files, nil
}
func main() {
	// Example usage
	filePath := "example.txt"
	data := []byte("This is some file content.")
	// Save a new version of the file
	if err := saveVersion(filePath, data); err != nil {
		fmt.Println("Error saving version:", err)
		return
	}
	// List all versions of the file
	versions, err := listVersions(filePath)
	if err != nil {
		fmt.Println("Error listing versions:", err)
		return
	}
	fmt.Println("Versions of the file:")
	for _, v := range versions {
		fmt.Println(v)
	}
}
```

In this example, the `saveVersion` function saves a new version of the file with a timestamp appended to the file name. The `listVersions` function lists all versions of the file based on the naming pattern.
This is a basic example, and a real-world application might require more features, such as version deletion, rollback, metadata storage, and more sophisticated error handling. You might also want to consider using a database to track file versions and metadata if your application requires complex versioning capabilities.

### File Ledger of changes

To keep a ledger of changes made to a file in Go, you can implement a system that records each change along with metadata such as timestamps, user information, and a description of the change. This ledger can be stored in a separate file, a database, or any other persistent storage system.
Here's a simple example of how you might implement a file change ledger using a JSON file to store the change records:

```go
package main
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)
// ChangeRecord represents a single change made to the file.
type ChangeRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	User        string    `json:"user"`
	Description string    `json:"description"`
}
// Ledger represents a list of change records.
type Ledger struct {
	Records []ChangeRecord `json:"records"`
}
// AddRecord adds a new change record to the ledger.
func (l *Ledger) AddRecord(user, description string) {
	record := ChangeRecord{
		Timestamp:   time.Now(),
		User:        user,
		Description: description,
	}
	l.Records = append(l.Records, record)
}
// Save writes the ledger to a JSON file.
func (l *Ledger) Save(filePath string) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0644)
}
// LoadLedger loads the ledger from a JSON file.
func LoadLedger(filePath string) (*Ledger, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var ledger Ledger
	err = json.Unmarshal(data, &ledger)
	if err != nil {
		return nil, err
	}
	return &ledger, nil
}
func main() {
	ledgerFile := "ledger.json"
	// Load the existing ledger or create a new one if it doesn't exist
	ledger, err := LoadLedger(ledgerFile)
	if err != nil {
		if os.IsNotExist(err) {
			ledger = &Ledger{}
		} else {
			fmt.Println("Error loading ledger:", err)
			return
		}
	}
	// Add a new record to the ledger
	ledger.AddRecord("username", "Made some changes to the file")
	// Save the updated ledger
	if err := ledger.Save(ledgerFile); err != nil {
		fmt.Println("Error saving ledger:", err)
		return
	}
	fmt.Println("Ledger updated successfully.")
}
```

In this example, we define two structs: `ChangeRecord` to represent individual changes and `Ledger` to represent the entire ledger. The `Ledger` struct has methods to add a new record and save the ledger to a JSON file. We also have a `LoadLedger` function to load the ledger from a file.
The `main` function demonstrates how to load an existing ledger, add a new change record, and save the updated ledger back to the file.
This is a basic implementation, and depending on your requirements, you might want to add more features, such as error handling for concurrent access, a more sophisticated storage system, or the ability to revert changes.

# Implementation Plan

The first thing that Is required effectively for both solutions is file versioning and send out what changes where made and to which files they were made to.

- There can be two ways to version a directory versioning and a individual file versioning.

  - the directory versioning would update very 6hrs -> 12hrs and would be a master snap shot of all the file versions
    uploaded to the server

    - this would handle the problem of what happens when a client has not connected to the server for a few days
      the client would iterate through the changes made in the ledger of the client and server to point to the
      file changes.

    - the directory versioning would also be able to handle file movement/renaming/and deletion

## Action Items

1. Action Item 1:
2. Action Item 2:
3. Action Item 3:

# Risks and Challenges

- Identify any potential risks, challenges, or obstacles that may arise.

# Lessons learned/Lessons to Learn

- Reflect on lessons learned during the implementation process.

# Conclusion

- Summarize the key points and reiterate the importance of the solution-oriented approach.
