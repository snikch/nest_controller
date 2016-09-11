package main

import (
	"net/http"
	"os"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
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

	// Run the controller
	go func() {
		log.Info("Starting controller")
		err = ctrl.Run()
		if err != nil {
			log.WithError(err).Error("Cannot start controller. Is this running on embedded hardware?")
			os.Exit(1)
		}
	}()

	// Create a new server for controlling the controller.
	srvr := server.NewServer(ctrl)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go func() {
		log.WithField("port", port).Info("Starting Server")
		log.Fatal(http.ListenAndServe(":"+port, srvr))
	}()
	lifecycle.WaitForShutdown()
}
