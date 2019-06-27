package main

import (
	"net/http"
	"os"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/snikch/api/config"
	"github.com/snikch/api/lifecycle"
	"github.com/snikch/api/log"
	"github.com/snikch/nest/go/controller"
	"github.com/snikch/nest/go/server"
	"github.com/stianeikeland/go-rpio"
)

func main() {

	// Start GPIO and defer its closing to ensure cleanup.
	err := embd.InitGPIO()
	if err != nil {
		log.WithError(err).Error("Cannot start embd GPIO. Is this running on embedded hardware?")
		os.Exit(1)
	}

	// Because embd doesn't support pull up or down, we do that with the rpio lib.
	err = rpio.Open()
	if err != nil {
		log.WithError(err).Error("Cannot start rpio GPIO. Is this running on embedded hardware?")
		os.Exit(1)
	}

	// Ensure that shutdown means we close GPIO access.
	lifecycle.RegisterShutdownCallback("close embd gpio", embd.CloseGPIO)
	lifecycle.RegisterShutdownCallback("close rpio gpio", rpio.Close)

	// Create the house zones.
	// Lounge 18
	// Kids 23
	// Downstairs 24
	kids := controller.NewZone("Kids", 23)
	downstairs := controller.NewZone("Downstairs", 24)

	kids.DamperOnPin = 17
	kids.DamperOffPin = 2
	downstairs.DamperOnPin = 27
	downstairs.DamperOffPin = 3

	// Create a new controller, and add zones.
	ctrl := controller.NewController(12)
	log.Info("Adding Downstairs zone")
	ctrl.AddZone(downstairs)
	log.Info("Adding Kids zone")
	ctrl.AddZone(kids)
	log.Info("Adding Lounge zone")
	ctrl.AddZone(controller.NewZone("Lounge", 18))

	// Now we generate an mqtt event handler
	broker := config.String("MQTT_BROKER", "tcp://192.168.0.119:1883")
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID("fireplace")
	opts.SetKeepAlive(20 * time.Second)
	// opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.WithError(err).WithField("address", broker).Fatal("Could not connect to mqtt broker")
	}
	lifecycle.RegisterShutdownCallback("mqtt disconnect", func() error {
		c.Disconnect(250)
		return nil
	})
	ctrl.AddEventHandler(controller.NewMQTTEventHandler("fireplace", c))

	// Run the controller
	go func() {
		log.Info("Starting controller")
		err = ctrl.Run()
		if err != nil {
			log.WithError(err).Fatal("Cannot start controller. Is this running on embedded hardware?")
		}
	}()

	// Create a new server for controlling the controller.
	srvr := server.NewServer(ctrl)
	port := config.String("PORT", "8080")
	go func() {
		log.WithField("port", port).Info("Starting Server")
		log.Fatal(http.ListenAndServe(":"+port, srvr))
	}()
	lifecycle.WaitForShutdown()
}
