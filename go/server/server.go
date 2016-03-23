package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/snikch/api/ctx"
	"github.com/snikch/api/vc"
	"github.com/snikch/nest/go/controller"
)

func NewServer(ctrl *controller.Controller) http.Handler {
	router := httprouter.New()
	processor := vc.ActionProcessor{}
	server := Server{ctrl}
	router.GET("/status", processor.HandleActionFunc("status", server.HandleStatus))
	return router
}

type Server struct {
	Controller *controller.Controller
}

func (server *Server) HandleStatus(context *ctx.Context) (interface{}, int, error) {
	return server.Controller, 0, nil
}
