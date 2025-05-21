package ffmpeg_test

import (
	"testing"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
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
