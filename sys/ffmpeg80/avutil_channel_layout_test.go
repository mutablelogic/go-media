package ffmpeg

import (
	"encoding/json"
	"sync"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST CHANNEL OPERATIONS

func Test_avutil_channel_name(t *testing.T) {
	assert := assert.New(t)

	// Test some known channels
	tests := []struct {
		channel     AVChannel
		shouldExist bool
	}{
		{0, true},   // Front left
		{1, true},   // Front right
		{2, true},   // Front center
		{-1, false}, // Invalid channel
	}

	for _, tc := range tests {
		name, err := AVUtil_channel_name(tc.channel)
		if tc.shouldExist {
			assert.NoError(err)
			assert.NotEmpty(name)
			t.Logf("Channel %d name: %q", tc.channel, name)
		}
	}
}

func Test_avutil_channel_description(t *testing.T) {
	assert := assert.New(t)

	// Test some known channels
	tests := []struct {
		channel     AVChannel
		shouldExist bool
	}{
		{0, true}, // Front left
		{1, true}, // Front right
		{2, true}, // Front center
	}

	for _, tc := range tests {
		desc, err := AVUtil_channel_description(tc.channel)
		if tc.shouldExist {
			assert.NoError(err)
			assert.NotEmpty(desc)
			t.Logf("Channel %d description: %q", tc.channel, desc)
		}
	}
}

func Test_avutil_channel_from_string(t *testing.T) {
	assert := assert.New(t)

	tests := []string{
		"FL",  // Front Left
		"FR",  // Front Right
		"FC",  // Front Center
		"LFE", // Low Frequency Effects
	}

	for _, name := range tests {
		channel := AVUtil_channel_from_string(name)
		t.Logf("Channel from string %q: %d", name, channel)

		// Verify we can get the name back
		retrievedName, err := AVUtil_channel_name(channel)
		assert.NoError(err)
		assert.NotEmpty(retrievedName)
	}
}

func Test_avutil_channel_from_string_invalid(t *testing.T) {
	// Test with invalid channel name - should not crash
	channel := AVUtil_channel_from_string("INVALID_CHANNEL_NAME")
	t.Logf("Invalid channel returned: %d", channel)
}

////////////////////////////////////////////////////////////////////////////////
// TEST CHANNEL LAYOUT OPERATIONS

func Test_avutil_channel_layout_standard(t *testing.T) {
	assert := assert.New(t)
	var iter uintptr
	count := 0

	for {
		layout := AVUtil_channel_layout_standard(&iter)
		if layout == nil {
			break
		}
		count++

		description, err := AVUtil_channel_layout_describe(layout)
		assert.NoError(err)
		assert.NotEmpty(description)

		channels := AVUtil_get_channel_layout_nb_channels(layout)
		assert.Greater(channels, 0)

		t.Logf("Layout %d: %q (%d channels)", count, description, channels)
	}

	assert.Greater(count, 0, "Should have at least one standard layout")
	t.Logf("Total standard layouts: %d", count)
}

func Test_avutil_channel_layout_describe(t *testing.T) {
	assert := assert.New(t)
	var iter uintptr

	layout := AVUtil_channel_layout_standard(&iter)
	assert.NotNil(layout)

	desc, err := AVUtil_channel_layout_describe(layout)
	assert.NoError(err)
	assert.NotEmpty(desc)
	t.Logf("Layout description: %q", desc)
}

func Test_avutil_channel_layout_default(t *testing.T) {
	assert := assert.New(t)

	tests := []int{1, 2, 4, 6, 8}

	for _, numChannels := range tests {
		var layout AVChannelLayout
		AVUtil_channel_layout_default(&layout, numChannels)
		defer AVUtil_channel_layout_uninit(&layout)

		channels := AVUtil_get_channel_layout_nb_channels(&layout)
		assert.Equal(numChannels, channels)

		desc, err := AVUtil_channel_layout_describe(&layout)
		assert.NoError(err)
		t.Logf("Default layout for %d channels: %q", numChannels, desc)
	}
}

func Test_avutil_channel_layout_from_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		str         string
		shouldWork  bool
		expectChans int
	}{
		{"mono", true, 1},
		{"stereo", true, 2},
		{"5.1", true, 6},
		{"7.1", true, 8},
		{"invalid_layout", false, 0},
	}

	for _, tc := range tests {
		var layout AVChannelLayout
		err := AVUtil_channel_layout_from_string(&layout, tc.str)

		if tc.shouldWork {
			assert.NoError(err, "Failed to parse %q", tc.str)
			defer AVUtil_channel_layout_uninit(&layout)

			channels := AVUtil_get_channel_layout_nb_channels(&layout)
			assert.Equal(tc.expectChans, channels, "Wrong channel count for %q", tc.str)

			desc, descErr := AVUtil_channel_layout_describe(&layout)
			assert.NoError(descErr)
			t.Logf("Layout %q: %q (%d channels)", tc.str, desc, channels)
		} else {
			assert.Error(err, "Should fail for %q", tc.str)
		}
	}
}

func Test_avutil_channel_layout_uninit(t *testing.T) {
	var layout AVChannelLayout
	err := AVUtil_channel_layout_from_string(&layout, "stereo")
	assert.NoError(t, err)

	// Should not crash
	AVUtil_channel_layout_uninit(&layout)

	// Calling again should also not crash
	AVUtil_channel_layout_uninit(&layout)
}

func Test_avutil_channel_layout_channel_from_index(t *testing.T) {
	assert := assert.New(t)

	var layout AVChannelLayout
	err := AVUtil_channel_layout_from_string(&layout, "stereo")
	assert.NoError(err)
	defer AVUtil_channel_layout_uninit(&layout)

	numChannels := AVUtil_get_channel_layout_nb_channels(&layout)
	assert.Equal(2, numChannels)

	// Get channels by index
	for i := 0; i < numChannels; i++ {
		channel := AVUtil_channel_layout_channel_from_index(&layout, i)
		name, nameErr := AVUtil_channel_name(channel)
		assert.NoError(nameErr)
		t.Logf("Channel at index %d: %q (id: %d)", i, name, channel)
	}
}

func Test_avutil_channel_layout_index_from_channel(t *testing.T) {
	assert := assert.New(t)

	var layout AVChannelLayout
	err := AVUtil_channel_layout_from_string(&layout, "stereo")
	assert.NoError(err)
	defer AVUtil_channel_layout_uninit(&layout)

	// Get first channel
	channel := AVUtil_channel_layout_channel_from_index(&layout, 0)

	// Find its index (should be 0)
	index := AVUtil_channel_layout_index_from_channel(&layout, channel)
	assert.Equal(0, index)

	t.Logf("Channel %d found at index %d", channel, index)
}

func Test_avutil_channel_layout_check(t *testing.T) {
	assert := assert.New(t)

	// Valid layout
	var validLayout AVChannelLayout
	err := AVUtil_channel_layout_from_string(&validLayout, "stereo")
	assert.NoError(err)
	defer AVUtil_channel_layout_uninit(&validLayout)

	assert.True(AVUtil_channel_layout_check(&validLayout), "Valid layout should pass check")

	// Empty/uninitialized layout
	var emptyLayout AVChannelLayout
	assert.False(AVUtil_channel_layout_check(&emptyLayout), "Empty layout should fail check")
}

func Test_avutil_channel_layout_compare(t *testing.T) {
	assert := assert.New(t)

	var layout1 AVChannelLayout
	err := AVUtil_channel_layout_from_string(&layout1, "stereo")
	assert.NoError(err)
	defer AVUtil_channel_layout_uninit(&layout1)

	var layout2 AVChannelLayout
	err = AVUtil_channel_layout_from_string(&layout2, "stereo")
	assert.NoError(err)
	defer AVUtil_channel_layout_uninit(&layout2)

	var layout3 AVChannelLayout
	err = AVUtil_channel_layout_from_string(&layout3, "mono")
	assert.NoError(err)
	defer AVUtil_channel_layout_uninit(&layout3)

	// Same layouts should be equal
	assert.True(AVUtil_channel_layout_compare(&layout1, &layout2), "Same layouts should be equal")

	// Different layouts should not be equal
	assert.False(AVUtil_channel_layout_compare(&layout1, &layout3), "Different layouts should not be equal")
}

////////////////////////////////////////////////////////////////////////////////
// TEST PROPERTIES

func Test_avutil_channel_layout_num_channels(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		str      string
		expected int
	}{
		{"mono", 1},
		{"stereo", 2},
		{"5.1", 6},
		{"7.1", 8},
	}

	for _, tc := range tests {
		var layout AVChannelLayout
		err := AVUtil_channel_layout_from_string(&layout, tc.str)
		assert.NoError(err)
		defer AVUtil_channel_layout_uninit(&layout)

		assert.Equal(tc.expected, layout.NumChannels())
		t.Logf("Layout %q has %d channels", tc.str, layout.NumChannels())
	}
}

func Test_avutil_channel_layout_order(t *testing.T) {
	assert := assert.New(t)

	var layout AVChannelLayout
	err := AVUtil_channel_layout_from_string(&layout, "stereo")
	assert.NoError(err)
	defer AVUtil_channel_layout_uninit(&layout)

	order := layout.Order()
	t.Logf("Stereo layout order: %d", order)
	// Order should be set to some valid value
	assert.GreaterOrEqual(int(order), 0)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avutil_channel_layout_marshal_json(t *testing.T) {
	assert := assert.New(t)

	var layout AVChannelLayout
	err := AVUtil_channel_layout_from_string(&layout, "stereo")
	assert.NoError(err)
	defer AVUtil_channel_layout_uninit(&layout)

	data, err := json.Marshal(&layout)
	assert.NoError(err)
	assert.NotEmpty(data)

	var result string
	err = json.Unmarshal(data, &result)
	assert.NoError(err)
	assert.NotEmpty(result)

	t.Logf("Stereo layout JSON: %s", string(data))
}

func Test_avutil_channel_layout_marshal_json_empty(t *testing.T) {
	assert := assert.New(t)

	var layout AVChannelLayout
	// Empty layout

	data, err := json.Marshal(&layout)
	assert.NoError(err)
	assert.Equal("null", string(data), "Empty layout should marshal to null")
}

func Test_avutil_channel_layout_marshal_json_various(t *testing.T) {
	assert := assert.New(t)

	layouts := []string{"mono", "stereo", "5.1", "7.1"}

	for _, layoutStr := range layouts {
		var layout AVChannelLayout
		err := AVUtil_channel_layout_from_string(&layout, layoutStr)
		assert.NoError(err)
		defer AVUtil_channel_layout_uninit(&layout)

		data, err := json.Marshal(&layout)
		assert.NoError(err)
		t.Logf("Layout %q JSON: %s", layoutStr, string(data))
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST THREAD SAFETY

func Test_avutil_channel_concurrent_operations(t *testing.T) {
	// Test that concurrent operations don't cause race conditions
	// This tests the fix for the global buffer issue
	assert := assert.New(t)

	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Test channel name
			name, err := AVUtil_channel_name(AVChannel(id % 10))
			if err == nil {
				assert.NotEmpty(name)
			}

			// Test channel description
			desc, err := AVUtil_channel_description(AVChannel(id % 10))
			if err == nil {
				assert.NotEmpty(desc)
			}

			// Test channel layout
			var layout AVChannelLayout
			layouts := []string{"mono", "stereo", "5.1"}
			err = AVUtil_channel_layout_from_string(&layout, layouts[id%len(layouts)])
			if err == nil {
				defer AVUtil_channel_layout_uninit(&layout)
				desc, err := AVUtil_channel_layout_describe(&layout)
				if err == nil {
					assert.NotEmpty(desc)
				}
			}
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent operations completed successfully")
}

func Test_avutil_channel_layout_concurrent_describe(t *testing.T) {
	// Specifically test the describe function concurrently
	assert := assert.New(t)

	var layout AVChannelLayout
	err := AVUtil_channel_layout_from_string(&layout, "5.1")
	assert.NoError(err)
	defer AVUtil_channel_layout_uninit(&layout)

	var wg sync.WaitGroup
	numGoroutines := 50
	results := make([]string, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			desc, err := AVUtil_channel_layout_describe(&layout)
			assert.NoError(err)
			results[id] = desc
		}(i)
	}

	wg.Wait()

	// All results should be the same and non-empty
	firstResult := results[0]
	assert.NotEmpty(firstResult)
	for i, result := range results {
		assert.Equal(firstResult, result, "Result %d differs", i)
	}

	t.Logf("All %d concurrent describe calls returned: %q", numGoroutines, firstResult)
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avutil_channel_layout_multiple_uninit(t *testing.T) {
	// Test that calling uninit multiple times doesn't crash
	var layout AVChannelLayout
	err := AVUtil_channel_layout_from_string(&layout, "stereo")
	assert.NoError(t, err)

	AVUtil_channel_layout_uninit(&layout)
	AVUtil_channel_layout_uninit(&layout)
	AVUtil_channel_layout_uninit(&layout)

	// Should not crash
	t.Log("Multiple uninit calls completed without crash")
}

func Test_avutil_channel_layout_zero_channels(t *testing.T) {
	assert := assert.New(t)

	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 0)

	channels := AVUtil_get_channel_layout_nb_channels(&layout)
	assert.Equal(0, channels)

	// Check should fail for zero channels
	assert.False(AVUtil_channel_layout_check(&layout))
}

func Test_avutil_get_channel_layout_nb_channels_various(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		layoutStr string
		expected  int
	}{
		{"mono", 1},
		{"stereo", 2},
		{"2.1", 3},
		{"5.1", 6},
		{"7.1", 8},
	}

	for _, tc := range tests {
		var layout AVChannelLayout
		err := AVUtil_channel_layout_from_string(&layout, tc.layoutStr)
		assert.NoError(err)
		defer AVUtil_channel_layout_uninit(&layout)

		count := AVUtil_get_channel_layout_nb_channels(&layout)
		assert.Equal(tc.expected, count, "Wrong channel count for %q", tc.layoutStr)
	}
}
