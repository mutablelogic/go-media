package ffmpeg_test

import (
	"testing"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
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
