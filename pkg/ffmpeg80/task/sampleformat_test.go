package task_test

import (
	"context"
	"testing"

	// Packages
	"github.com/mutablelogic/go-media/pkg/ffmpeg80/schema"
	"github.com/mutablelogic/go-media/pkg/ffmpeg80/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSampleFormat_All(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListSampleFormat(context.Background(), &schema.ListSampleFormatRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	t.Logf("Found %d sample formats", len(response))

	// Verify each format has valid data
	for _, sf := range response {
		assert.NotEmpty(t, sf.Name)
		assert.Greater(t, sf.BytesPerSample, 0)
		assert.Greater(t, sf.BitsPerSample, 0)
		assert.Equal(t, sf.BitsPerSample, sf.BytesPerSample*8)
	}
}

func TestListSampleFormat_FilterByName(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	tests := []struct {
		name          string
		bitsPerSample int
		isPlanar      bool
	}{
		{"u8", 8, false},
		{"s16", 16, false},
		{"s32", 32, false},
		{"flt", 32, false},
		{"dbl", 64, false},
		{"u8p", 8, true},
		{"s16p", 16, true},
		{"fltp", 32, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := m.ListSampleFormat(context.Background(), &schema.ListSampleFormatRequest{
				Name: tc.name,
			})
			require.NoError(t, err)
			require.Len(t, response, 1)
			assert.Equal(t, tc.name, response[0].Name)
			assert.Equal(t, tc.bitsPerSample, response[0].BitsPerSample)
			assert.Equal(t, tc.isPlanar, response[0].IsPlanar)
			t.Logf("%s: %d bits, planar=%v, packed=%s, planar=%s",
				response[0].Name, response[0].BitsPerSample, response[0].IsPlanar,
				response[0].PackedName, response[0].PlanarName)
		})
	}
}

func TestListSampleFormat_FilterByPlanar(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Filter planar formats
	isPlanar := true
	response, err := m.ListSampleFormat(context.Background(), &schema.ListSampleFormatRequest{
		IsPlanar: &isPlanar,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, sf := range response {
		assert.True(t, sf.IsPlanar, "expected planar format, got %s", sf.Name)
	}
	t.Logf("Found %d planar sample formats", len(response))

	// Filter packed formats
	isPacked := false
	response, err = m.ListSampleFormat(context.Background(), &schema.ListSampleFormatRequest{
		IsPlanar: &isPacked,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, sf := range response {
		assert.False(t, sf.IsPlanar, "expected packed format, got %s", sf.Name)
	}
	t.Logf("Found %d packed sample formats", len(response))
}

func TestListSampleFormat_FilterByNameAndPlanar(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Filter by both name and isPlanar (should match)
	isPlanar := true
	response, err := m.ListSampleFormat(context.Background(), &schema.ListSampleFormatRequest{
		Name:     "fltp",
		IsPlanar: &isPlanar,
	})
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, "fltp", response[0].Name)
	assert.True(t, response[0].IsPlanar)

	// Mismatched filter (should not match)
	isPacked := false
	response, err = m.ListSampleFormat(context.Background(), &schema.ListSampleFormatRequest{
		Name:     "fltp", // planar format
		IsPlanar: &isPacked,
	})
	require.NoError(t, err)
	assert.Empty(t, response)
}

func TestListSampleFormat_FilterNoMatch(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Non-existent name
	response, err := m.ListSampleFormat(context.Background(), &schema.ListSampleFormatRequest{
		Name: "nonexistent_format",
	})
	require.NoError(t, err)
	assert.Empty(t, response)
}

func TestListSampleFormat_NilRequest(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Nil request should return all formats
	response, err := m.ListSampleFormat(context.Background(), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, response)
}

func TestListSampleFormat_PackedPlanarEquivalents(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListSampleFormat(context.Background(), &schema.ListSampleFormatRequest{})
	require.NoError(t, err)

	// Check that packed/planar equivalents are set
	for _, sf := range response {
		assert.NotEmpty(t, sf.PackedName)
		assert.NotEmpty(t, sf.PlanarName)
		if sf.IsPlanar {
			assert.NotEqual(t, sf.Name, sf.PackedName, "planar format should have different packed name")
			assert.Equal(t, sf.Name, sf.PlanarName, "planar format should be its own planar equivalent")
		} else {
			assert.Equal(t, sf.Name, sf.PackedName, "packed format should be its own packed equivalent")
			assert.NotEqual(t, sf.Name, sf.PlanarName, "packed format should have different planar name")
		}
	}
}

func TestListSampleFormat_BitDepths(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListSampleFormat(context.Background(), &schema.ListSampleFormatRequest{})
	require.NoError(t, err)

	// Group by bit depth
	bitDepths := make(map[int]int)
	for _, sf := range response {
		bitDepths[sf.BitsPerSample]++
	}

	t.Logf("Sample format bit depths:")
	for bits, count := range bitDepths {
		t.Logf("  %d-bit: %d formats", bits, count)
	}
}
