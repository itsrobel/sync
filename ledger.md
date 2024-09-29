# File Ledger of changes

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
