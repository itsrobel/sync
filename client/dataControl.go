package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NOTE: ID's are created by default
// time?
type File struct {
	contents string
	location string
	active   bool // this can decide whether or not to sync
}

// TODO: when a file is change it can write a change log and then
// write to the file to update
type FileChange struct {
	location string // This is the current location of the file when the change happens
	fileId   int64  // This is the id file of the file we are writing to
	active   bool
}

// make this return the conenction
func connectMongo() (*mongo.Client, context.Context) {
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
	return client, ctx
}

func createFile(location string, active bool, contents string) {
	client, ctx := connectMongo()
	collection := client.Database("sync").Collection("client")
	file := File{location: location, active: active, contents: contents}

	result, err := collection.InsertOne(ctx, file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted document with ID: %v\n", result.InsertedID)

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()
}

// TODO: I need to figure out what information I can get from the file watch
func updateFile(location string) {
}
