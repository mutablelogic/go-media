package task

import (
	"regexp"
	"strconv"
	"strings"
	"sync"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// Pattern to match AVFoundation device list output
	// Example: [AVFoundation indev @ 0x...] [0] FaceTime HD Camera
	avfoundationDevicePattern = regexp.MustCompile(`\[(\d+)\]\s+(.+)$`)
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Enumerate AVFoundation devices by opening the format with -list_devices option
// and parsing the log output
func enumerateAVFoundationDevices(format *ff.AVInputFormat) []schema.Device {
	var devices []schema.Device
	var mu sync.Mutex
	var capturedLines []string

	// Set up a log callback to capture device list output
	oldLevel := ff.AVUtil_log_get_level()
	ff.AVUtil_log_set_level(ff.AV_LOG_VERBOSE)
	defer ff.AVUtil_log_set_level(oldLevel)

	ff.AVUtil_log_set_callback(func(level ff.AVLog, message string, userInfo any) {
		// Filter by log level to reduce unrelated messages
		if level < ff.AV_LOG_INFO || level > ff.AV_LOG_VERBOSE {
			return
		}

		// Further restrict to AVFoundation-related messages
		if !strings.Contains(message, "AVFoundation") {
			return
		}

		mu.Lock()
		capturedLines = append(capturedLines, message)
		mu.Unlock()
	})
	defer ff.AVUtil_log_set_callback(nil)

	// Create options with list_devices=true
	options := ff.AVUtil_dict_alloc()
	if options == nil {
		return devices
	}
	defer ff.AVUtil_dict_free(options)

	ff.AVUtil_dict_set(options, "list_devices", "true", 0)

	// Try to open the format (this will trigger device enumeration)
	// We expect this to fail, but it will output the device list
	// Use empty string which should trigger the list
	ctx, _ := ff.AVFormat_open_device(format, options)
	if ctx != nil {
		// Try to find stream info which might trigger more logging
		ff.AVFormat_find_stream_info(ctx, nil)
		ff.AVFormat_close_input(ctx)
	}

	// Parse the captured log output
	mu.Lock()
	defer mu.Unlock()

	var currentType string
	// Track if we've seen device index 0 for each type
	defaultSeen := make(map[string]bool)

	for _, line := range capturedLines {
		line = strings.TrimSpace(line)

		// Detect device type headers
		if strings.Contains(line, "video devices:") {
			currentType = "video"
			continue
		} else if strings.Contains(line, "audio devices:") {
			currentType = "audio"
			continue
		}

		// Parse device entries: [0] Device Name
		if matches := avfoundationDevicePattern.FindStringSubmatch(line); len(matches) == 3 {
			deviceName := strings.TrimSpace(matches[2])
			if deviceName == "" {
				continue
			}

			// Parse device index
			deviceIndex := 0
			if idx := matches[1]; idx != "" {
				if n, err := strconv.Atoi(idx); err == nil {
					deviceIndex = n
				}
			}

			// Determine if this is the default device for its type
			// Only the first device (index 0) of each type is default
			isDefault := false
			if deviceIndex == 0 && currentType != "" && !defaultSeen[currentType] {
				isDefault = true
				defaultSeen[currentType] = true
			}

			// Create device entry
			device := schema.Device{
				Index:       deviceIndex,
				Name:        deviceName,
				Description: deviceName,
				IsDefault:   isDefault,
			}

			// Add media type if we captured it
			if currentType != "" {
				device.MediaTypes = []string{currentType}
			}

			devices = append(devices, device)
		}
	}

	return devices
}
