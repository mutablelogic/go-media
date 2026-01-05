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

func TestListAudioChannelLayout_All(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListAudioChannelLayout(context.Background(), &schema.ListAudioChannelLayoutRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	t.Logf("Found %d channel layouts", len(response))

	// Verify each layout has valid data
	for _, layout := range response {
		assert.NotEmpty(t, layout.Name)
		assert.Greater(t, layout.NumChannels, 0)
		assert.NotEmpty(t, layout.Order)
		assert.Len(t, layout.Channels, layout.NumChannels)
	}
}

func TestListAudioChannelLayout_FilterByName(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	tests := []struct {
		name        string
		numChannels int
	}{
		{"mono", 1},
		{"stereo", 2},
		{"5.1", 6},
		{"7.1", 8},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := m.ListAudioChannelLayout(context.Background(), &schema.ListAudioChannelLayoutRequest{
				Name: tc.name,
			})
			require.NoError(t, err)
			require.Len(t, response, 1)
			assert.Equal(t, tc.name, response[0].Name)
			assert.Equal(t, tc.numChannels, response[0].NumChannels)
			t.Logf("%s: %d channels, order=%s", response[0].Name, response[0].NumChannels, response[0].Order)
		})
	}
}

func TestListAudioChannelLayout_FilterByNumChannels(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	tests := []struct {
		numChannels   int
		minExpected   int
		expectedNames []string // some expected layout names
	}{
		{1, 1, []string{"mono"}},
		{2, 1, []string{"stereo"}},
		{6, 1, []string{"5.1"}},
		{8, 1, []string{"7.1"}},
	}

	for _, tc := range tests {
		t.Run(string(rune('0'+tc.numChannels))+"ch", func(t *testing.T) {
			response, err := m.ListAudioChannelLayout(context.Background(), &schema.ListAudioChannelLayoutRequest{
				NumChannels: tc.numChannels,
			})
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(response), tc.minExpected)

			// All returned layouts should have the requested number of channels
			for _, layout := range response {
				assert.Equal(t, tc.numChannels, layout.NumChannels)
			}

			// Check that expected names are present
			names := make(map[string]bool)
			for _, layout := range response {
				names[layout.Name] = true
			}
			for _, expected := range tc.expectedNames {
				assert.True(t, names[expected], "expected layout %q not found", expected)
			}

			t.Logf("Found %d layouts with %d channels", len(response), tc.numChannels)
		})
	}
}

func TestListAudioChannelLayout_FilterByNameAndNumChannels(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Filter by both name and numChannels (should match)
	response, err := m.ListAudioChannelLayout(context.Background(), &schema.ListAudioChannelLayoutRequest{
		Name:        "stereo",
		NumChannels: 2,
	})
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, "stereo", response[0].Name)
	assert.Equal(t, 2, response[0].NumChannels)
}

func TestListAudioChannelLayout_FilterNoMatch(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Non-existent name
	response, err := m.ListAudioChannelLayout(context.Background(), &schema.ListAudioChannelLayoutRequest{
		Name: "nonexistent_layout",
	})
	require.NoError(t, err)
	assert.Empty(t, response)

	// Mismatched name and numChannels
	response, err = m.ListAudioChannelLayout(context.Background(), &schema.ListAudioChannelLayoutRequest{
		Name:        "stereo",
		NumChannels: 6, // stereo has 2 channels, not 6
	})
	require.NoError(t, err)
	assert.Empty(t, response)
}

func TestListAudioChannelLayout_NilRequest(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Nil request should return all layouts
	response, err := m.ListAudioChannelLayout(context.Background(), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, response)
}

func TestListAudioChannelLayout_ChannelDetails(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Get stereo layout and verify channel details
	response, err := m.ListAudioChannelLayout(context.Background(), &schema.ListAudioChannelLayoutRequest{
		Name: "stereo",
	})
	require.NoError(t, err)
	require.Len(t, response, 1)

	layout := response[0]
	assert.Equal(t, 2, len(layout.Channels))

	// Verify channel structure
	for i, ch := range layout.Channels {
		assert.Equal(t, i, ch.Index)
		assert.NotEmpty(t, ch.Name)
		t.Logf("Channel %d: %s (%s)", ch.Index, ch.Name, ch.Description)
	}
}
