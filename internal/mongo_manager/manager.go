package manager

import (
	"context"
	"fmt"
	"log"

	ft "github.com/itsrobel/sync/internal/services/filetransfer"
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

// NOTE: When I move the structure to build file content via
// func CreateFileInital(collection *mongo.Collection, file string) (string, error) {
// 	file := ct.File{Location: location, Active: true, Content: ""}
// 	ensureIndexes(collection)
// 	result, err := collection.InsertOne(context.Background(), file)
// 	if err != nil {
// 		log.Fatal("Error when trying to create a file: ", err)
// 		return "", nil
// 	}
// 	log.Printf("Inserted document with ID: %s at %s", result.InsertedID, location)
// 	return result.InsertedID.(string), nil
// }

func CreateFileVersion(collection *mongo.Collection, file *ft.FileVersionData) {
	// TODO: check for uuid duplicates later from multiple clients creating files
	fileVersion := ct.FileVersion{Timestamp: file.Timestamp.AsTime(), Location: file.Location, Content: string(file.Content), FileId: file.FileId}
	result, err := collection.InsertOne(context.Background(), fileVersion)
	if err != nil {
		log.Fatal("Error when trying to create a file version: ", err)
	}
	log.Printf("Inserted document with ID: %s at %s", result.InsertedID, file.Location)
}

func FindFileByLocation(collection *mongo.Collection, location string) (*ct.File, error) {
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

func FindFileById(collection *mongo.Collection, fileID string) (*ct.File, error) {
	filter := bson.M{"ID": fileID}
	var result ct.File
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no document found with ID: %s", fileID)
		}
		return nil, fmt.Errorf("error finding document: %v", err)
	}
	return &result, nil
}

func GetAllDocuments(collection *mongo.Collection) ([]ct.FileVersion, error) {
	// Create a context (you might want to use a timeout context in a real application)
	ctx := context.Background()

	// Find all documents
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Create a slice to store the documents
	var documents []ct.FileVersion

	// Iterate through the cursor and decode each document
	for cursor.Next(ctx) {
		var doc ct.FileVersion
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	// Check if the cursor encountered any errors while iterating
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	log.Println(documents)
	return documents, nil
}

func DeleteAllDocuments(collection *mongo.Collection) (int64, error) {
	// Create a context (you might want to use a timeout context in a real application)
	ctx := context.Background()

	// Delete all documents
	result, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("failed to delete documents: %v", err)
	}

	return result.DeletedCount, nil
}
