package ffmpeg_test

import (
	"testing"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
)

func Test_avfilter_filter_000(t *testing.T) {
	//assert := assert.New(t)

	// Iterate over all filters
	var opaque uintptr
	for {
		filter := ff.AVFilter_iterate(&opaque)
		if filter == nil {
			break
		}

		t.Log("filter=", filter)
	}
}

func Test_avfilter_filter_001(t *testing.T) {
	assert := assert.New(t)

	// Iterate over all filters
	var opaque uintptr
	for {
		filter := ff.AVFilter_iterate(&opaque)
		if filter == nil {
			break
		}
		filter2 := ff.AVFilter_get_by_name(filter.Name())
		assert.NotNil(filter2)
		assert.Equal(filter, filter2)
	}
}

func Test_avfilter_filter_002(t *testing.T) {
	assert := assert.New(t)

	// Get a specific filter and test its methods
	filter := ff.AVFilter_get_by_name("scale")
	assert.NotNil(filter)
	assert.Equal("scale", filter.Name())
	assert.NotEmpty(filter.Description())
	assert.Greater(filter.NumInputs(), uint(0))
	assert.Greater(filter.NumOutputs(), uint(0))

	t.Log("filter=", filter)
	t.Log("description=", filter.Description())
	t.Log("flags=", filter.Flags())
	t.Log("inputs=", filter.NumInputs())
	t.Log("outputs=", filter.NumOutputs())
}

func Test_avfilter_filter_003(t *testing.T) {
	assert := assert.New(t)

	// Test filter with dynamic inputs/outputs
	filter := ff.AVFilter_get_by_name("concat")
	assert.NotNil(filter)

	flags := filter.Flags()
	if flags.Is(ff.AVFILTER_FLAG_DYNAMIC_INPUTS) {
		t.Log("concat has dynamic inputs")
	}
	if flags.Is(ff.AVFILTER_FLAG_DYNAMIC_OUTPUTS) {
		t.Log("concat has dynamic outputs")
	}
}

func Test_avfilter_filter_004(t *testing.T) {
	assert := assert.New(t)

	// Test invalid filter lookup
	filter := ff.AVFilter_get_by_name("nonexistent_filter_12345")
	assert.Nil(filter)
}

func Test_avfilter_flag_000(t *testing.T) {
	assert := assert.New(t)

	// Test flag operations
	flag := ff.AVFILTER_FLAG_DYNAMIC_INPUTS | ff.AVFILTER_FLAG_SLICE_THREADS
	assert.True(flag.Is(ff.AVFILTER_FLAG_DYNAMIC_INPUTS))
	assert.True(flag.Is(ff.AVFILTER_FLAG_SLICE_THREADS))
	assert.False(flag.Is(ff.AVFILTER_FLAG_DYNAMIC_OUTPUTS))

	// Test string representation
	assert.Contains(flag.String(), "DYNAMIC_INPUTS")
	assert.Contains(flag.String(), "SLICE_THREADS")

	t.Log("flag=", flag.String())
}

func Test_avfilter_flag_001(t *testing.T) {
	// Test AVFILTER_FLAG_NONE
	flag := ff.AVFILTER_FLAG_NONE
	assert.Equal(t, "AVFILTER_FLAG_NONE", flag.String())

	// Test AVFILTER_FLAG_SUPPORT_TIMELINE
	flag = ff.AVFILTER_FLAG_SUPPORT_TIMELINE
	assert.Contains(t, flag.String(), "TIMELINE")
}
