package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	sse "github.com/r3labs/sse/v2"
)

func main() {
	broker := flag.String("b", "tcp://127.0.0.1:1883", "MQTT broker URL e.g. tcp://127.0.0.1:1883")
	flag.Parse()

	log.Println("Starting")
	eventserver := sse.New()
	eventserver.CreateStream("messages")

	opts := mqtt.NewClientOptions().AddBroker(*broker).SetAutoReconnect(true).SetOnConnectHandler(
		func(client mqtt.Client) {
			client.Subscribe("systemd/#", 0, func(client mqtt.Client, msg mqtt.Message) {
				var dat map[string]interface{}
				err := json.Unmarshal(msg.Payload(), &dat)
				if err != nil {
					log.Println(err)
					return
				}
				splittopic := strings.Split(msg.Topic(), "/")
				dat["host"] = splittopic[1]
				bs, _ := json.Marshal(dat)
				eventserver.Publish("messages", &sse.Event{Data: bs})
			})
		})
	client := mqtt.NewClient(opts)
	t := client.Connect()
	if t.Error() != nil {
		log.Fatal(t.Error())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/events", eventserver.HTTPHandler)
	mux.Handle("/", http.FileServer(http.Dir("./")))

	go startMonitoringUnits(client)
	log.Println(http.ListenAndServe(":8080", mux))
}

func startMonitoringUnits(client mqtt.Client) {
	hname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	conn, err := dbus.NewSystemConnection()
	if err != nil {
		log.Fatal(err)
	}
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		if !client.IsConnected() {
			client.Connect()
		} else {
			err := publishUnitStatuses("systemd/"+hname, conn, client)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func publishUnitStatuses(prefix string, conn *dbus.Conn, client mqtt.Client) error {
	statuses, err := conn.ListUnits()
	if err != nil {
		return err
	}
	for _, s := range statuses {
		data, err := json.Marshal(s)
		if err != nil {
			log.Println(err)
		} else {
			topic := prefix + "/" + s.Name
			t := client.Publish(topic, 0, false, data)
			if t.Error() != nil {
				log.Println(t.Error())
			}
		}
	}
	return nil
}
