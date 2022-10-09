package db

import (
	"context"
	"fmt"
	"github.com/fulecorafa/IoT_server/pkg/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB connection
func ConnectMongoDB(hostname, collectionName string) (client *mongo.Client, collection *mongo.Collection) {
	uri := fmt.Sprintf("mongodb://%s:27017", hostname)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err.Error())
	}
	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	collection = client.Database("iot").Collection(collectionName)
	return
}
