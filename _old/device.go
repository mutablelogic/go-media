package media

import (
	"encoding/json"

	// Package imports
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type device struct {
	devicemeta
}

type devicemeta struct {
	Format      string    `json:"format"`
	Name        string    `json:"name" writer:",wrap,width:50"`
	Description string    `json:"description" writer:",wrap,width:40"`
	Default     bool      `json:"default,omitempty"`
	Type        MediaType `json:"type,omitempty" writer:",wrap,width:21"`
}

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newDevice(name string, d *ff.AVDeviceInfo, t MediaType, def bool) *device {
	meta := &devicemeta{
		Format:      name,
		Name:        d.Name(),
		Description: d.Description(),
		Default:     def,
		Type:        DEVICE | t,
	}
	return &device{devicemeta: *meta}
}

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v *device) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Device name, format depends on the device
func (v *device) Name() string {
	return v.devicemeta.Name
}

// Description of the device
func (v *device) Description() string {
	return v.devicemeta.Description
}

// Flags indicating the type INPUT or OUTPUT, AUDIO or VIDEO
func (v *device) Type() MediaType {
	return v.devicemeta.Type
}

// Whether this is the default device
func (v *device) Default() bool {
	return v.devicemeta.Default
}
