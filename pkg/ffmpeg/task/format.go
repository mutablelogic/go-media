package task

import (
	"context"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all supported formats (input/output formats and devices)
func (manager *Manager) ListFormats(_ context.Context, req *schema.ListFormatRequest) (schema.ListFormatResponse, error) {
	result := make(schema.ListFormatResponse, 0, 256)

	// Filter function
	matches := func(req *schema.ListFormatRequest, f *schema.Format) bool {
		if req == nil {
			return true
		}
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

	// Helper to add devices to an input format
	addInputDevices := func(f *schema.Format, input *ff.AVInputFormat) {
		list, err := ff.AVDevice_list_input_sources(input, "", nil)
		if err != nil || list == nil {
			return
		}
		defer ff.AVDevice_free_list_devices(list)

		devices := make([]schema.Device, 0, list.NumDevices())
		for i, device := range list.Devices() {
			if d := schema.NewDevice(device, list.Default() == i); d != nil {
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
			if d := schema.NewDevice(device, list.Default() == i); d != nil {
				devices = append(devices, *d)
			}
		}
		f.SetDevices(devices)
	}

	// Iterate over all input formats (demuxers)
	var opaque uintptr
	for {
		demuxer := ff.AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}
		if f := schema.NewInputFormat(demuxer, false); f != nil {
			if matches(req, f) {
				result = append(result, *f)
			}
		}
	}

	// Iterate over all output formats (muxers)
	var opaque2 uintptr
	for {
		muxer := ff.AVFormat_muxer_iterate(&opaque2)
		if muxer == nil {
			break
		}
		if f := schema.NewOutputFormat(muxer, false); f != nil {
			if matches(req, f) {
				result = append(result, *f)
			}
		}
	}

	// Track added device names to avoid duplicates (some devices like lavfi handle both audio and video)
	// Store pointers to formats so we can add media types to them
	addedInputDevices := make(map[string]*schema.Format)
	addedOutputDevices := make(map[string]*schema.Format)

	// Iterate over all audio input devices
	for d := ff.AVDevice_input_audio_device_first(); d != nil; d = ff.AVDevice_input_audio_device_next(d) {
		if existing := addedInputDevices[d.Name()]; existing != nil {
			existing.AddMediaType("audio")
			continue
		}
		if f := schema.NewInputFormat(d, true); f != nil {
			f.AddMediaType("audio")
			addInputDevices(f, d)
			if matches(req, f) {
				result = append(result, *f)
				addedInputDevices[d.Name()] = &result[len(result)-1]
			}
		}
	}

	// Iterate over all video input devices
	for d := ff.AVDevice_input_video_device_first(); d != nil; d = ff.AVDevice_input_video_device_next(d) {
		if existing := addedInputDevices[d.Name()]; existing != nil {
			existing.AddMediaType("video")
			continue
		}
		if f := schema.NewInputFormat(d, true); f != nil {
			f.AddMediaType("video")
			addInputDevices(f, d)
			if matches(req, f) {
				result = append(result, *f)
				addedInputDevices[d.Name()] = &result[len(result)-1]
			}
		}
	}

	// Iterate over all audio output devices
	for d := ff.AVDevice_output_audio_device_first(); d != nil; d = ff.AVDevice_output_audio_device_next(d) {
		if existing := addedOutputDevices[d.Name()]; existing != nil {
			existing.AddMediaType("audio")
			continue
		}
		if f := schema.NewOutputFormat(d, true); f != nil {
			f.AddMediaType("audio")
			addOutputDevices(f, d)
			if matches(req, f) {
				result = append(result, *f)
				addedOutputDevices[d.Name()] = &result[len(result)-1]
			}
		}
	}

	// Iterate over all video output devices
	for d := ff.AVDevice_output_video_device_first(); d != nil; d = ff.AVDevice_output_video_device_next(d) {
		if existing := addedOutputDevices[d.Name()]; existing != nil {
			existing.AddMediaType("video")
			continue
		}
		if f := schema.NewOutputFormat(d, true); f != nil {
			f.AddMediaType("video")
			addOutputDevices(f, d)
			if matches(req, f) {
				result = append(result, *f)
				addedOutputDevices[d.Name()] = &result[len(result)-1]
			}
		}
	}

	return result, nil
}
