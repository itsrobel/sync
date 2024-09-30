# File transfer over Gin

## Server to client transfer

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
To run this code, you need to have the Gin package installed.

## Client to server transfer

Since the changes will be made on the client instead of the server
handling the transfer to server is more "Important"

To send a file from a Go client to a Go server using the Gin framework, you'll need to set up a Gin server to handle file uploads and a client that can send a POST request with a multipart/form-data content type. Below are examples of how to implement both the server and the client.

### Gin HTTP Server

```go
package main
import (
	"github.com/gin-gonic/gin"
	"net/http"
)
func main() {
	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("/upload", func(c *gin.Context) {
		// Single file
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "Get form err: %s", err.Error())
			return
		}
		filename := file.Filename
		// Save the file to the server's local storage
		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.String(http.StatusInternalServerError, "Save file err: %s", err.Error())
			return
		}
		c.String(http.StatusOK, "File %s uploaded successfully.", filename)
	})
	router.Run(":8080")
}
```

### Gin HTTP Client

```go
package main
import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)
func main() {
	filePath := "example.txt"
	fileName := filepath.Base(filePath)
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// Create a buffer to write our multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Create a form field writer for the file
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		panic(err)
	}
	// Copy the file into the form field writer
	_, err = io.Copy(part, file)
	if err != nil {
		panic(err)
	}
	// Close the writer to finalize the multipart form
	writer.Close()
	// Create a new request with the form data
	request, err := http.NewRequest("POST", "http://localhost:8080/upload", body)
	if err != nil {
		panic(err)
	}
	// Set the content type header, which contains the boundary string for the form
	request.Header.Set("Content-Type", writer.FormDataContentType())
	// Perform the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	// Check the response
	if response.StatusCode == http.StatusOK {
		fmt.Println("File uploaded successfully")
	} else {
		fmt.Printf("Error uploading file: %s\n", response.Status)
	}
}
```

In the server code, we define a POST endpoint `/upload` that handles file uploads. The client code creates a multipart form and sends the file to the server's `/upload` endpoint.
To test this, first run the server code, and then run the client code. Make sure to replace `"example.txt"` with the path to the actual file you want to send from the client to the server.
Remember to handle errors and edge cases appropriately in a production environment, such as checking for file size limits, handling file overwrites, and securing the file upload endpoint.
