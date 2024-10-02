To create a connection between a local SQLite database and Go, you can use the `go-sqlite3` package, which is a Go SQLite driver for the `database/sql` package. Here's how to set it up:

1. Install the `go-sqlite3` package:
   First, you need to install the `go-sqlite3` package by running the following command:

```bash
go get github.com/mattn/go-sqlite3
```

2. Import the necessary packages:
   In your Go file, import the `database/sql` package along with the SQLite driver:

```go
package main
import (
    "database/sql"
    "log"
    _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)
```

Note that the underscore `_` before the import is used to import the package solely for its side-effects (i.e., registering the driver). 3. Open a connection to the SQLite database:
Now, you can open a connection to your SQLite database file:

```go
func main() {
    // Open the SQLite database file
    db, err := sql.Open("sqlite3", "./mydatabase.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    // Perform database operations here
}
```

Replace `./mydatabase.db` with the path to your SQLite database file. If the file does not exist, `go-sqlite3` will create it for you. 4. Use the database connection:
Once connected, you can use the `db` object to interact with your SQLite database. For example, to create a table:

```sql
CREATE TABLE IF NOT EXISTS
```

```go
func main() {
    // Open the SQLite database file
    db, err := sql.Open("sqlite3", "./mydatabase.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    // Create a table
    sqlStmt := `
    CREATE TABLE IF NOT EXISTS userinfo (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        created_at DATETIME
    );
    `
    _, err = db.Exec(sqlStmt)
    if err != nil {
        log.Printf("%q: %s\n", err, sqlStmt)
        return
    }
    // Insert a new user
    _, err = db.Exec("INSERT INTO userinfo(username, created_at) VALUES(?, ?)", "johndoe", "2023-04-01")
    if err != nil {
        log.Fatal(err)
    }
    // Query the database
    rows, err := db.Query("SELECT id, username, created_at FROM userinfo")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    for rows.Next() {
        var id int
        var username string
        var createdAt string
        err = rows.Scan(&id, &username, &createdAt)
        if err != nil {
            log.Fatal(err)
        }
        log.Println(id, username, createdAt)
    }
    err = rows.Err()
    if err != nil {
        log.Fatal(err)
    }
}
```

This example demonstrates how to create a table, insert a row, and query the table for all rows. Remember to handle errors and close the database connection when you're done with it.
That's it! You've now established a connection to a local SQLite database and performed some basic operations using Go.

```sql
CREATE TABLE files_changed (
  id INTEGER PRIMARY KEY,
  file_name TEXT,
  version TEXT
);

CREATE TABLE files_deleted (
  id INTEGER PRIMARY KEY,
  file_name TEXT
);

CREATE TABLE files (
  id INTEGER PRIMARY KEY,
  uuid TEXT,
  file_name TEXT,
  contents TEXT,
  location TEXT,
  changes TEXT
);

CREATE TABLE file_change_timestamp (
  id INTEGER PRIMARY KEY,
  timestamp TEXT,
  device TEXT,
  file_diff TEXT
);
```

## Data Structures

To keep track of these versioning, I need the project to be hooked up to a database.
In terms of data structures and what the database might look like.

How do I handle non markdown files?

```json file
{
  "id": "<the uuid of the file>",
  "file-name": "<name of the file given from the user",
  "location": "<location of the file in the directory",
  "contents": "<contents of the file>",
  "status": "<active/deleted>"
}
```

Any changes from the user from the client would be made to the
would point to the id of the file by searching with the location string

If the file is renamed on the server and tries to be synced, if the
file id matches then the location of the latest timestamp is taken

The ID system can be used for conflict resolution

```json file-change-{timestamp}
{
  "device": "<device of where the change was made from>",
  "id": "<the id of the file>",
  "location": "<location of the file",
  "file-diff": "<content changes of the file>",
  "status": "<deleted/moved/edited>"
}
```

If the file status is deleted for more than 30 days or something I can remove the contents and the ID from the sql table

nvm, I still hate sql, using mongo, i'll make a sql project next semster
when I take databases. Or I will just migrate this project later

1. **Install the MongoDB Go Driver**:
   First, you need to install the MongoDB Go driver by running the following command:
   ```bash
   go get go.mongodb.org/mongo-driver/mongo
   ```
2. **Import the necessary packages**:
   In your Go file, import the MongoDB driver and other necessary packages:
   ```go
   package main
   import (
       "context"
       "log"
       "time"
       "go.mongodb.org/mongo-driver/bson"
       "go.mongodb.org/mongo-driver/mongo"
       "go.mongodb.org/mongo-driver/mongo/options"
   )
   ```
3. **Create a MongoDB client and connect**:
   Now, you can create a MongoDB client, configure the connection options, and connect to the MongoDB instance:
   ```go
   func main() {
       // Set client options
       clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
       // Connect to MongoDB
       client, err := mongo.Connect(context.TODO(), clientOptions)
       if err != nil {
           log.Fatal(err)
       }
       // Check the connection
       ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
       defer cancel()
       err = client.Ping(ctx, nil)
       if err != nil {
           log.Fatal(err)
       }
       fmt.Println("Connected to MongoDB!")
       // Don't forget to close the connection when you're done
       defer func() {
           if err = client.Disconnect(ctx); err != nil {
               log.Fatal(err)
           }
       }()
   }
   ```
   Replace `"mongodb://localhost:27017"` with the URI of your MongoDB instance. If you're connecting to a MongoDB Atlas cluster, you'll need to use the connection string provided in your Atlas dashboard, which will include your username, password, and cluster details.
4. **Use the client**:
   Once connected, you can use the `client` object to interact with your MongoDB databases and collections. For example, to insert a document into a collection:

   ```go
   collection := client.Database("testdb").Collection("testcollection")
   doc := bson.D{{"name", "John Doe"}, {"age", 30}}
   result, err := collection.InsertOne(ctx, doc)
   if err != nil {
   	log.Fatal(err)
   }
   fmt.Printf("Inserted document with ID: %v\n", result.InsertedID)


   ```

   Remember to handle errors and context timeouts appropriately in a production environment. The above example uses `context.TODO()` and `context.WithTimeout()` to create contexts with and without timeouts, respectively. You should use a proper context with a timeout for real applications.

https://www.mongodb.com/docs/drivers/go/current/fundamentals/bson/
