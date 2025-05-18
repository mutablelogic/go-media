package ffmpeg

import (
	"encoding/json"
	"strings"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type metaFormat struct {
	Type media.Type `json:"type"`
	Name string     `json:"name"`
}

type Format struct {
	metaFormat
	Input   *ff.AVInputFormat  `json:"input,omitempty"`
	Output  *ff.AVOutputFormat `json:"output,omitempty"`
	Devices []*Device          `json:"devices,omitempty"`
}

type Device struct {
	metaDevice
}

type metaDevice struct {
	Name        string `json:"name" writer:",wrap,width:50"`
	Description string `json:"description" writer:",wrap,width:40"`
	Default     bool   `json:"default,omitempty"`
}

var _ media.Format = &Format{}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newInputFormats(demuxer *ff.AVInputFormat, t media.Type) []media.Format {
	names := strings.Split(demuxer.Name(), ",")
	result := make([]media.Format, 0, len(names))

	// Populate devices by name
	for _, name := range names {
		result = append(result, &Format{
			metaFormat: metaFormat{Type: t, Name: name},
			Input:      demuxer,
		})
	}

	if !t.Is(media.DEVICE) {
		return result
	}

	// Get devices
	list, err := ff.AVDevice_list_input_sources(demuxer, "", nil)
	if err != nil {
		// Bail out if we can't get the list of devices
		return result
	}
	defer ff.AVDevice_free_list_devices(list)

	// Make device list
	devices := make([]*Device, 0, list.NumDevices())
	for i, device := range list.Devices() {
		devices = append(devices, &Device{
			metaDevice{
				Name:        device.Name(),
				Description: device.Description(),
				Default:     list.Default() == i,
			},
		})
	}

	// Append to result
	for _, format := range result {
		format.(*Format).Devices = devices
	}

	// Return result
	return result
}

func newOutputFormats(muxer *ff.AVOutputFormat, t media.Type) []media.Format {
	names := strings.Split(muxer.Name(), ",")
	result := make([]media.Format, 0, len(names))
	for _, name := range names {
		result = append(result, &Format{
			metaFormat: metaFormat{Type: t, Name: name},
			Output:     muxer,
		})
	}

	if !t.Is(media.DEVICE) {
		return result
	}

	// Get devices
	list, err := ff.AVDevice_list_output_sinks(muxer, "", nil)
	if err != nil {
		// Bail out if we can't get the list of devices
		return result
	}
	defer ff.AVDevice_free_list_devices(list)

	// Make device list
	devices := make([]*Device, 0, list.NumDevices())
	for i, device := range list.Devices() {
		devices = append(devices, &Device{
			metaDevice{
				Name:        device.Name(),
				Description: device.Description(),
				Default:     list.Default() == i,
			},
		})
	}

	// Append to result
	for _, format := range result {
		format.(*Format).Devices = devices
	}

	// Return result
	return result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f *Format) String() string {
	data, _ := json.MarshalIndent(f, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (f *Format) Type() media.Type {
	return f.metaFormat.Type
}

func (f *Format) Name() string {
	return f.metaFormat.Name
}

func (f *Format) Description() string {
	switch {
	case f.Input != nil:
		return f.Input.LongName()
	case f.Output != nil:
		return f.Output.LongName()
	default:
		return f.metaFormat.Name
	}
}
