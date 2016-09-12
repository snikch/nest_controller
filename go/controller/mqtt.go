package controller

import (
	"fmt"
	"strings"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/snikch/api/log"
)

type MQTTEventHandler struct {
	mqtt.Client
	EntityName string
}

func NewMQTTEventHandler(name string, client mqtt.Client) *MQTTEventHandler {
	return &MQTTEventHandler{
		Client:     client,
		EntityName: strings.ToLower(name),
	}
}

func (handler *MQTTEventHandler) Push(event Event) error {
	message := fmt.Sprintf("sensor/%s/%s", handler.EntityName, event.Entity)
	log.WithField("message", message).WithField("value", event.Value).Info("Publishing message")
	token := handler.Publish(message, 0, true, event.Value)
	token.Wait()
	return token.Error()
}

func (handler *MQTTEventHandler) Name() string {
	return "mqtt"
}
