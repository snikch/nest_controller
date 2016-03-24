package controller

import (
	"fmt"
	"time"

	"github.com/kidoman/embd"
	"github.com/snikch/api/log"
)

// Zone represents a single zone that can optionally be turned on or off.
type Zone struct {
	Name           string        `json:"name"`
	Active         bool          `json:"active"`
	CallForHeatPin uint32        `json:"call_for_heat_pin"`
	DamperOnPin    *uint32       `json:"damper_on_pin"`
	DamperOffPin   *uint32       `json:"damper_off_pin"`
	Interval       time.Duration `json:"-"`
	callForHeatPin embd.DigitalPin
	damperOnPin    embd.DigitalPin
	damperOffPin   embd.DigitalPin
}

// NewZone returns an initialized zone.
func NewZone(name string, pin uint32) *Zone {
	return &Zone{
		Name:           name,
		CallForHeatPin: pin,
		Interval:       time.Second,
	}
}

func (zone *Zone) run(change chan<- bool) error {
	err := zone.initPins()
	if err != nil {
		return err
	}
	zone.UpdateDamperPins()

	for {
		reading, err := zone.callForHeatPin.Read()
		if err != nil {
			log.Error(fmt.Sprintf("Error reading call for heat pin %d in zone %s: %s", zone.CallForHeatPin, zone.Name, err))
		} else {
			// Toggle and set the damper pins if we've changed.
			if reading == embd.High && !zone.Active {
				zone.Active = true
				zone.UpdateDamperPins()
				change <- true
			} else if reading == embd.Low && zone.Active {
				zone.Active = false
				zone.UpdateDamperPins()
				change <- true
			}
		}
		time.Sleep(zone.Interval)
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

// UpdateDamperPins ensure the damper pins represent the current active state.
func (zone *Zone) UpdateDamperPins() {
	onTarget := embd.Low
	offTarget := embd.High
	if zone.Active {
		onTarget = embd.High
		offTarget = embd.Low
	}

	if zone.damperOnPin != nil {
		zone.damperOnPin.Write(onTarget)
	}

	if zone.damperOffPin != nil {
		zone.damperOffPin.Write(offTarget)
	}
}

func (zone *Zone) initPins() error {
	// Get the pin for the call for heat.
	pin, err := embd.NewDigitalPin(zone.CallForHeatPin)
	if err != nil {
		return err
	}
	pin.SetDirection(embd.In)
	zone.callForHeatPin = pin

	// Set the damper on pin if required.
	if zone.DamperOnPin != nil {
		pin, err := embd.NewDigitalPin(zone.DamperOnPin)
		if err != nil {
			return err
		}
		pin.SetDirection(embd.Out)
		zone.damperOnPin = pin
	}

	// Set the damper on pin if required.
	if zone.DamperOnPin != nil {
		pin, err := embd.NewDigitalPin(zone.DamperOffPin)
		if err != nil {
			return err
		}
		pin.SetDirection(embd.Out)
		zone.damperOffPin = pin
	}
	return nil
}
