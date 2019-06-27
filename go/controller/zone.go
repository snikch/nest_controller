package controller

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/kidoman/embd"
	"github.com/snikch/api/log"
	"github.com/stianeikeland/go-rpio"
)

// Zone represents a single zone that can optionally be turned on or off.
type Zone struct {
	Name           string        `json:"name"`
	Active         bool          `json:"active"`
	Reading        bool          `json:"reading"`
	Override       *bool         `json:"override"`
	CallForHeatPin uint32        `json:"call_for_heat_pin"`
	DamperOnPin    uint32        `json:"damper_on_pin"`
	DamperOffPin   uint32        `json:"damper_off_pin"`
	Interval       time.Duration `json:"-"`

	callForHeatPin  embd.DigitalPin
	damperOnPin     embd.DigitalPin
	damperOffPin    embd.DigitalPin
	stopCh          chan bool
	currentChangeCh chan<- bool
	running         bool
	events          chan<- Event
}

// NewZone returns an initialized zone.
func NewZone(name string, pin uint32) *Zone {
	return &Zone{
		Name:           name,
		CallForHeatPin: pin,
		Interval:       time.Second,
		stopCh:         make(chan bool),
	}
}

// SetOverride applies an override to the zone.
func (zone *Zone) SetOverride(override bool) {
	zone.Override = &override
	zone.processReading(zone.Reading)
}

// ClearOverride clears any override on the zone.
func (zone *Zone) ClearOverride() {
	zone.Override = nil
	zone.processReading(zone.Reading)
}

// Stop will stop a running zone.
func (zone *Zone) Stop() {
	if zone.running {
		zone.stopCh <- true
	}
	zone.running = false
}

// Run is a blocking function that starts processing pin readings for a zone.
func (zone *Zone) Run(ch chan<- bool, events chan<- Event) error {
	zone.currentChangeCh = ch
	zone.events = events
RUNLOOP:
	for {
		select {
		case <-zone.stopCh:
			break RUNLOOP
		case <-time.NewTimer(zone.Interval).C:
			reading, err := zone.callForHeatPin.Read()
			if err != nil {
				log.Error(fmt.Sprintf("Error reading call for heat pin %d in zone %s: %s", zone.CallForHeatPin, zone.Name, err))
			} else {
				zone.processReading(reading == embd.Low)
			}
		}
	}

	/**
	 * Possible alternate implementation
	 */

	// err := zone.callForHeatPin.Watch(embd.EdgeRising, func(pin embd.DigitalPin) {
	// 	zone.Active = true
	// 	zone.UpdateDamperPins()
	// 	change <- true
	// })
	// if err != nil {
	// 	panic(err)
	// }
	//
	// err = zone.callForHeatPin.Watch(embd.EdgeFalling, func(pin embd.DigitalPin) {
	// 	zone.Active = false
	// 	zone.UpdateDamperPins()
	// 	change <- true
	// })
	// if err != nil {
	// 	panic(err)
	// }
	return nil
}

func (zone *Zone) processReading(target bool) {
	// Toggle and set the damper pins if we've changed.
	log.WithFields(logrus.Fields{
		"zone":    zone.Name,
		"setting": zone.Active,
		"reading": target,
	}).Debug("Read pin")
	didChange := false
	if target && !zone.Active && zone.Override == nil {
		// Set the zone active if we don't have an override.
		zone.Active = true
		didChange = true
	} else if !target && zone.Active && zone.Override == nil {
		// Set the zone inactive if we don't have an override.
		zone.Active = false
		didChange = true
	} else if zone.Override != nil && zone.Active != *zone.Override {
		// If the override isn't set appropriately, do so.
		zone.Active = *zone.Override
		didChange = true
	}

	// Update the pins if required.
	if didChange {
		zone.UpdateDamperPins()
		go func() {
			zone.events <- Event{
				Entity: fmt.Sprintf("zone/%s", zone.Name),
				Value:  fmt.Sprintf("%t", zone.Active),
			}
		}()
		zone.currentChangeCh <- true
	}
	zone.Reading = target
}

// UpdateDamperPins ensure the damper pins represent the current active state.
func (zone *Zone) UpdateDamperPins() {
	onTarget := embd.Low
	offTarget := embd.Low
	if zone.Active {
		onTarget = embd.High
		offTarget = embd.High
	}

	if zone.damperOnPin != nil {
		log.WithFields(logrus.Fields{
			"target": onTarget,
			"zone":   zone.Name,
			"pin":    zone.DamperOnPin,
		}).Info("Setting Damper On Pin")
		zone.damperOnPin.Write(onTarget)
	}

	if zone.damperOffPin != nil {
		log.WithFields(logrus.Fields{
			"target": offTarget,
			"zone":   zone.Name,
			"pin":    zone.DamperOffPin,
		}).Info("Setting Damper Off Pin")
		zone.damperOffPin.Write(offTarget)
	}
	log.WithField("zone", zone.Name).Info("Damper pins updated")
}

func (zone *Zone) initPins() error {
	zone.running = true
	l := log.WithField("zone", zone.Name)
	l.WithField("pin", zone.CallForHeatPin).Info("Initializing call for heat pin")

	// Get the pin for the call for heat.
	pin, err := embd.NewDigitalPin(int(zone.CallForHeatPin))
	if err != nil {
		return err
	}
	pin.SetDirection(embd.In)
	rpin := rpio.Pin(int(zone.CallForHeatPin))
	rpin.PullUp()

	zone.callForHeatPin = pin

	// Set the damper on pin if required.
	if zone.DamperOnPin != 0 {
		l.WithField("pin", zone.DamperOnPin).Info("Initializing damper on pin")
		pin, err := embd.NewDigitalPin(int(zone.DamperOnPin))
		if err != nil {
			return err
		}
		pin.SetDirection(embd.Out)
		rpin := rpio.Pin(int(zone.DamperOnPin))
		rpin.PullUp()
		zone.damperOnPin = pin
	}

	// Set the damper on pin if required.
	if zone.DamperOnPin != 0 {
		l.WithField("pin", zone.DamperOffPin).Info("Initializing damper off pin")
		pin, err := embd.NewDigitalPin(int(zone.DamperOffPin))
		if err != nil {
			return err
		}
		pin.SetDirection(embd.Out)
		rpin := rpio.Pin(int(zone.DamperOffPin))
		rpin.PullUp()
		zone.damperOffPin = pin
	}
	l.Info("Zone initialized")

	zone.UpdateDamperPins()
	return nil
}
