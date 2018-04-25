/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package mmal

import (
	"fmt"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/mmal"
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open creates a new MMAL component
func (config Component) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<mmal.Component>Open{ name=%v }", config.Name)

	if config.Hardware == nil || config.Name == "" {
		return nil, gopi.ErrBadParameter
	}

	// create new MMAL driver
	this := new(component)
	this.log = log
	this.hw = config.Hardware

	// Create component
	if err := mmal_component_create(config.Name, &this.handle); err != nil {
		return nil, err
	}

	// Control port
	this.control = &port{
		handle: mmal_component_control_port(this.handle),
	}

	// Set up ports
	this.input = make([]mmal.Port, mmal_component_num_input_port(this.handle))
	for i := 0; i < len(this.input); i++ {
		this.input[i] = &port{
			handle: mmal_component_input_port(this.handle, uint(i)),
		}
	}
	this.output = make([]mmal.Port, mmal_component_num_output_port(this.handle))
	for i := 0; i < len(this.output); i++ {
		this.output[i] = &port{
			handle: mmal_component_output_port(this.handle, uint(i)),
		}
	}
	this.clock = make([]mmal.Port, mmal_component_num_clock_port(this.handle))
	for i := 0; i < len(this.clock); i++ {
		this.clock[i] = &port{
			handle: mmal_component_clock_port(this.handle, uint(i)),
		}
	}
	return this, nil
}

// Close MMAL connection
func (this *component) Close() error {
	this.log.Debug("<mmal.Component>Close{ name=%v id=0x%08X  }", this.Name(), this.Id())

	if err := mmal_component_destroy(this.handle); err != nil {
		return err
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// COMPONENT INTERFACE IMPLEMENTATION

func (this *component) Name() string {
	return mmal_component_name(this.handle)
}

func (this *component) Id() uint32 {
	return mmal_component_id(this.handle)
}

func (this *component) Enabled() bool {
	return mmal_component_is_enabled(this.handle)
}

func (this *component) SetEnabled(flag bool) error {
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

func (this *component) NumPort() uint {
	return mmal_component_num_port(this.handle)
}

func (this *component) Control() mmal.Port {
	return this.control
}

func (this *component) Input() []mmal.Port {
	return this.input
}

func (this *component) Output() []mmal.Port {
	return this.output
}

func (this *component) Clock() []mmal.Port {
	return this.clock
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *component) String() string {
	return fmt.Sprintf("<mmal.Component>{ name=%v id=0x%08X enabled=%v num_port=%v control_port=%v input_port=%v output_port=%v clock_port=%v }", this.Name(), this.Id(), this.Enabled(), this.NumPort(), this.control, this.input, this.output, this.clock)
}
