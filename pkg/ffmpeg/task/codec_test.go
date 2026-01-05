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

func TestListCodec_All(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	t.Logf("Found %d codecs", len(response))

	// Verify each codec has valid data
	for _, c := range response {
		assert.NotEmpty(t, c.Name)
		assert.NotEmpty(t, c.Type)
		assert.NotEmpty(t, c.ID)
		// A codec should be either encoder or decoder (or both)
		assert.True(t, c.IsEncoder || c.IsDecoder, "codec %s should be encoder or decoder", c.Name)
	}
}

func TestListCodec_FilterByName(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	tests := []struct {
		name string
	}{
		{"h264"},
		{"aac"},
		{"mp3"},
		{"png"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{
				Name: tc.name,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, response, "expected to find codec matching %q", tc.name)

			for _, c := range response {
				assert.Contains(t, c.Name, tc.name)
				t.Logf("%s: %s (%s, encoder=%v, decoder=%v)", c.Name, c.LongName, c.Type, c.IsEncoder, c.IsDecoder)
			}
		})
	}
}

func TestListCodec_FilterByType(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	tests := []struct {
		mediaType     string
		expectedCodec string // At least one codec we expect to find
	}{
		{"video", "h264"},
		{"audio", "aac"},
		{"subtitle", "ass"},
	}

	for _, tc := range tests {
		t.Run(tc.mediaType, func(t *testing.T) {
			response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{
				Type: tc.mediaType,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, response)

			// All returned codecs should have the correct type
			for _, c := range response {
				assert.Equal(t, tc.mediaType, c.Type)
			}

			// Check if expected codec is in the list
			found := false
			for _, c := range response {
				if c.Name == tc.expectedCodec || c.ID == tc.expectedCodec {
					found = true
					break
				}
			}
			assert.True(t, found, "expected to find %s codec in %s type", tc.expectedCodec, tc.mediaType)

			t.Logf("Found %d %s codecs", len(response), tc.mediaType)
		})
	}
}

func TestListCodec_FilterByEncoder(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Get encoders only
	isEncoder := true
	response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{
		IsEncoder: &isEncoder,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, c := range response {
		assert.True(t, c.IsEncoder, "codec %s should be an encoder", c.Name)
	}
	t.Logf("Found %d encoders", len(response))
}

func TestListCodec_FilterByDecoder(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Get decoders only
	isEncoder := false
	response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{
		IsEncoder: &isEncoder,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, c := range response {
		assert.True(t, c.IsDecoder, "codec %s should be a decoder", c.Name)
	}
	t.Logf("Found %d decoders", len(response))
}

func TestListCodec_FilterByTypeAndEncoder(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	isEncoder := true
	response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{
		Type:      "video",
		IsEncoder: &isEncoder,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, c := range response {
		assert.Equal(t, "video", c.Type)
		assert.True(t, c.IsEncoder, "codec %s should be an encoder", c.Name)
	}
	t.Logf("Found %d video encoders", len(response))
}

func TestListCodec_FilterNoMatch(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{
		Name: "nonexistent_codec_xyz",
	})
	require.NoError(t, err)
	assert.Empty(t, response)
}

func TestListCodec_NilRequest(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListCodec(context.Background(), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	t.Logf("Found %d codecs with nil request", len(response))
}

func TestListCodec_VideoCodecFormats(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Get a specific video encoder and check its pixel formats
	response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{
		Name: "libx264",
	})
	require.NoError(t, err)

	for _, c := range response {
		if c.IsEncoder && c.Type == "video" {
			t.Logf("Codec %s supports pixel formats: %v", c.Name, c.PixelFormats)
			if len(c.PixelFormats) > 0 {
				assert.NotEmpty(t, c.PixelFormats[0])
			}
		}
	}
}

func TestListCodec_AudioCodecFormats(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Get audio encoders and check their sample formats
	isEncoder := true
	response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{
		Type:      "audio",
		IsEncoder: &isEncoder,
	})
	require.NoError(t, err)

	codecsWithFormats := 0
	for _, c := range response {
		if len(c.SampleFormats) > 0 {
			codecsWithFormats++
			t.Logf("Codec %s supports sample formats: %v", c.Name, c.SampleFormats)
		}
	}
	t.Logf("%d audio encoders have sample format info", codecsWithFormats)
}

func TestListCodec_Capabilities(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{})
	require.NoError(t, err)

	// Count codecs with various capabilities
	hwAccel := 0
	experimental := 0
	withCaps := 0

	for _, c := range response {
		if c.IsHardware {
			hwAccel++
		}
		if c.IsExperiment {
			experimental++
		}
		if len(c.Capabilities) > 0 {
			withCaps++
		}
	}

	t.Logf("Hardware accelerated: %d", hwAccel)
	t.Logf("Experimental: %d", experimental)
	t.Logf("With capabilities: %d", withCaps)
}

func TestListCodec_Profiles(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Find codecs with profiles (like h264)
	response, err := m.ListCodec(context.Background(), &schema.ListCodecRequest{})
	require.NoError(t, err)

	codecsWithProfiles := 0
	for _, c := range response {
		if len(c.Profiles) > 0 {
			codecsWithProfiles++
			if c.Name == "h264" || c.Name == "libx264" {
				t.Logf("Codec %s profiles: %v", c.Name, c.Profiles)
			}
		}
	}
	t.Logf("%d codecs have profile info", codecsWithProfiles)
}
