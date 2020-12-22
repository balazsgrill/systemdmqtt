package systemdmqtt

import (
	"encoding/json"
	"log"

	"github.com/coreos/go-systemd/v22/dbus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func PublishUnitStatuses(prefix string, conn *dbus.Conn, client mqtt.Client) error {
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

			client.Publish(topic, 0, false, data)
		}
	}
	return nil
}
