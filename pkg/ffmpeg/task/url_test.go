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
