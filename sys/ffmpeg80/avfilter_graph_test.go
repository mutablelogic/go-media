package ffmpeg_test

import (
	"testing"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
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

func Test_avfilter_graph_003(t *testing.T) {
	assert := assert.New(t)
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	// Create two filters and configure
	buffersrc := ff.AVFilter_get_by_name("buffer")
	assert.NotNil(buffersrc)
	buffersink := ff.AVFilter_get_by_name("buffersink")
	assert.NotNil(buffersink)

	// Create source and sink contexts
	src, err := ff.AVFilterGraph_create_filter(graph, buffersrc, "src", "video_size=640x480:pix_fmt=0:time_base=1/25:pixel_aspect=1/1")
	assert.NoError(err)
	assert.NotNil(src)

	sink, err := ff.AVFilterGraph_create_filter(graph, buffersink, "sink", "")
	assert.NoError(err)
	assert.NotNil(sink)

	// Test filter context methods
	assert.Equal("src", src.Name())
	assert.Equal(buffersrc, src.Filter())
	assert.Equal(uint(0), src.NumInputs())
	assert.Equal(uint(1), src.NumOutputs())

	assert.Equal("sink", sink.Name())
	assert.Equal(buffersink, sink.Filter())
	assert.Equal(uint(1), sink.NumInputs())
	assert.Equal(uint(0), sink.NumOutputs())

	// Test NumFilters
	assert.Equal(uint(2), graph.NumFilters())

	t.Log("src=", src)
	t.Log("sink=", sink)
	t.Log("graph=", graph)
}

func Test_avfilter_graph_004(t *testing.T) {
	assert := assert.New(t)
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	// Test NumFilters on empty graph
	assert.Equal(uint(0), graph.NumFilters())

	// Add a filter
	filter := ff.AVFilter_get_by_name("null")
	assert.NotNil(filter)
	ctx, err := ff.AVFilterGraph_create_filter(graph, filter, "test", "")
	assert.NoError(err)
	assert.NotNil(ctx)

	// Test NumFilters
	assert.Equal(uint(1), graph.NumFilters())

	t.Log("graph with 1 filter, num_filters=", graph.NumFilters())
}

func Test_avfilter_graph_005(t *testing.T) {
	assert := assert.New(t)
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	// Test parse error with invalid filter string
	in, out, err := ff.AVFilterGraph_parse(graph, "[invalid filter syntax!!!")
	assert.Error(err)
	assert.Nil(in)
	assert.Nil(out)

	t.Log("parse error=", err)
}

func Test_avfilter_graph_006(t *testing.T) {
	assert := assert.New(t)
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	// Test create_filter with invalid filter
	ctx, err := ff.AVFilterGraph_create_filter(graph, nil, "test", "")
	assert.Error(err)
	assert.Nil(ctx)
}

func Test_avfilter_graph_007(t *testing.T) {
	assert := assert.New(t)

	// Test NumFilters on nil graph
	var graph *ff.AVFilterGraph
	assert.Equal(uint(0), graph.NumFilters())
}
