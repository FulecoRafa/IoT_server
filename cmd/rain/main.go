package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fulecorafa/IoT_server/pkg/discord"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fulecorafa/IoT_server/pkg/db"
	"github.com/fulecorafa/IoT_server/pkg/log"
)

type RainEntry struct {
	HumidityLevel float64   `json:"humidityLevel"`
	EntryDate     time.Time `json:"entryDate"`
}

// MongoDB connection
var client *mongo.Client
var collection *mongo.Collection

// Global state
var wasRaining = false

func main() {
	/* Initializing logger */
	log.Init("rain")

	/* Starting microservice */
	log.Info("Starting rain microservice...")
	defer log.Info("Rain microservice stopped!")

	/* Connecting to MongoDB */
	log.Info("Connecting to MongoDB...")
	mongoHostname := os.Getenv("MONGO_HOSTNAME")
	if mongoHostname == "" {
		log.Fatal("MONGO_HOSTNAME environment variable not set")
	}
	client, collection = db.ConnectMongoDB(mongoHostname, "rain")
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
	http.HandleFunc("/", rainHandler)

	log.Info("HTTP server up at http://localhost:6969")
	err := http.ListenAndServe(":6969", nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func rainHandler(w http.ResponseWriter, r *http.Request) {
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
			return
		}
		return
	case http.MethodPost:
		var entry RainEntry
		body, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(body, &entry)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("%v\n", entry)
		err = postData(entry)
		if err != nil {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		isRaining := entry.HumidityLevel > 0.5
		if !isRaining && wasRaining {
			wasRaining = false
		}
		if isRaining && !wasRaining {
			const message = "Hey there! My sensors are telling me that it's raining outside. I suggest you take a look at the windows!"

			if err := discord.SendMessage(message); err != nil {
				log.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(strconv.FormatBool(isRaining)))
		if err != nil {
			log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func getData(page, limit int) (rainEntries []RainEntry, err error) {
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

	err = cursor.All(context.TODO(), &rainEntries)
	if err != nil {
		return nil, err
	}
	return rainEntries, nil
}

func postData(entry RainEntry) error {
	_, err := collection.InsertOne(context.TODO(), entry)
	if err != nil {
		return err
	}
	return nil
}
