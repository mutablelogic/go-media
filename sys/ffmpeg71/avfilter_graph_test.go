package ffmpeg_test

import (
	"testing"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
	assert "github.com/stretchr/testify/assert"
)

func Test_avfilter_graph_000(t *testing.T) {
	assert := assert.New(t)
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	ff.AVFilterGraph_free(graph)
}

func Test_avfilter_graph_001(t *testing.T) {
	assert := assert.New(t)
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	// Create a filter
	filter := ff.AVFilter_get_by_name("null")
	assert.NotNil(filter)

	// Create a filter context
	ctx, err := ff.AVFilterGraph_create_filter(graph, filter, "null", "")
	assert.NoError(err)
	assert.NotNil(ctx)

	// We don't need to free the filter context, as it is freed when the graph is freed
}
