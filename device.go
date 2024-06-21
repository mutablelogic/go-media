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
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Default     bool      `json:"default,omitempty"`
	MediaType   MediaType `json:"type,omitempty" writer:",wrap,width:21"`
}

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newInputDevice(ctx *ff.AVInputFormat, d *ff.AVDeviceInfo, t MediaType, def bool) *device {
	meta := &devicemeta{
		Format:      ctx.Name(),
		Name:        d.Name(),
		Description: d.Description(),
		Default:     def,
		MediaType:   INPUT | t,
	}
	return &device{devicemeta: *meta}
}

func newOutputDevice(ctx *ff.AVOutputFormat, d *ff.AVDeviceInfo, t MediaType, def bool) *device {
	meta := &devicemeta{
		Format:      ctx.Name(),
		Name:        d.Name(),
		Description: d.Description(),
		Default:     def,
		MediaType:   OUTPUT | t,
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
	return v.devicemeta.MediaType
}

// Whether this is the default device
func (v *device) Default() bool {
	return v.devicemeta.Default
}
