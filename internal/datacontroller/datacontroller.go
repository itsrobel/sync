package datacontroller

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NOTE: ID's are created by default
// time?
type File struct {
	Location string `bson:"location"`
	Contents string `bson:"contents"`
	Active   bool   `bson:"active"` // this can decide whether or not to sync
}

// Every Hour if changes have been made create a new Version
// Shouldn't the file just point to the latest version?
type FileVersion struct {
	Timestamp time.Time `bson:"time_stamp"` // Time when this version was created
	Location  string    `bson:"location"`   // File location
	Contents  string    `bson:"contents"`   // Full contents of the file at this version
	FileID    string    `bson:"file_id"`    // Unique ID for the file
}

// TODO: when a file is change it can write a change log and then
// write to the file to update
type FileChange struct {
	Timestamp     time.Time `bson:"time_stamp"`     // Time when this version was created
	ContentChange string    `bson:"content_change"` // Full contents of the file at this version
	Location      string    `bson:"location"`       // File location
	VersionID     string    `bson:"version_id"`     // Unique ID for the file
}

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
func ConnectMongo() (*mongo.Client, context.Context) {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	// The connection context only lasts as long as specified in the timemout, since
	// We are running these commands not on a time frame we should be able to use contex.TODO although that is likely
	// not best practice
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB!")
	return client, ctx
}

// func updateFile(client *mongo.Client, location string, content string) {
// 	collection := client.Database("sync").Collection("server")
// 	fileUpdate := FileChange{contentChange: content, location: location}
//
// 	// find the fileID of the file with the specified location
// }

// TODO: turn this function into one that accepts a connection as a param
// TODO: I need to restrict file types
func CreateFile(collection *mongo.Collection, location string) string {
	file := File{Location: location, Active: true, Contents: ""}
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

	fileVersion := FileVersion{Timestamp: time.Now(), Location: location, Contents: contents, FileID: fileID}
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

func findFile(collection *mongo.Collection, location string) (*File, error) {
	filter := bson.M{"location": location}
	var result File
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no document found with location: %s", location)
		}
		return nil, fmt.Errorf("error finding document: %v", err)
	}

	return &result, nil
}

func GetAllDocuments(collection *mongo.Collection) ([]File, error) {
	// Create a context (you might want to use a timeout context in a real application)
	ctx := context.Background()

	// Find all documents
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Create a slice to store the documents
	var documents []File

	// Iterate through the cursor and decode each document
	for cursor.Next(ctx) {
		var doc File
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
