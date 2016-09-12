package controller

import (
	"fmt"
	"strings"

	"github.com/kidoman/embd"
	"github.com/snikch/api/log"
	"github.com/stianeikeland/go-rpio"
)

// Controller represents a fan in thermostat controller for multiple zones.
type Controller struct {
	Running       bool `json:"running"`
	Heat          `json:"heat"`
	Zones         map[string]*Zone `json:"zones"`
	change        chan bool
	Events        chan Event     `json:"-"`
	EventHandlers []EventHandler `json:"-"`
}

// Heat represents the current heat state.
type Heat struct {
	Override *bool  `json:"override"`
	Active   bool   `json:"active"`
	Pin      uint32 `json:"pin"`
	pin      embd.DigitalPin
}

// NewController returns an initialized controller on the supplied pin.
func NewController(pin uint32) *Controller {
	controller := &Controller{
		Events: make(chan Event),
		Zones:  map[string]*Zone{},
		Heat: Heat{
			Pin: pin,
		},
		EventHandlers: []EventHandler{},
		change:        make(chan bool),
	}
	go controller.HandleEvents()
	return controller
}

// HandleEvents is a blocking function that pulls events off the event handler
// channel and pushes them to the various event handlers.
func (controller *Controller) HandleEvents() {
	for {
		event := <-controller.Events
		log.WithField("event", event).Info("Received event")
		for _, handler := range controller.EventHandlers {
			err := handler.Push(event)
			if err != nil {
				log.WithError(err).WithField("handler", handler.Name).Error("Failed to push event")
			}
		}
	}
}

// AddEventHandler adds an event handler to the controllers list.
func (controller *Controller) AddEventHandler(handler EventHandler) {
	controller.EventHandlers = append(controller.EventHandlers, handler)
}

// AddZone adds a zone and starts running it.
func (controller *Controller) AddZone(zone *Zone) {
	controller.Zones[strings.ToLower(zone.Name)] = zone
}

// Run starts the controller running and listening for zone changes.
func (controller *Controller) Run() error {
	go controller.run()
	err := controller.initPins()
	if err != nil {
		return err
	}

	// Turn all zones on.
	for _, zone := range controller.Zones {
		log.WithField("zone", zone.Name).Info("Starting zone")
		err := zone.initPins()
		if err != nil {
			return err
		}
		go zone.Run(controller.change, controller.Events)
		log.WithField("zone", zone.Name).Info("Zone started")
	}

	// Mark the controller as running.
	controller.Running = true
	return nil
}

func (controller *Controller) run() {
	// Start a loop to check on the zones when a change is emitted.
	for {
		log.Info("Controller is waiting")
		<-controller.change
		log.Info("Controller received change event")
		targetState := false

		// First check for any override, and apply that over others.
		if controller.Heat.Override != nil {
			targetState = *controller.Heat.Override
		} else {
			// If we're not overriding, use the zone states.
			for _, zone := range controller.Zones {
				if zone.Active {
					targetState = true
					break
				}
			}
		}

		didChange := false
		if targetState && !controller.Heat.Active {
			controller.Heat.Active = true
			controller.SetHeaterState()
			didChange = true
		} else if !targetState && controller.Heat.Active {
			controller.Heat.Active = false
			controller.SetHeaterState()
			didChange = true
		}

		// Submit a heat change event.
		if didChange {
			go func() {
				controller.Events <- Event{
					Entity: "heat",
					Value:  fmt.Sprintf("%t", controller.Heat.Active),
				}
			}()
		}
	}
}

// SetOverride allows a controller wide override of the target state.
func (controller *Controller) SetOverride(override bool) {
	// Set the new override value.
	controller.Heat.Override = &override
	// Let the controller know a change has occurred.
	controller.change <- true
}

// ClearOverride clears an existing override if one exists.
func (controller *Controller) ClearOverride() {
	// Clear the override value.
	controller.Heat.Override = nil
	// Let the controller know a change has occurred.
	controller.change <- true
}

// SetHeaterState will adjust the heater pin to the correct high / low value.
func (controller *Controller) SetHeaterState() {
	if controller.Heat.Active {
		controller.pin.Write(embd.High)
	} else {
		controller.pin.Write(embd.Low)
	}
}

func (controller *Controller) initPins() error {
	// Get the pin for the call for heat.
	pin, err := embd.NewDigitalPin(int(controller.Pin))
	if err != nil {
		return err
	}
	pin.SetDirection(embd.Out)
	rpin := rpio.Pin(int(controller.Pin))
	rpin.PullUp()
	controller.pin = pin
	return nil
}
