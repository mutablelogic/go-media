package avfilter_test

import (
	"testing"

	// Packages
	avfilter "github.com/mutablelogic/go-media/pkg/avfilter"
	assert "github.com/stretchr/testify/assert"
)

func Test_filter_001(t *testing.T) {
	assert := assert.New(t)

	filters := avfilter.Filters()
	assert.NotNil(filters)
	assert.NotEmpty(filters)
	for _, filter := range filters {
		filter2 := avfilter.NewFilter(filter.Key())
		assert.NotNil(filter2)
		assert.Equal(filter2, filter.Any())
	}
}
