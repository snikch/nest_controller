package server

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/snikch/api/ctx"
	"github.com/snikch/api/fail"
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
	processor := vc.NewActionProcessor()
	server := Server{ctrl}

	router.GET("/status", processor.HandleActionFunc("status", "show status", server.HandleStatus))
	router.POST("/heat/override/on", processor.HandleActionFunc("override", "on", server.HandleOverrideOn))
	router.POST("/heat/override/off", processor.HandleActionFunc("override", "off", server.HandleOverrideOff))
	router.POST("/heat/override/clear", processor.HandleActionFunc("override", "clear", server.HandleOverrideClear))
	router.POST("/zone/:zone_name/override/on", processor.HandleActionFunc("zone_override", "on", server.HandleZoneOverrideOn))
	router.POST("/zone/:zone_name/override/off", processor.HandleActionFunc("zone_override", "off", server.HandleZoneOverrideOff))
	router.POST("/zone/:zone_name/override/clear", processor.HandleActionFunc("zone_override", "clear", server.HandleZoneOverrideClear))
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

// HandleZoneOverrideOn sets the controller override to on.
func (server *Server) HandleZoneOverrideOn(context *ctx.Context) (interface{}, int, error) {
	zone, ok := server.Controller.Zones[vc.ContextParams(context).ByName("zone_name")]
	if !ok {
		return nil, 0, fail.NewValidationError(fmt.Errorf("Zone does not exist"))
	}
	zone.SetOverride(true)
	return nil, http.StatusAccepted, nil
}

// HandleZoneOverrideOff sets the controller override to off.
func (server *Server) HandleZoneOverrideOff(context *ctx.Context) (interface{}, int, error) {
	zone, ok := server.Controller.Zones[vc.ContextParams(context).ByName("zone_name")]
	if !ok {
		return nil, 0, fail.NewValidationError(fmt.Errorf("Zone does not exist"))
	}
	zone.SetOverride(false)
	return nil, http.StatusAccepted, nil
}

// HandleZoneOverrideClear clears the controller override.
func (server *Server) HandleZoneOverrideClear(context *ctx.Context) (interface{}, int, error) {
	zone, ok := server.Controller.Zones[vc.ContextParams(context).ByName("zone_name")]
	if !ok {
		return nil, 0, fail.NewValidationError(fmt.Errorf("Zone does not exist"))
	}
	zone.ClearOverride()
	return nil, http.StatusAccepted, nil
}
