package manager

import (
	"context"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListFormats returns all supported formats (input/output formats and devices).
func (m *Media) ListFormats(_ context.Context, req schema.ListFormatRequest) (schema.ListFormatResponse, error) {
	result := make(schema.ListFormatResponse, 0, 256)

	matches := func(f *schema.Format) bool {
		if req.Name != "" && !strings.Contains(f.Name, req.Name) {
			return false
		}
		if req.IsInput != nil && *req.IsInput != f.IsInput {
			return false
		}
		if req.IsOutput != nil && *req.IsOutput != f.IsOutput {
			return false
		}
		if req.IsDevice != nil && *req.IsDevice != f.IsDevice {
			return false
		}
		return true
	}

	addInputDevices := func(f *schema.Format, input *ff.AVInputFormat) {
		if input.Name() == "avfoundation" {
			devices := enumerateAVFoundationDevices(input)
			if len(devices) > 0 {
				f.SetDevices(devices)
			}
			return
		}

		list, err := ff.AVDevice_list_input_sources(input, "", nil)
		if err != nil || list == nil {
			return
		}
		defer ff.AVDevice_free_list_devices(list)

		devices := make([]schema.Device, 0, list.NumDevices())
		for i, device := range list.Devices() {
			if d := schema.NewDevice(device, i, list.Default() == i); d != nil {
				devices = append(devices, *d)
			}
		}
		f.SetDevices(devices)
	}

	addOutputDevices := func(f *schema.Format, output *ff.AVOutputFormat) {
		list, err := ff.AVDevice_list_output_sinks(output, "", nil)
		if err != nil || list == nil {
			return
		}
		defer ff.AVDevice_free_list_devices(list)

		devices := make([]schema.Device, 0, list.NumDevices())
		for i, device := range list.Devices() {
			if d := schema.NewDevice(device, i, list.Default() == i); d != nil {
				devices = append(devices, *d)
			}
		}
		f.SetDevices(devices)
	}

	var opaque uintptr
	for {
		demuxer := ff.AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}
		if f := schema.NewInputFormat(demuxer, false); f != nil {
			if f.IsDevice {
				addInputDevices(f, demuxer)
			}
			if matches(f) {
				result = append(result, *f)
			}
		}
	}

	var opaque2 uintptr
	for {
		muxer := ff.AVFormat_muxer_iterate(&opaque2)
		if muxer == nil {
			break
		}
		if f := schema.NewOutputFormat(muxer, false); f != nil {
			if f.IsDevice {
				addOutputDevices(f, muxer)
			}
			if matches(f) {
				result = append(result, *f)
			}
		}
	}

	addedInputDevices := make(map[string]*schema.Format)
	addedOutputDevices := make(map[string]*schema.Format)

	for d := ff.AVDevice_input_audio_device_first(); d != nil; d = ff.AVDevice_input_audio_device_next(d) {
		if existing := addedInputDevices[d.Name()]; existing != nil {
			existing.AddMediaType("audio")
			continue
		}
		if f := schema.NewInputFormat(d, true); f != nil {
			f.AddMediaType("audio")
			addInputDevices(f, d)
			if matches(f) {
				result = append(result, *f)
				addedInputDevices[d.Name()] = &result[len(result)-1]
			}
		}
	}

	for d := ff.AVDevice_input_video_device_first(); d != nil; d = ff.AVDevice_input_video_device_next(d) {
		if existing := addedInputDevices[d.Name()]; existing != nil {
			existing.AddMediaType("video")
			continue
		}
		if f := schema.NewInputFormat(d, true); f != nil {
			f.AddMediaType("video")
			addInputDevices(f, d)
			if matches(f) {
				result = append(result, *f)
				addedInputDevices[d.Name()] = &result[len(result)-1]
			}
		}
	}

	for d := ff.AVDevice_output_audio_device_first(); d != nil; d = ff.AVDevice_output_audio_device_next(d) {
		if existing := addedOutputDevices[d.Name()]; existing != nil {
			existing.AddMediaType("audio")
			continue
		}
		if f := schema.NewOutputFormat(d, true); f != nil {
			f.AddMediaType("audio")
			addOutputDevices(f, d)
			if matches(f) {
				result = append(result, *f)
				addedOutputDevices[d.Name()] = &result[len(result)-1]
			}
		}
	}

	for d := ff.AVDevice_output_video_device_first(); d != nil; d = ff.AVDevice_output_video_device_next(d) {
		if existing := addedOutputDevices[d.Name()]; existing != nil {
			existing.AddMediaType("video")
			continue
		}
		if f := schema.NewOutputFormat(d, true); f != nil {
			f.AddMediaType("video")
			addOutputDevices(f, d)
			if matches(f) {
				result = append(result, *f)
				addedOutputDevices[d.Name()] = &result[len(result)-1]
			}
		}
	}

	return result, nil
}
