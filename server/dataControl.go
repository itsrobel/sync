package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

// TODO: turn this function into one that accepts a connection as a param
func createFile(client *mongo.Client, ctx context.Context, location string) {
	collection := client.Database("sync").Collection("server")
	file := File{location: location, active: true, contents: ""}
	result, err := collection.InsertOne(context.TODO(), file)
	if err != nil {
		log.Fatal("Error when trying to create a file: ", err)
	}
	log.Println("Inserted document with ID: ", result.InsertedID)
	// defer func() {
	// 	if err = client.Disconnect(ctx); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }()
}

// TODO: search by string location and turn active to false
func deleteFile(location string) {
}

// TODO: I need to figure out what information I can get from the file watch
func findFile(client *mongo.Client, ctx context.Context, location string) {
	collection := client.Database("sync").Collection("server")

	filter := bson.D{{"location", location}}
	opts := options.FindOne().SetProjection(bson.D{{"item", 1}, {"rating", 1}})

	var result File

	err := collection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("No documents found")
		} else {
			panic(err)
		}
	}
	res, _ := bson.MarshalExtJSON(result, false, false)
	log.Println(string(res))
}

// func getAllFiles()
