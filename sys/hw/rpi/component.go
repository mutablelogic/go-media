/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package rpi

import (
	"fmt"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
    #cgo CFLAGS: -I/opt/vc/include -I/opt/vc/include/interface/mmal
    #cgo LDFLAGS:  -L/opt/vc/lib -lmmal -lmmal_components -lmmal_core
	#include <mmal.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MMALComponent struct {
	// Hardware device
	Hardware gopi.Hardware
	// name of the component to create
	Name string
}

type mmalComponent struct {
	log    gopi.Logger
	hw     gopi.Hardware
	name   string
	handle mmalComponentHandle
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open creates a new MMAL component
func (config MMALComponent) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<mmal.Component>Open{ name=%v }", config.Name)

	if config.Hardware == nil || config.Name == "" {
		return nil, gopi.ErrBadParameter
	}

	// create new MMAL driver
	this := new(mmalComponent)
	this.log = log
	this.name = config.Name
	this.hw = config.Hardware

	// Create component
	if err := mmal_component_create(this.name, &this.handle); err != nil {
		return nil, err
	}

	return this, nil
}

// Close MMAL connection
func (this *mmalComponent) Close() error {
	this.log.Debug("<mmal.Component>Close{ name=%v id=0x%08X  }", this.Name(), this.Id())

	if err := mmal_component_destroy(this.handle); err != nil {
		return err
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *mmalComponent) String() string {
	return fmt.Sprintf("<mmal.Component>{ name=%v id=0x%08X enabled=%v }", this.Name(), this.Id(), this.Enabled())
}

////////////////////////////////////////////////////////////////////////////////
// INTERFACE IMPLEMENTATION

func (this *mmalComponent) Name() string {
	return this.name
}

func (this *mmalComponent) Id() uint32 {
	return uint32(this.handle.id)
}

func (this *mmalComponent) Enabled() bool {
	return (this.handle.is_enabled != 0)
}

func (this *mmalComponent) SetEnabled(flag bool) error {
	if flag {
		if err := mmal_component_enable(this.handle); err != nil {
			return err
		}
	} else {
		if err := mmal_component_disable(this.handle); err != nil {
			return err
		}
	}
	return nil
}
