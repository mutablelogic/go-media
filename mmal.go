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
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type PortType uint32

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

type Component interface {
	gopi.Driver

	// Return component name and ID
	Name() string
	Id() uint32

	// Enable component
	Enabled() bool
	SetEnabled(bool) error
}

type Port interface {
	// Return port type & name
	Type() PortType
	Name() string

	// Enable & Disable port
	Enabled() bool
	SetEnabled(bool) error

	// Connect and disconnect
	Connect(Port) error
	Disconnect(Port) error

	// Flush port
	Flush() error
}

////////////////////////////////////////////////////////////////////////////////
// PORT TYPES

const (
	MMAL_PORT_NONE    PortType = 0x0000
	MMAL_PORT_CONTROL PortType = 0x0001
	MMAL_PORT_INPUT   PortType = 0x0002
	MMAL_PORT_OUTPUT  PortType = 0x0003
	MMAL_PORT_CLOCK   PortType = 0x0004
)

////////////////////////////////////////////////////////////////////////////////
// COMPONENT NAMES

const (
	MMAL_COMPONENT_DEFAULT_VIDEO_DECODER   = "vc.ril.video_decode"
	MMAL_COMPONENT_DEFAULT_VIDEO_ENCODER   = "vc.ril.video_encode"
	MMAL_COMPONENT_DEFAULT_VIDEO_RENDERER  = "vc.ril.video_render"
	MMAL_COMPONENT_DEFAULT_IMAGE_DECODER   = "vc.ril.image_decode"
	MMAL_COMPONENT_DEFAULT_IMAGE_ENCODER   = "vc.ril.image_encode"
	MMAL_COMPONENT_DEFAULT_CAMERA          = "vc.ril.camera"
	MMAL_COMPONENT_DEFAULT_VIDEO_CONVERTER = "vc.video_convert"
	MMAL_COMPONENT_DEFAULT_SPLITTER        = "vc.splitter"
	MMAL_COMPONENT_DEFAULT_SCHEDULER       = "vc.scheduler"
	MMAL_COMPONENT_DEFAULT_VIDEO_INJECTER  = "vc.video_inject"
	MMAL_COMPONENT_DEFAULT_VIDEO_SPLITTER  = "vc.ril.video_splitter"
	MMAL_COMPONENT_DEFAULT_AUDIO_DECODER   = "none"
	MMAL_COMPONENT_DEFAULT_AUDIO_RENDERER  = "vc.ril.audio_render"
	MMAL_COMPONENT_DEFAULT_MIRACAST        = "vc.miracast"
	MMAL_COMPONENT_DEFAULT_CLOCK           = "vc.clock"
	MMAL_COMPONENT_DEFAULT_CAMERA_INFO     = "vc.camera_info"
)
