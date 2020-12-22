package main

import (
	"log"
	"time"

	"github.com/balazsgrill/systemdmqtt"
	"github.com/coreos/go-systemd/v22/dbus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	conn, err := dbus.NewSystemConnection()
	if err != nil {
		log.Fatal(err)
	}
	opts := mqtt.NewClientOptions().AddBroker("192.168.0.1").SetAutoReconnect(true)
	client := mqtt.NewClient(opts)
	client.Connect().Wait()

	ticker := time.NewTicker(100 * time.Millisecond)
	for range ticker.C {
		err := systemdmqtt.PublishUnitStatuses("systemd/test", conn, client)
		if err != nil {
			log.Fatal(err)
		}
	}
}
