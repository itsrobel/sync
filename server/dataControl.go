package main

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
	location string `bson:"location"`
	contents string
	active   bool // this can decide whether or not to sync
}

// TODO: when a file is change it can write a change log and then
// write to the file to update
type FileChange struct {
	contentChange string
	location      string // This is the current location of the file when the change happens
	fileId        int64  // This is the id file of the file we are writing to
	active        bool
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
func connectMongo() (*mongo.Client, context.Context) {
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
func createFile(collection *mongo.Collection, location string) {
	file := File{location: location, active: true, contents: ""}
	err := ensureIndexes(collection)
	result, err := collection.InsertOne(context.Background(), file)
	if err != nil {
		log.Fatal("Error when trying to create a file: ", err)
	}
	log.Printf("Inserted document with ID: %s at %s", result.InsertedID, location)
}

// TODO: search by string location and turn active to false
// func deleteFile(location string) {
// }

// TODO: I need to figure out what information I can get from the file watch
// I need to return ID
func validFileExtension(location string) bool {
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

func getAllDocuments(collection *mongo.Collection) ([]File, error) {
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
