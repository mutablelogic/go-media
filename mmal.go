/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package media

import (
	"encoding/binary"
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	PortType           uint32
	StreamType         uint32
	PortCapabilityType uint32
	EncodingType       uint32
)

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

	// Return port information
	NumPort() uint
	Control() Port
	Input() []Port
	Output() []Port
	Clock() []Port
}

type Port interface {
	// Return port information
	Name() string
	Type() PortType
	Index() uint
	Capabilities() PortCapabilityType

	// Enable & Disable port
	Enabled() bool
	SetEnabled(bool) error

	// Connect and disconnect this port to another
	Connect(Port) error
	Disconnect() error

	// Flush port, commit format changes
	Flush() error
	Commit() error

	// Implements common parameters
	CommonParameters
}

type CommonParameters interface {

	// Get Parameters
	SupportedEncodings() ([]EncodingType, error)
	ZeroCopy() (bool, error)
	PowerMonEnable() (bool, error)
	NoImagePadding() (bool, error)
	LockstepEnable() (bool, error)

	// Set Parameters
	SetZeroCopy(bool) error
	SetPowerMonEnable(bool) error
	SetNoImagePadding(bool) error
	SetLockstepEnable(bool) error
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

const (
	MMAL_ES_TYPE_NONE       StreamType = 0x0000
	MMAL_ES_TYPE_CONTROL    StreamType = 0x0001
	MMAL_ES_TYPE_AUDIO      StreamType = 0x0002
	MMAL_ES_TYPE_VIDEO      StreamType = 0x0003
	MMAL_ES_TYPE_SUBPICTURE StreamType = 0x0004
)

const (
	MMAL_PORT_CAPABILITY_NONE                         PortCapabilityType = 0x00
	MMAL_PORT_CAPABILITY_PASSTHROUGH                  PortCapabilityType = 0x01
	MMAL_PORT_CAPABILITY_ALLOCATION                   PortCapabilityType = 0x02
	MMAL_PORT_CAPABILITY_SUPPORTS_EVENT_FORMAT_CHANGE PortCapabilityType = 0x04
	MMAL_PORT_CAPABILITY_MIN                          PortCapabilityType = MMAL_PORT_CAPABILITY_PASSTHROUGH
	MMAL_PORT_CAPABILITY_MAX                          PortCapabilityType = MMAL_PORT_CAPABILITY_SUPPORTS_EVENT_FORMAT_CHANGE
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

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t PortType) String() string {
	switch t {
	case MMAL_PORT_NONE:
		return "MMAL_PORT_NONE"
	case MMAL_PORT_CONTROL:
		return "MMAL_PORT_CONTROL"
	case MMAL_PORT_INPUT:
		return "MMAL_PORT_INPUT"
	case MMAL_PORT_OUTPUT:
		return "MMAL_PORT_OUTPUT"
	case MMAL_PORT_CLOCK:
		return "MMAL_PORT_CLOCK"
	default:
		return "[?? Invalid PortType value]"
	}
}

func (t PortCapabilityType) String() string {
	if t == MMAL_PORT_CAPABILITY_NONE {
		return "MMAL_PORT_CAPABILITY_NONE"
	}
	v := ""
	for b := MMAL_PORT_CAPABILITY_MIN; b <= MMAL_PORT_CAPABILITY_MAX; b <<= 1 {
		if t&b == 0 {
			continue
		}
		switch b {
		case MMAL_PORT_CAPABILITY_PASSTHROUGH:
			v += "MMAL_PORT_CAPABILITY_PASSTHROUGH|"
		case MMAL_PORT_CAPABILITY_ALLOCATION:
			v += "MMAL_PORT_CAPABILITY_ALLOCATION|"
		case MMAL_PORT_CAPABILITY_SUPPORTS_EVENT_FORMAT_CHANGE:
			v += "MMAL_PORT_CAPABILITY_SUPPORTS_EVENT_FORMAT_CHANGE|"
		}
	}
	return strings.TrimSuffix(v, "|")
}

func (s StreamType) String() string {
	switch s {
	case MMAL_ES_TYPE_NONE:
		return "MMAL_ES_TYPE_NONE"
	case MMAL_ES_TYPE_CONTROL:
		return "MMAL_ES_TYPE_CONTROL"
	case MMAL_ES_TYPE_AUDIO:
		return "MMAL_ES_TYPE_AUDIO"
	case MMAL_ES_TYPE_VIDEO:
		return "MMAL_ES_TYPE_VIDEO"
	case MMAL_ES_TYPE_SUBPICTURE:
		return "MMAL_ES_TYPE_SUBPICTURE"
	default:
		return "[?? Invalid StreamType value]"
	}
}

func (e EncodingType) String() string {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(e))
	return "'" + string(buf) + "'"
}
