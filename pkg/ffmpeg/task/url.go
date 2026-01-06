package task

import (
	"fmt"
	"net/url"
	"strings"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// ParsedURL represents a parsed media URL with device/format information
type ParsedURL struct {
	Scheme    string            // URL scheme (file, http, https, rtmp, device, etc.)
	Path      string            // File path, URL, or device identifier
	Format    string            // Format name (for device:// URLs)
	Options   map[string]string // Query parameters as device options
	IsDevice  bool              // True if this is a device URL
	IsNetwork bool              // True if this is a network stream
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// ParseMediaURL parses a media URL and extracts device/format information
// Supports:
//   - Regular file paths: /path/to/file.mp4, ./file.mp4, file.mp4
//   - Network streams: http://..., https://..., rtmp://...
//   - Device URLs: device://format/device-path?options
//
// Examples:
//   - device://avfoundation/0:0
//   - device://avfoundation/Logitech%20StreamCam:ZoomAudioDevice?framerate=30
//   - device://v4l2//dev/video0
//   - device://alsa/hw:0
func ParseMediaURL(input string) (*ParsedURL, error) {
	result := &ParsedURL{
		Options: make(map[string]string),
	}

	// Try to parse as URL
	u, err := url.Parse(input)
	if err != nil || u.Scheme == "" || u.Scheme == "file" {
		// Treat as regular file path
		result.Scheme = "file"
		result.Path = input
		return result, nil
	}

	result.Scheme = u.Scheme

	// Check if it's a device URL
	if u.Scheme == "device" {
		result.IsDevice = true
		result.Format = u.Host

		// Path contains the device identifier
		// For device URLs like device://avfoundation/0:0, u.Path will be "/0:0"
		// Use EscapedPath() to preserve URL encoding if present
		path := u.EscapedPath()
		if path == "" && u.RawPath != "" {
			path = u.RawPath
		} else if path == "" {
			path = u.Path
		}
		result.Path = strings.TrimPrefix(path, "/")

		// Parse query parameters as device options
		for key, values := range u.Query() {
			if len(values) > 0 {
				result.Options[key] = values[0]
			}
		}

		// Validate format is specified
		if result.Format == "" {
			return nil, fmt.Errorf("device URL missing format: %s", input)
		}

		return result, nil
	}

	// Network streams (http, https, rtmp, rtsp, etc.)
	result.IsNetwork = true
	result.Path = input // Use full URL as path
	return result, nil
}

// OpenReader creates a Reader from a parsed media URL
func OpenReader(parsed *ParsedURL) (*ffmpeg.Reader, error) {
	var opts []ffmpeg.Opt

	// For device URLs, specify the format and options
	if parsed.IsDevice {
		// Build options string from query parameters
		var optStrings []string
		for key, value := range parsed.Options {
			optStrings = append(optStrings, fmt.Sprintf("%s=%s", key, value))
		}

		// For device formats, we need to find the input format via device iteration
		// since AVFormat_find_input_format may not work for device-only formats.
		//
		// Search precedence: video devices are checked first, then audio devices.
		// If a format name exists in both video and audio device lists (which is rare),
		// the video device format will be used. This is intentional as most device formats
		// are specific to either video (v4l2, avfoundation) or audio (alsa, pulse).
		var inputFormat *ff.AVInputFormat

		// Try video devices first
		for d := ff.AVDevice_input_video_device_first(); d != nil; d = ff.AVDevice_input_video_device_next(d) {
			if d.Name() == parsed.Format {
				inputFormat = d
				break
			}
		}

		// Try audio devices if not found
		if inputFormat == nil {
			for d := ff.AVDevice_input_audio_device_first(); d != nil; d = ff.AVDevice_input_audio_device_next(d) {
				if d.Name() == parsed.Format {
					inputFormat = d
					break
				}
			}
		}

		if inputFormat == nil {
			return nil, fmt.Errorf("device format not found: %s", parsed.Format)
		}

		// Add format and options using the format we found
		if len(optStrings) > 0 {
			opts = append(opts, ffmpeg.WithInputFormat(inputFormat, optStrings...))
		} else {
			opts = append(opts, ffmpeg.WithInputFormat(inputFormat))
		}
	}

	// Open the reader
	return ffmpeg.Open(parsed.Path, opts...)
}

// OpenReaderFromURL is a convenience function that parses and opens in one call
func OpenReaderFromURL(input string) (*ffmpeg.Reader, error) {
	parsed, err := ParseMediaURL(input)
	if err != nil {
		return nil, err
	}
	return OpenReader(parsed)
}
