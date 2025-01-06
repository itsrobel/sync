package manager

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	ct "github.com/itsrobel/sync/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ensureIndexes(collection *mongo.Collection) error {
	_, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "location", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

// make this return the conenction

// func updateFile(client *mongo.Client, location string, content string) {
// 	collection := client.Database("sync").Collection("server")
// 	fileUpdate := FileChange{contentChange: content, location: location}
//
// 	// find the fileID of the file with the specified location
// }

func CreateFile(collection *mongo.Collection, location string) string {
	file := ct.File{Location: location, Active: true, Contents: ""}
	err := ensureIndexes(collection)
	result, err := collection.InsertOne(context.Background(), file)
	if err != nil {
		log.Fatal("Error when trying to create a file: ", err)
	}
	log.Printf("Inserted document with ID: %s at %s", result.InsertedID, location)
	return result.InsertedID.(string)
}

func CreateFileVersion(collection *mongo.Collection, fileID string) {
	file, _ := findFile(collection, fileID)
	location := file.Location
	contents := file.Contents

	fileVersion := ct.FileVersion{Timestamp: time.Now(), Location: location, Contents: contents, FileID: fileID}
	result, err := collection.InsertOne(context.Background(), fileVersion)
	if err != nil {
		log.Fatal("Error when trying to create a file version: ", err)
	}
	log.Printf("Inserted document with ID: %s at %s", result.InsertedID, location)
}

// TODO: search by string location and turn active to false
// func deleteFile(location string) {
// }

// TODO: I need to figure out what information I can get from the file watch
// I need to return ID
func ValidFileExtension(location string) bool {
	extensions := []string{".md", ".pdf"}
	for _, ext := range extensions {
		if strings.HasSuffix(location, ext) {
			return true
		}
	}
	return false
}

func findFile(collection *mongo.Collection, location string) (*ct.File, error) {
	filter := bson.M{"location": location}
	var result ct.File
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no document found with location: %s", location)
		}
		return nil, fmt.Errorf("error finding document: %v", err)
	}

	return &result, nil
}

func GetAllDocuments(collection *mongo.Collection) ([]ct.File, error) {
	// Create a context (you might want to use a timeout context in a real application)
	ctx := context.Background()

	// Find all documents
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Create a slice to store the documents
	var documents []ct.File

	// Iterate through the cursor and decode each document
	for cursor.Next(ctx) {
		var doc ct.File
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	// Check if the cursor encountered any errors while iterating
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}

func deleteAllDocuments(collection *mongo.Collection) (int64, error) {
	// Create a context (you might want to use a timeout context in a real application)
	ctx := context.Background()

	// Delete all documents
	result, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to delete documents: %v", err)
	}

	return result.DeletedCount, nil
}

// func getAllFiles()
