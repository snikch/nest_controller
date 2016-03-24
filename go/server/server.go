package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/snikch/api/ctx"
	"github.com/snikch/api/render"
	"github.com/snikch/api/vc"
	"github.com/snikch/nest/go/controller"
)

func init() {
	// Ensure our output is json.
	vc.DefaultRenderer = render.JSONRenderer{}
}

// NewServer returns a fully initialized http.Handler.
func NewServer(ctrl *controller.Controller) http.Handler {
	router := httprouter.New()
	processor := vc.ActionProcessor{}
	server := Server{ctrl}

	router.GET("/status", processor.HandleActionFunc("status", server.HandleStatus))
	router.POST("/heat/override/on", processor.HandleActionFunc("override", server.HandleOverrideOn))
	router.GET("/heat/override/off", processor.HandleActionFunc("override", server.HandleOverrideOff))
	router.GET("/heat/override/clear", processor.HandleActionFunc("override", server.HandleOverrideClear))
	return router
}

// Server holds access to the supplied controller.
type Server struct {
	Controller *controller.Controller
}

// HandleStatus returns the controller, which marshalls safely into the current
// state of the controller system.
func (server *Server) HandleStatus(context *ctx.Context) (interface{}, int, error) {
	return server.Controller, 0, nil
}

// HandleOverrideOn sets the controller override to on.
func (server *Server) HandleOverrideOn(context *ctx.Context) (interface{}, int, error) {
	server.Controller.SetOverride(true)
	return nil, http.StatusAccepted, nil
}

// HandleOverrideOff sets the controller override to off.
func (server *Server) HandleOverrideOff(context *ctx.Context) (interface{}, int, error) {
	server.Controller.SetOverride(false)
	return nil, http.StatusAccepted, nil
}

// HandleOverrideClear clears the controller override.
func (server *Server) HandleOverrideClear(context *ctx.Context) (interface{}, int, error) {
	server.Controller.ClearOverride()
	return nil, http.StatusAccepted, nil
}
