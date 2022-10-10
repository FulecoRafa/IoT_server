package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fulecorafa/IoT_server/pkg/db"
	"github.com/fulecorafa/IoT_server/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type DoorEntry struct {
	Distance  float64   `json:"distance"`
	EntryDate time.Time `json:"entryDate"`
}

// MongoDB connection
var client *mongo.Client
var collection *mongo.Collection

func main() {
	/* Initializing logger */
	log.Init("door")

	/* Starting microservice */
	log.Info("Starting door microservice...")
	defer log.Info("Door microservice stopped!")

	/* Connecting to MongoDB */
	log.Info("Connecting to MongoDB...")
	mongoHostname := os.Getenv("MONGO_HOSTNAME")
	if mongoHostname == "" {
		log.Fatal("MONGO_HOSTNAME environment variable not set")
	}
	client, collection = db.ConnectMongoDB(mongoHostname, "door")
	defer log.Info("MongoDB connection closed successfully!")
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Error(err.Error())
		}
	}(client, nil)
	defer log.Info("Trying to free MongoDB connection...")
	log.Info("Connected to MongoDB!")

	/* Starting HTTP server */
	log.Info("Starting HTTP server...")
	http.HandleFunc("/", doorHandler)

	log.Info("HTTP server up at http://localhost:6970")
	err := http.ListenAndServe(":6970", nil)
	if err != nil {
		log.Fatal(err.Error())
	}

}

func doorHandler(w http.ResponseWriter, r *http.Request) {
	log.LogRequest(r)
	query := r.URL.Query()
	switch r.Method {
	case http.MethodGet:
		page, err := strconv.Atoi(query.Get("page"))
		if err != nil {
			page = 1
		}
		limit, err := strconv.Atoi(query.Get("limit"))
		if err != nil {
			limit = 10
		}

		data, err := getData(page, limit)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		parsedData, err := json.Marshal(data)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(parsedData)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	case http.MethodPost:
		var entry DoorEntry
		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &entry)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println(entry)
		err = postData(entry)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		isOpen := entry.Distance > 10
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(strconv.FormatBool(isOpen)))
		if err != nil {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	default:
		http.Error(w, `{"error": "Method not allowed"`, http.StatusMethodNotAllowed)
	}
}

func getData(page int, limit int) (doorEntries []DoorEntry, err error) {
	cursor, err := collection.Find(context.TODO(), bson.D{}, options.Find().SetSkip(int64((page-1)*limit)).SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			log.Error(err.Error())
		}
	}(cursor, context.TODO())
	err = cursor.All(context.TODO(), &doorEntries)
	if err != nil {
		return nil, err
	}
	return doorEntries, nil
}

func postData(entry DoorEntry) error {
	_, err := collection.InsertOne(context.TODO(), entry)
	if err != nil {
		return err
	}
	return nil
}
