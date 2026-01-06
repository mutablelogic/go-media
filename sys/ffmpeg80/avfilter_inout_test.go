package ffmpeg_test

import (
	"testing"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
)

func Test_avfilter_inout_000(t *testing.T) {
	assert := assert.New(t)

	inout := ff.AVFilterInOut_alloc("in", nil, 0)
	assert.NotNil(inout)
	assert.Equal("in", inout.Name())
	assert.Equal(0, inout.Pad())
	assert.Nil(inout.Filter())
	assert.Nil(inout.Next())

	t.Log("inout=", inout)

	ff.AVFilterInOut_free(inout)
}

func Test_avfilter_inout_001(t *testing.T) {
	assert := assert.New(t)

	head := ff.AVFilterInOut_link(ff.AVFilterInOut_alloc("in1", nil, 0), ff.AVFilterInOut_alloc("in2", nil, 0))
	assert.NotNil(head)

	arr := ff.AVFilterInOut_list(head)
	assert.Equal(2, len(arr))

	t.Log("arr=", arr)
}

func Test_avfilter_inout_002(t *testing.T) {
	assert := assert.New(t)

	// Test SetNext and Next
	in1 := ff.AVFilterInOut_alloc("in1", nil, 0)
	in2 := ff.AVFilterInOut_alloc("in2", nil, 1)
	in3 := ff.AVFilterInOut_alloc("in3", nil, 2)

	in1.SetNext(in2)
	in2.SetNext(in3)

	assert.Equal(in2, in1.Next())
	assert.Equal(in3, in2.Next())
	assert.Nil(in3.Next())

	// Test Pad method
	assert.Equal(0, in1.Pad())
	assert.Equal(1, in2.Pad())
	assert.Equal(2, in3.Pad())

	ff.AVFilterInOut_free(in1)
}

func Test_avfilter_inout_003(t *testing.T) {
	assert := assert.New(t)

	// Test link with empty array
	head := ff.AVFilterInOut_link()
	assert.Nil(head)

	// Test link with one valid entry followed by nil (should return the valid entry)
	in1 := ff.AVFilterInOut_alloc("in1", nil, 0)
	head = ff.AVFilterInOut_link(in1, nil)
	assert.Equal(in1, head)
	assert.Nil(head.Next())
	ff.AVFilterInOut_free(in1)
}

func Test_avfilter_inout_004(t *testing.T) {
	assert := assert.New(t)

	// Test list on nil
	var inout *ff.AVFilterInOut
	arr := ff.AVFilterInOut_list(inout)
	assert.Nil(arr)

	// Test list_free on empty array
	ff.AVFilterInOut_list_free(nil)
	ff.AVFilterInOut_list_free([]*ff.AVFilterInOut{})
}

func Test_avfilter_inout_005(t *testing.T) {
	assert := assert.New(t)

	// Test with filter context
	graph := ff.AVFilterGraph_alloc()
	assert.NotNil(graph)
	defer ff.AVFilterGraph_free(graph)

	filter := ff.AVFilter_get_by_name("null")
	assert.NotNil(filter)

	ctx, err := ff.AVFilterGraph_create_filter(graph, filter, "test", "")
	assert.NoError(err)
	assert.NotNil(ctx)

	// Create inout with the filter context
	inout := ff.AVFilterInOut_alloc("test_in", ctx, 0)
	assert.NotNil(inout)
	assert.Equal("test_in", inout.Name())
	assert.Equal(ctx, inout.Filter())
	assert.Equal(0, inout.Pad())

	t.Log("inout with context=", inout)

	ff.AVFilterInOut_free(inout)
}
