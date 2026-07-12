package manager

import (
	"regexp"
	"strconv"
	"strings"
	"sync"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	// Example line: [AVFoundation indev @ 0x...] [0] FaceTime HD Camera
	avfoundationDevicePattern = regexp.MustCompile(`\[(\d+)\]\s+(.+)$`)
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// enumerateAVFoundationDevices opens avfoundation with list_devices=true and
// parses FFmpeg logs to build device entries.
func enumerateAVFoundationDevices(format *ff.AVInputFormat) []schema.Device {
	var devices []schema.Device
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

	ctx, _ := ff.AVFormat_open_device(format, options)
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

		if matches := avfoundationDevicePattern.FindStringSubmatch(line); len(matches) == 3 {
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

			device := schema.Device{
				Index:       deviceIndex,
				Name:        deviceName,
				Description: deviceName,
				IsDefault:   isDefault,
			}
			if currentType != "" {
				device.MediaTypes = []string{currentType}
			}

			devices = append(devices, device)
		}
	}

	return devices
}
