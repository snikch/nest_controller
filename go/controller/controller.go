package controller

import (
	"strings"

	"github.com/kidoman/embd"
)

// Controller represents a fan in thermostat controller for multiple zones.
type Controller struct {
	Zones           map[string]*Zone `json:"zones"`
	Active          bool             `json:"active"`
	heaterPinNumber uint32
	heaterPin       embd.DigitalPin
	change          chan bool
}

// NewController returns an initialized controller on the supplied pin.
func NewController(pin uint32) *Controller {
	return &Controller{
		Zones:           map[string]*Zone{},
		heaterPinNumber: pin,
		change:          make(chan bool),
	}
}

// AddZone adds a zone and starts running it.
func (controller *Controller) AddZone(zone *Zone) {
	controller.Zones[strings.ToLower(zone.Name)] = zone
	zone.run(controller.change)
}

// Run starts the controller running and listening for zone changes.
func (controller *Controller) Run() {
	// Start GPIO and defer its closing to ensure cleanup.
	embd.InitGPIO()
	defer embd.CloseGPIO()
	controller.initPins()

	// Start a loop to check on the zones when a change is emitted.
	for {
		<-controller.change
		anyActive := false
		for _, zone := range controller.Zones {
			if zone.Active {
				anyActive = true
				if !controller.Active {
					controller.Active = true
					controller.SetHeaterState()
				}
				break
			}
		}
		if !anyActive && controller.Active {
			controller.Active = false
			controller.SetHeaterState()
		}
	}
}

// SetHeaterState will adjust the heater pin to the correct high / low value.
func (controller *Controller) SetHeaterState() {
	if controller.Active {
		controller.heaterPin.Write(embd.High)
	} else {
		controller.heaterPin.Write(embd.Low)
	}
}

func (controller *Controller) initPins() {
	// Get the pin for the call for heat.
	pin, err := embd.NewDigitalPin(controller.heaterPinNumber)
	if err != nil {
		panic(err)
	}
	pin.SetDirection(embd.Out)
	controller.heaterPin = pin
}
