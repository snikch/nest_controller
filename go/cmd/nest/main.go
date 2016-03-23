package main

import (
	"log"
	"net/http"
	"os"

	"github.com/snikch/api/render"
	"github.com/snikch/api/vc"
	"github.com/snikch/nest/go/controller"
	"github.com/snikch/nest/go/server"
)

func init() {
	vc.DefaultRenderer = render.JSONRenderer{}
}

func main() {
	ctrl := controller.NewController(9)
	ctrl.AddZone(controller.NewZone("Kids", 6))
	ctrl.AddZone(controller.NewZone("Downstairs", 7))
	ctrl.AddZone(controller.NewZone("Lounge", 8))
	srvr := server.NewServer(ctrl)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, srvr))
}
