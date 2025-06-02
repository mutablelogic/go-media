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
	ctx, err := ff.AVFilterGraph_create_filter(graph, filter, "zzz", "")
	assert.NoError(err)
	assert.NotNil(ctx)

	// We don't need to free the filter context, as it is freed when the graph is freed
}

func Test_avfilter_graph_002(t *testing.T) {
	assert := assert.New(t)
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	// Parse a filter graph, and return the inputs and outputs, which should be
	// freed when the graph is freed
	in, out, err := ff.AVFilterGraph_parse(graph, "[a]null[b]")
	assert.NoError(err)
	defer ff.AVFilterInOut_list_free(in)
	defer ff.AVFilterInOut_list_free(out)

	t.Log("graph=", graph)
	t.Log("in=", in)
	t.Log("out=", out)

	// One input and one output
	assert.Len(in, 1)
	assert.Equal("a", in[0].Name())
	assert.Len(out, 1)
	assert.Equal("b", out[0].Name())
}
