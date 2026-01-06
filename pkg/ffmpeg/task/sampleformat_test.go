package task_test

import (
	"context"
	"testing"

	// Packages
	"github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	"github.com/mutablelogic/go-media/pkg/ffmpeg/task"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListSampleFormat_All(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListSampleFormats(context.Background(), &schema.ListSampleFormatRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	t.Logf("Found %d sample formats", len(response))

	// Verify each format has valid data
	for _, sf := range response {
		name := ff.AVUtil_get_sample_fmt_name(sf.AVSampleFormat)
		bytesPerSample := ff.AVUtil_get_bytes_per_sample(sf.AVSampleFormat)
		assert.NotEmpty(t, name)
		assert.Greater(t, bytesPerSample, 0)
		assert.Greater(t, bytesPerSample*8, 0)
		assert.Equal(t, bytesPerSample*8, bytesPerSample*8)
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
			response, err := m.ListSampleFormats(context.Background(), &schema.ListSampleFormatRequest{
				Name: tc.name,
			})
			require.NoError(t, err)
			require.Len(t, response, 1)
			name := ff.AVUtil_get_sample_fmt_name(response[0].AVSampleFormat)
			bytesPerSample := ff.AVUtil_get_bytes_per_sample(response[0].AVSampleFormat)
			isPlanar := ff.AVUtil_sample_fmt_is_planar(response[0].AVSampleFormat)
			packedFmt := ff.AVUtil_get_packed_sample_fmt(response[0].AVSampleFormat)
			planarFmt := ff.AVUtil_get_planar_sample_fmt(response[0].AVSampleFormat)
			assert.Equal(t, tc.name, name)
			assert.Equal(t, tc.bitsPerSample, bytesPerSample*8)
			assert.Equal(t, tc.isPlanar, isPlanar)
			t.Logf("%s: %d bits, planar=%v, packed=%s, planar=%s",
				name, bytesPerSample*8, isPlanar,
				ff.AVUtil_get_sample_fmt_name(packedFmt), ff.AVUtil_get_sample_fmt_name(planarFmt))
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
	response, err := m.ListSampleFormats(context.Background(), &schema.ListSampleFormatRequest{
		IsPlanar: &isPlanar,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, sf := range response {
		assert.True(t, ff.AVUtil_sample_fmt_is_planar(sf.AVSampleFormat), "expected planar format, got %s", ff.AVUtil_get_sample_fmt_name(sf.AVSampleFormat))
	}
	t.Logf("Found %d planar sample formats", len(response))

	// Filter packed formats
	isPacked := false
	response, err = m.ListSampleFormats(context.Background(), &schema.ListSampleFormatRequest{
		IsPlanar: &isPacked,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, response)

	for _, sf := range response {
		assert.False(t, ff.AVUtil_sample_fmt_is_planar(sf.AVSampleFormat), "expected packed format, got %s", ff.AVUtil_get_sample_fmt_name(sf.AVSampleFormat))
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
	response, err := m.ListSampleFormats(context.Background(), &schema.ListSampleFormatRequest{
		Name:     "fltp",
		IsPlanar: &isPlanar,
	})
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, "fltp", ff.AVUtil_get_sample_fmt_name(response[0].AVSampleFormat))
	assert.True(t, ff.AVUtil_sample_fmt_is_planar(response[0].AVSampleFormat))

	// Mismatched filter (should not match)
	isPacked := false
	response, err = m.ListSampleFormats(context.Background(), &schema.ListSampleFormatRequest{
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
	response, err := m.ListSampleFormats(context.Background(), &schema.ListSampleFormatRequest{
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
	response, err := m.ListSampleFormats(context.Background(), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, response)
}

func TestListSampleFormat_PackedPlanarEquivalents(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListSampleFormats(context.Background(), &schema.ListSampleFormatRequest{})
	require.NoError(t, err)

	// Check that packed/planar equivalents are set
	for _, sf := range response {
		packedFmt := ff.AVUtil_get_packed_sample_fmt(sf.AVSampleFormat)
		planarFmt := ff.AVUtil_get_planar_sample_fmt(sf.AVSampleFormat)
		packedName := ff.AVUtil_get_sample_fmt_name(packedFmt)
		planarName := ff.AVUtil_get_sample_fmt_name(planarFmt)
		sfName := ff.AVUtil_get_sample_fmt_name(sf.AVSampleFormat)
		isPlanar := ff.AVUtil_sample_fmt_is_planar(sf.AVSampleFormat)
		assert.NotEmpty(t, packedName)
		assert.NotEmpty(t, planarName)
		if isPlanar {
			assert.NotEqual(t, sfName, packedName, "planar format should have different packed name")
			assert.Equal(t, sfName, planarName, "planar format should be its own planar equivalent")
		} else {
			assert.Equal(t, sfName, packedName, "packed format should be its own packed equivalent")
			assert.NotEqual(t, sfName, planarName, "packed format should have different planar name")
		}
	}
}

func TestListSampleFormat_BitDepths(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListSampleFormats(context.Background(), &schema.ListSampleFormatRequest{})
	require.NoError(t, err)

	// Group by bit depth
	bitDepths := make(map[int]int)
	for _, sf := range response {
		bytesPerSample := ff.AVUtil_get_bytes_per_sample(sf.AVSampleFormat)
		bitDepths[bytesPerSample*8]++
	}

	t.Logf("Sample format bit depths:")
	for bits, count := range bitDepths {
		t.Logf("  %d-bit: %d formats", bits, count)
	}
}
