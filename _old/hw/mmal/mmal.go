/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package mmal

import (
	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/djthorpe/mmal"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Component struct {
	// Hardware device
	Hardware gopi.Hardware
	// name of the component to create
	Name string
}

type port struct {
	handle portHandle
}

type component struct {
	log     gopi.Logger
	hw      gopi.Hardware
	handle  componentHandle
	control mmal.Port
	input   []mmal.Port
	output  []mmal.Port
	clock   []mmal.Port
}

type streamformat struct {
	handle esFormatHandle
}

type paramType uint32
