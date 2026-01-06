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

func TestListPixelFormat_All(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListPixelFormats(context.Background(), &schema.ListPixelFormatRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, response)
	t.Logf("Found %d pixel formats", len(response))

	// Verify each format has valid data
	// Note: Hardware-accelerated formats may have 0 for NumComponents and BitsPerPixel
	for _, pf := range response {
		assert.NotEmpty(t, pf.Name)
		if !pf.IsHWAccel {
			// Only software formats should have these properties
			assert.Greater(t, pf.NumComponents, 0, "format %s should have components", pf.Name)
			assert.Greater(t, pf.BitsPerPixel, 0, "format %s should have bits per pixel", pf.Name)
		}
	}
}

func TestListPixelFormat_FilterByName(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	tests := []struct {
		name         string
		bitsPerPixel int
		numPlanes    int
	}{
		{"yuv420p", 12, 3},
		{"rgb24", 24, 1},
		{"rgba", 32, 1},
		{"nv12", 12, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response, err := m.ListPixelFormats(context.Background(), &schema.ListPixelFormatRequest{
				Name: tc.name,
			})
			require.NoError(t, err)
			require.Len(t, response, 1)
			assert.Equal(t, tc.name, response[0].Name)
			assert.Equal(t, tc.bitsPerPixel, response[0].BitsPerPixel)
			assert.Equal(t, tc.numPlanes, response[0].NumPlanes)
			t.Logf("%s: %d bpp, %d planes, planar=%v, rgb=%v, alpha=%v",
				response[0].Name, response[0].BitsPerPixel, response[0].NumPlanes,
				response[0].IsPlanar, response[0].IsRGB, response[0].HasAlpha)
		})
	}
}

func TestListPixelFormat_FilterByNumPlanes(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	tests := []struct {
		numPlanes     int
		minExpected   int
		expectedNames []string
	}{
		{1, 1, []string{"rgb24", "rgba"}},
		{2, 1, []string{"nv12"}},
		{3, 1, []string{"yuv420p", "yuv422p"}},
	}

	for _, tc := range tests {
		t.Run(string(rune('0'+tc.numPlanes))+"planes", func(t *testing.T) {
			response, err := m.ListPixelFormats(context.Background(), &schema.ListPixelFormatRequest{
				NumPlanes: tc.numPlanes,
			})
			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(response), tc.minExpected)

			// All returned formats should have the requested number of planes
			for _, pf := range response {
				assert.Equal(t, tc.numPlanes, pf.NumPlanes)
			}

			// Check that expected names are present
			names := make(map[string]bool)
			for _, pf := range response {
				names[pf.Name] = true
			}
			for _, expected := range tc.expectedNames {
				assert.True(t, names[expected], "expected format %q not found", expected)
			}

			t.Logf("Found %d formats with %d planes", len(response), tc.numPlanes)
		})
	}
}

func TestListPixelFormat_FilterByNameAndNumPlanes(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Filter by both name and numPlanes (should match)
	response, err := m.ListPixelFormats(context.Background(), &schema.ListPixelFormatRequest{
		Name:      "yuv420p",
		NumPlanes: 3,
	})
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, "yuv420p", response[0].Name)
	assert.Equal(t, 3, response[0].NumPlanes)
}

func TestListPixelFormat_FilterNoMatch(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Non-existent name
	response, err := m.ListPixelFormats(context.Background(), &schema.ListPixelFormatRequest{
		Name: "nonexistent_format",
	})
	require.NoError(t, err)
	assert.Empty(t, response)

	// Mismatched name and numPlanes
	response, err = m.ListPixelFormats(context.Background(), &schema.ListPixelFormatRequest{
		Name:      "yuv420p",
		NumPlanes: 1, // yuv420p has 3 planes, not 1
	})
	require.NoError(t, err)
	assert.Empty(t, response)
}

func TestListPixelFormat_NilRequest(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	// Nil request should return all formats
	response, err := m.ListPixelFormats(context.Background(), nil)
	require.NoError(t, err)
	assert.NotEmpty(t, response)
}

func TestListPixelFormat_RGBFormats(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListPixelFormats(context.Background(), &schema.ListPixelFormatRequest{})
	require.NoError(t, err)

	// Filter RGB formats
	var rgbCount int
	for _, pf := range response {
		if pf.IsRGB {
			rgbCount++
		}
	}
	assert.Greater(t, rgbCount, 0)
	t.Logf("Found %d RGB formats out of %d total", rgbCount, len(response))
}

func TestListPixelFormat_AlphaFormats(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListPixelFormats(context.Background(), &schema.ListPixelFormatRequest{})
	require.NoError(t, err)

	// Filter formats with alpha
	var alphaCount int
	for _, pf := range response {
		if pf.HasAlpha {
			alphaCount++
		}
	}
	assert.Greater(t, alphaCount, 0)
	t.Logf("Found %d formats with alpha out of %d total", alphaCount, len(response))
}

func TestListPixelFormat_PlanarFormats(t *testing.T) {
	m, err := task.NewManager()
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, m)

	response, err := m.ListPixelFormats(context.Background(), &schema.ListPixelFormatRequest{})
	require.NoError(t, err)

	// Filter planar formats
	var planarCount int
	for _, pf := range response {
		if pf.IsPlanar {
			planarCount++
		}
	}
	assert.Greater(t, planarCount, 0)
	t.Logf("Found %d planar formats out of %d total", planarCount, len(response))
}
