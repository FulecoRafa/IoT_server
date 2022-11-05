package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"net/http"
	"os"
	"time"
)

type Target struct {
	Topic string `json:"topic"`
	Addr  string `json:"addr"`
}

type Config struct {
	Broker struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"broker"`
	Targets []Target `json:"targets"`
}

func loadConfig(configFile string) (string, string, []Target) {
	var c Config

	content, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(content, &c)
	if err != nil {
		log.Fatal(err)
	}

	// Print loaded config
	fmt.Printf("Loaded config: %+v\n", c)

	return c.Broker.Host, c.Broker.Port, c.Targets
}

func main() {
	configFile := flag.String("config", "config.json", "Configuration file")
	flag.Parse()
	host, port, targets := loadConfig(*configFile)
	options := mqtt.NewClientOptions()
	connStr := "tcp://" + host + ":" + port
	options.AddBroker(connStr)
	options.SetClientID("mqtt-redirect")
	options.SetDefaultPublishHandler(handler)
	options.OnConnect = func(client mqtt.Client) {
		log.Println("Connected to MQTT broker!")
	}
	options.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Fatal("Connection lost to MQTT broker!")
	}
	options.OnReconnecting = func(client mqtt.Client, opts *mqtt.ClientOptions) {
		log.Println("Reconnecting to MQTT broker...")
	}

	client := mqtt.NewClient(options)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		log.Panic(token.Error())
	}

	for _, t := range targets {
		subscribe(token, client, t.Topic, t.Addr)
	}

	for {
		time.Sleep(60 * time.Second)
		fmt.Println("Still alive!")
	}
}

func subscribe(token mqtt.Token, client mqtt.Client, topic string, addr string) {
	token = client.Subscribe(topic, 0, topicHandler(addr))
	if token.Wait() && token.Error() != nil {
		log.Panic(token.Error())
	}
}

func handler(_client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message on topic: %s\nMessage: %s\n", msg.Topic(), msg.Payload())
}

func topicHandler(target string) func(client mqtt.Client, msg mqtt.Message) {
	httpServer := "http://" + target
	return func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message on topic: %s\nMessage: %s\n", msg.Topic(), msg.Payload())
		http.Post(httpServer, "application/json", bytes.NewReader(msg.Payload()))
	}
}
