package manager

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListDevices returns all devices available for the registered input/output
// audio/video device formats (e.g. avfoundation, alsa, pulse).
func (profile *Profile) ListDevices(ctx context.Context, req schema.DeviceListRequest) (_ schema.DeviceList, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "ListDevices",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	result := make(schema.DeviceList, 0, 16)

	matches := func(d *schema.Device) bool {
		if req.Name != nil && d.Name != *req.Name {
			return false
		}
		if req.Format != nil && d.Format != *req.Format {
			return false
		}
		if req.IsInput != nil && d.IsInput != *req.IsInput {
			return false
		}
		if req.IsOutput != nil && d.IsOutput != *req.IsOutput {
			return false
		}
		return true
	}

	addInputDevices := func(format string, input *ff.AVInputFormat) {
		list, err := ff.AVDevice_list_input_sources(input, "", nil)
		if err != nil || list == nil {
			return
		}
		defer ff.AVDevice_free_list_devices(list)
		for i, device := range list.Devices() {
			if d := schema.NewDevice(format, true, false, device, i, list.Default() == i); d != nil && matches(d) {
				result = append(result, *d)
			}
		}
	}

	addOutputDevices := func(format string, output *ff.AVOutputFormat) {
		list, err := ff.AVDevice_list_output_sinks(output, "", nil)
		if err != nil || list == nil {
			return
		}
		defer ff.AVDevice_free_list_devices(list)
		for i, device := range list.Devices() {
			if d := schema.NewDevice(format, false, true, device, i, list.Default() == i); d != nil && matches(d) {
				result = append(result, *d)
			}
		}
	}

	// Device formats are returned by separate audio/video iterators, but the
	// underlying driver (e.g. avfoundation) is the same one either way, so
	// dedupe by format name to avoid listing its devices twice.
	seenInput := make(map[string]bool)
	for d := ff.AVDevice_input_audio_device_first(); d != nil; d = ff.AVDevice_input_audio_device_next(d) {
		if name := d.Name(); !seenInput[name] {
			seenInput[name] = true
			addInputDevices(name, d)
		}
	}
	for d := ff.AVDevice_input_video_device_first(); d != nil; d = ff.AVDevice_input_video_device_next(d) {
		if name := d.Name(); !seenInput[name] {
			seenInput[name] = true
			addInputDevices(name, d)
		}
	}

	seenOutput := make(map[string]bool)
	for d := ff.AVDevice_output_audio_device_first(); d != nil; d = ff.AVDevice_output_audio_device_next(d) {
		if name := d.Name(); !seenOutput[name] {
			seenOutput[name] = true
			addOutputDevices(name, d)
		}
	}
	for d := ff.AVDevice_output_video_device_first(); d != nil; d = ff.AVDevice_output_video_device_next(d) {
		if name := d.Name(); !seenOutput[name] {
			seenOutput[name] = true
			addOutputDevices(name, d)
		}
	}

	return result, nil
}
