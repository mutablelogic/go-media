package manager

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"sync"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// Example line: [AVFoundation indev @ 0x...] [0] FaceTime HD Camera
	avfoundationDevicePattern = regexp.MustCompile(`\[(\d+)\]\s+(.+)$`)
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
		var devices []*schema.Device
		if input.Name() == "avfoundation" {
			devices = enumerateAVFoundationDevices(format, input)
		} else {
			list, err := ff.AVDevice_list_input_sources(input, "", nil)
			if err != nil || list == nil {
				return
			}
			defer ff.AVDevice_free_list_devices(list)
			for i, device := range list.Devices() {
				if d := schema.NewDevice(format, true, false, device, i, list.Default() == i); d != nil {
					devices = append(devices, d)
				}
			}
		}
		for _, d := range devices {
			if matches(d) {
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

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// enumerateAVFoundationDevices opens avfoundation with list_devices=true and
// parses FFmpeg logs to build device entries, since AVDevice_list_input_sources
// doesn't return meaningful results for avfoundation.
func enumerateAVFoundationDevices(format string, input *ff.AVInputFormat) []*schema.Device {
	var devices []*schema.Device
	var mu sync.Mutex
	var capturedLines []string

	oldLevel := ff.AVUtil_log_get_level()
	ff.AVUtil_log_set_level(ff.AV_LOG_VERBOSE)
	defer ff.AVUtil_log_set_level(oldLevel)

	ff.AVUtil_log_set_callback(func(level ff.AVLog, message string, userInfo any) {
		if level < ff.AV_LOG_INFO || level > ff.AV_LOG_VERBOSE {
			return
		}

		// FFmpeg output format can vary by version/build. Capture AVFoundation
		// tagged lines, section headers, and raw "[index] device" lines.
		line := strings.TrimSpace(message)
		lower := strings.ToLower(line)
		if !(strings.Contains(lower, "avfoundation") ||
			strings.Contains(lower, "video devices:") ||
			strings.Contains(lower, "audio devices:") ||
			avfoundationDevicePattern.MatchString(line)) {
			return
		}
		mu.Lock()
		capturedLines = append(capturedLines, line)
		mu.Unlock()
	})
	defer ff.AVUtil_log_set_callback(nil)

	options := ff.AVUtil_dict_alloc()
	if options == nil {
		return devices
	}
	defer ff.AVUtil_dict_free(options)

	ff.AVUtil_dict_set(options, "list_devices", "true", 0)

	ctx, _ := ff.AVFormat_open_device(input, options)
	if ctx != nil {
		ff.AVFormat_find_stream_info(ctx, nil)
		ff.AVFormat_close_input(ctx)
	}

	mu.Lock()
	defer mu.Unlock()

	var currentType string
	defaultSeen := make(map[string]bool)

	for _, line := range capturedLines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "video devices:") {
			currentType = "video"
			continue
		}
		if strings.Contains(line, "audio devices:") {
			currentType = "audio"
			continue
		}

		matches := avfoundationDevicePattern.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}
		deviceName := strings.TrimSpace(matches[2])
		if deviceName == "" {
			continue
		}

		deviceIndex := 0
		if idx := matches[1]; idx != "" {
			if n, err := strconv.Atoi(idx); err == nil {
				deviceIndex = n
			}
		}

		isDefault := false
		if deviceIndex == 0 && currentType != "" && !defaultSeen[currentType] {
			isDefault = true
			defaultSeen[currentType] = true
		}

		device := &schema.Device{
			Format:      format,
			Index:       deviceIndex,
			Name:        deviceName,
			Description: deviceName,
			IsDefault:   isDefault,
			IsInput:     true,
		}
		if currentType != "" {
			device.MediaTypes = []string{currentType}
		}

		devices = append(devices, device)
	}

	return devices
}
