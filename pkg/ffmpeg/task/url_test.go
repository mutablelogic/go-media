package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMediaURL_FilePaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ParsedURL
	}{
		{
			name:  "absolute path",
			input: "/path/to/file.mp4",
			expected: &ParsedURL{
				Scheme:  "file",
				Path:    "/path/to/file.mp4",
				Options: map[string]string{},
			},
		},
		{
			name:  "relative path",
			input: "./relative/file.mp4",
			expected: &ParsedURL{
				Scheme:  "file",
				Path:    "./relative/file.mp4",
				Options: map[string]string{},
			},
		},
		{
			name:  "simple filename",
			input: "file.mp4",
			expected: &ParsedURL{
				Scheme:  "file",
				Path:    "file.mp4",
				Options: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseMediaURL(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Scheme, parsed.Scheme)
			assert.Equal(t, tt.expected.Path, parsed.Path)
			assert.False(t, parsed.IsDevice)
			assert.False(t, parsed.IsNetwork)
		})
	}
}

func TestParseMediaURL_DeviceURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ParsedURL
	}{
		{
			name:  "avfoundation with indices",
			input: "device://avfoundation/0:0",
			expected: &ParsedURL{
				Scheme:   "device",
				Format:   "avfoundation",
				Path:     "0:0",
				IsDevice: true,
				Options:  map[string]string{},
			},
		},
		{
			name:  "avfoundation with options",
			input: "device://avfoundation/0:0?framerate=30&video_size=1920x1080",
			expected: &ParsedURL{
				Scheme:   "device",
				Format:   "avfoundation",
				Path:     "0:0",
				IsDevice: true,
				Options: map[string]string{
					"framerate":  "30",
					"video_size": "1920x1080",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseMediaURL(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Scheme, parsed.Scheme)
			assert.Equal(t, tt.expected.Format, parsed.Format)
			assert.Equal(t, tt.expected.Path, parsed.Path)
			assert.True(t, parsed.IsDevice)
			assert.False(t, parsed.IsNetwork)
		})
	}
}

func TestParseMediaURL_InvalidDeviceURL(t *testing.T) {
	_, err := ParseMediaURL("device:///0:0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing format")
}

func TestParseMediaURL_NetworkURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ParsedURL
	}{
		{
			name:  "http URL",
			input: "http://example.com/stream.m3u8",
			expected: &ParsedURL{
				Scheme:    "http",
				Path:      "http://example.com/stream.m3u8",
				IsNetwork: true,
				Options:   map[string]string{},
			},
		},
		{
			name:  "https URL",
			input: "https://example.com/video.mp4",
			expected: &ParsedURL{
				Scheme:    "https",
				Path:      "https://example.com/video.mp4",
				IsNetwork: true,
				Options:   map[string]string{},
			},
		},
		{
			name:  "rtmp URL",
			input: "rtmp://server/live/stream",
			expected: &ParsedURL{
				Scheme:    "rtmp",
				Path:      "rtmp://server/live/stream",
				IsNetwork: true,
				Options:   map[string]string{},
			},
		},
		{
			name:  "rtsp URL",
			input: "rtsp://camera/stream",
			expected: &ParsedURL{
				Scheme:    "rtsp",
				Path:      "rtsp://camera/stream",
				IsNetwork: true,
				Options:   map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseMediaURL(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Scheme, parsed.Scheme)
			assert.Equal(t, tt.expected.Path, parsed.Path)
			assert.False(t, parsed.IsDevice)
			assert.True(t, parsed.IsNetwork)
		})
	}
}

func TestParseMediaURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
		expected    *ParsedURL
	}{
		{
			name:        "empty string",
			input:       "",
			shouldError: false,
			expected: &ParsedURL{
				Scheme:  "file",
				Path:    "",
				Options: map[string]string{},
			},
		},
		{
			name:        "device URL with missing format",
			input:       "device:///path",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseMediaURL(tt.input)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Scheme, parsed.Scheme)
				assert.Equal(t, tt.expected.Path, parsed.Path)
			}
		})
	}
}

func TestParseMediaURL_URLEncodedDevicePaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ParsedURL
	}{
		{
			name:  "URL-encoded device name with spaces",
			input: "device://avfoundation/Logitech%20StreamCam:ZoomAudioDevice",
			expected: &ParsedURL{
				Scheme:   "device",
				Format:   "avfoundation",
				Path:     "Logitech%20StreamCam:ZoomAudioDevice",
				IsDevice: true,
				Options:  map[string]string{},
			},
		},
		{
			name:  "URL-encoded device with options",
			input: "device://avfoundation/Logitech%20StreamCam:ZoomAudioDevice?framerate=30",
			expected: &ParsedURL{
				Scheme:   "device",
				Format:   "avfoundation",
				Path:     "Logitech%20StreamCam:ZoomAudioDevice",
				IsDevice: true,
				Options: map[string]string{
					"framerate": "30",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseMediaURL(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Scheme, parsed.Scheme)
			assert.Equal(t, tt.expected.Format, parsed.Format)
			assert.Equal(t, tt.expected.Path, parsed.Path)
			assert.True(t, parsed.IsDevice)
			assert.False(t, parsed.IsNetwork)
			assert.Equal(t, tt.expected.Options, parsed.Options)
		})
	}
}

func TestParseMediaURL_DeviceEmptyPaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ParsedURL
	}{
		{
			name:  "device with trailing slash",
			input: "device://alsa/",
			expected: &ParsedURL{
				Scheme:   "device",
				Format:   "alsa",
				Path:     "",
				IsDevice: true,
				Options:  map[string]string{},
			},
		},
		{
			name:  "pulse with empty path",
			input: "device://pulse/",
			expected: &ParsedURL{
				Scheme:   "device",
				Format:   "pulse",
				Path:     "",
				IsDevice: true,
				Options:  map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseMediaURL(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Scheme, parsed.Scheme)
			assert.Equal(t, tt.expected.Format, parsed.Format)
			assert.Equal(t, tt.expected.Path, parsed.Path)
			assert.True(t, parsed.IsDevice)
			assert.False(t, parsed.IsNetwork)
		})
	}
}

func TestParseMediaURL_MultipleQueryParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *ParsedURL
	}{
		{
			name:  "multiple options with different types",
			input: "device://avfoundation/0:0?framerate=30&video_size=1920x1080&pixel_format=uyvy422",
			expected: &ParsedURL{
				Scheme:   "device",
				Format:   "avfoundation",
				Path:     "0:0",
				IsDevice: true,
				Options: map[string]string{
					"framerate":    "30",
					"video_size":   "1920x1080",
					"pixel_format": "uyvy422",
				},
			},
		},
		{
			name:  "v4l2 with multiple options",
			input: "device://v4l2//dev/video0?framerate=30&video_size=640x480&pixel_format=mjpeg",
			expected: &ParsedURL{
				Scheme:   "device",
				Format:   "v4l2",
				Path:     "/dev/video0",
				IsDevice: true,
				Options: map[string]string{
					"framerate":    "30",
					"video_size":   "640x480",
					"pixel_format": "mjpeg",
				},
			},
		},
		{
			name:  "alsa with audio buffer size",
			input: "device://alsa/hw:0?audio_buffer_size=4096&channels=2",
			expected: &ParsedURL{
				Scheme:   "device",
				Format:   "alsa",
				Path:     "hw:0",
				IsDevice: true,
				Options: map[string]string{
					"audio_buffer_size": "4096",
					"channels":          "2",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseMediaURL(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.Scheme, parsed.Scheme)
			assert.Equal(t, tt.expected.Format, parsed.Format)
			assert.Equal(t, tt.expected.Path, parsed.Path)
			assert.True(t, parsed.IsDevice)
			assert.False(t, parsed.IsNetwork)
			assert.Equal(t, tt.expected.Options, parsed.Options)
		})
	}
}
