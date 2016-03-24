package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/kidoman/embd"
	"github.com/snikch/api/log"
	"github.com/snikch/nest/go/controller"
	"github.com/snikch/nest/go/server"
)

var (
	force = flag.Bool("force", false, "Forces the server to run regardless of hardware errors")
)

func main() {
	flag.Parse()

	// Start GPIO and defer its closing to ensure cleanup.
	err := embd.InitGPIO()
	if err != nil {
		log.WithError(err).Error("Cannot start controller. Is this running on embedded hardware?")
		os.Exit(1)
	}
	defer embd.CloseGPIO()

	// Create a new controller.
	ctrl := controller.NewController(9)
	ctrl.AddZone(controller.NewZone("Kids", 6))
	ctrl.AddZone(controller.NewZone("Downstairs", 7))
	ctrl.AddZone(controller.NewZone("Lounge", 8))

	// Run the controller
	err = ctrl.Run()
	if err != nil {
		log.WithError(err).Error("Cannot start controller. Is this running on embedded hardware?")
		os.Exit(1)
	}

	// Create a new server for controlling the controller.
	srvr := server.NewServer(ctrl)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, srvr))
}
