package task_test

import (
	"context"
	"testing"

	// Packages
	"github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	"github.com/mutablelogic/go-media/pkg/ffmpeg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListFormat_All(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	t.Logf("Found %d formats", len(response))

	// Count by type
	inputs := 0
	outputs := 0
	devices := 0
	for _, f := range response {
		if f.IsInput {
			inputs++
		}
		if f.IsOutput {
			outputs++
		}
		if f.IsDevice {
			devices++
		}
	}
	t.Logf("Inputs: %d, Outputs: %d, Devices: %d", inputs, outputs, devices)
}

func TestListFormat_FilterByName(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	tests := []struct {
		name string
	}{
		{"mp4"},
		{"mp3"},
		{"wav"},
		{"avi"},
		{"mov"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
				Name: tc.name,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, response, "expected to find format matching %q", tc.name)

			for _, f := range response {
				assert.Contains(t, f.Name, tc.name)
				t.Logf("%s: %s (input=%v, output=%v, device=%v)", f.Name, f.Description, f.IsInput, f.IsOutput, f.IsDevice)
			}
		})
	}
}

func TestListFormat_FilterByInput(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	isInput := true
	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
		IsInput: &isInput,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, f := range response {
		assert.True(t, f.IsInput, "format %s should be input", f.Name)
	}
	t.Logf("Found %d input formats", len(response))
}

func TestListFormat_FilterByOutput(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	isOutput := true
	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
		IsOutput: &isOutput,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, f := range response {
		assert.True(t, f.IsOutput, "format %s should be output", f.Name)
	}
	t.Logf("Found %d output formats", len(response))
}

func TestListFormat_FilterByDevice(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	isDevice := true
	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
		IsDevice: &isDevice,
	})
	require.NoError(t, err)
	// Devices may or may not be available depending on the system
	t.Logf("Found %d device formats", len(response))

	for _, f := range response {
		assert.True(t, f.IsDevice, "format %s should be device", f.Name)
		t.Logf("Device: %s (%s, input=%v, output=%v)", f.Name, f.Description, f.IsInput, f.IsOutput)
		if len(f.Devices) > 0 {
			for _, d := range f.Devices {
				t.Logf("  - %s: %s (default=%v)", d.Name, d.Description, d.IsDefault)
			}
		}
	}
}

func TestListFormat_FilterByInputAndNotDevice(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	isInput := true
	isDevice := false
	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
		IsInput:  &isInput,
		IsDevice: &isDevice,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, f := range response {
		assert.True(t, f.IsInput, "format %s should be input", f.Name)
		assert.False(t, f.IsDevice, "format %s should not be device", f.Name)
	}
	t.Logf("Found %d non-device input formats (demuxers)", len(response))
}

func TestListFormat_FilterByOutputAndNotDevice(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	isOutput := true
	isDevice := false
	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
		IsOutput: &isOutput,
		IsDevice: &isDevice,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, f := range response {
		assert.True(t, f.IsOutput, "format %s should be output", f.Name)
		assert.False(t, f.IsDevice, "format %s should not be device", f.Name)
	}
	t.Logf("Found %d non-device output formats (muxers)", len(response))
}

func TestListFormat_FilterNoMatch(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
		Name: "nonexistent_format_xyz",
	})
	require.NoError(t, err)
	assert.Empty(t, response)
}

func TestListFormat_NilRequest(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListFormat(context.Background(), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	t.Logf("Found %d formats with nil request", len(response))
}

func TestListFormat_FormatDetails(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Check some common formats for expected properties
	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
		Name: "mp4",
	})
	require.NoError(t, err)

	for _, f := range response {
		if f.Name == "mp4" && f.IsOutput {
			t.Logf("MP4 muxer:")
			t.Logf("  Description: %s", f.Description)
			t.Logf("  Extensions: %v", f.Extensions)
			t.Logf("  MIME Types: %v", f.MimeTypes)
			t.Logf("  Default Video Codec: %s", f.DefaultVideoCodec)
			t.Logf("  Default Audio Codec: %s", f.DefaultAudioCodec)
			t.Logf("  Flags: %v", f.Flags)
		}
	}
}

func TestListFormat_ExtensionsAndMimeTypes(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	isOutput := true
	isDevice := false
	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
		IsOutput: &isOutput,
		IsDevice: &isDevice,
	})
	require.NoError(t, err)

	withExtensions := 0
	withMimeTypes := 0
	for _, f := range response {
		if len(f.Extensions) > 0 {
			withExtensions++
		}
		if len(f.MimeTypes) > 0 {
			withMimeTypes++
		}
	}
	t.Logf("%d formats have extensions, %d have MIME types", withExtensions, withMimeTypes)
}

func TestListFormat_DefaultCodecs(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	isOutput := true
	isDevice := false
	response, err := m.ListFormat(context.Background(), &schema.ListFormatRequest{
		IsOutput: &isOutput,
		IsDevice: &isDevice,
	})
	require.NoError(t, err)

	withVideoCodec := 0
	withAudioCodec := 0
	withSubtitleCodec := 0
	for _, f := range response {
		if f.DefaultVideoCodec != "" {
			withVideoCodec++
		}
		if f.DefaultAudioCodec != "" {
			withAudioCodec++
		}
		if f.DefaultSubtitleCodec != "" {
			withSubtitleCodec++
		}
	}
	t.Logf("Output formats with default codecs: video=%d, audio=%d, subtitle=%d",
		withVideoCodec, withAudioCodec, withSubtitleCodec)
}
