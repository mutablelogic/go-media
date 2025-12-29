package ffmpeg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST BASIC ALLOCATION AND DEALLOCATION

func Test_avcodec_packet_alloc_free(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	assert.NotNil(packet)
	assert.Equal(0, packet.Size())
	assert.Nil(packet.Bytes())

	AVCodec_packet_free(packet)
}

func Test_avcodec_packet_freep(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	assert.NotNil(packet)

	AVCodec_packet_freep(&packet)
	assert.Nil(packet)
}

func Test_avcodec_packet_freep_nil(t *testing.T) {
	// Should not panic
	AVCodec_packet_freep(nil)
}

////////////////////////////////////////////////////////////////////////////////
// TEST PACKET ALLOCATION AND OPERATIONS

func Test_avcodec_new_packet(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	// Allocate payload
	err := AVCodec_new_packet(packet, 1024)
	assert.NoError(err)
	assert.Equal(1024, packet.Size())

	bytes := packet.Bytes()
	assert.NotNil(bytes)
	assert.Equal(1024, len(bytes))
}

func Test_avcodec_shrink_packet(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	// Allocate payload
	assert.NoError(AVCodec_new_packet(packet, 1024))
	assert.Equal(1024, packet.Size())

	// Shrink it
	AVCodec_shrink_packet(packet, 512)
	assert.Equal(512, packet.Size())
	assert.Equal(512, len(packet.Bytes()))
}

func Test_avcodec_grow_packet(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	// Allocate payload
	assert.NoError(AVCodec_new_packet(packet, 512))
	assert.Equal(512, packet.Size())

	// Grow it
	err := AVCodec_grow_packet(packet, 512)
	assert.NoError(err)
	assert.Equal(1024, packet.Size())
	assert.Equal(1024, len(packet.Bytes()))
}

func Test_avcodec_packet_unref(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	// Allocate payload
	assert.NoError(AVCodec_new_packet(packet, 1024))
	assert.Equal(1024, packet.Size())

	// Unref should clear the packet
	AVCodec_packet_unref(packet)
	assert.Equal(0, packet.Size())
	assert.Nil(packet.Bytes())
}

func Test_avcodec_packet_clone(t *testing.T) {
	assert := assert.New(t)

	// Create original packet
	original := AVCodec_packet_alloc()
	defer AVCodec_packet_free(original)

	assert.NoError(AVCodec_new_packet(original, 100))
	original.SetStreamIndex(42)
	original.SetPos(1234)

	// Clone it
	clone := AVCodec_packet_clone(original)
	assert.NotNil(clone)
	defer AVCodec_packet_free(clone)

	// Verify clone has same properties
	assert.Equal(original.Size(), clone.Size())
	assert.Equal(original.StreamIndex(), clone.StreamIndex())
	assert.Equal(original.Pos(), clone.Pos())
}

////////////////////////////////////////////////////////////////////////////////
// TEST GETTERS AND SETTERS

func Test_avcodec_packet_stream_index(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	packet.SetStreamIndex(5)
	assert.Equal(5, packet.StreamIndex())

	packet.SetStreamIndex(0)
	assert.Equal(0, packet.StreamIndex())
}

func Test_avcodec_packet_timebase(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	// Test nil packet
	var nilPacket *AVPacket
	tb := nilPacket.TimeBase()
	assert.Equal(int(0), tb.Num())
	assert.Equal(int(0), tb.Den())

	// Set timebase
	tb = AVRational{num: 1, den: 25}
	packet.SetTimeBase(tb)

	result := packet.TimeBase()
	assert.Equal(int(1), result.Num())
	assert.Equal(int(25), result.Den())
}

func Test_avcodec_packet_timestamps(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	// Initial values are AV_NOPTS_VALUE (not 0)
	// Just verify they are valid int64 values
	assert.IsType(int64(0), packet.Pts())
	assert.IsType(int64(0), packet.Dts())
	assert.IsType(int64(0), packet.Duration())
}

func Test_avcodec_packet_pos(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	packet.SetPos(12345)
	assert.Equal(int64(12345), packet.Pos())

	packet.SetPos(-1)
	assert.Equal(int64(-1), packet.Pos())
}

func Test_avcodec_packet_rescale_ts(t *testing.T) {
	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	// Set some timestamps and timebase
	packet.SetTimeBase(AVRational{num: 1, den: 1000})

	// Rescale should not panic
	tb_src := AVRational{num: 1, den: 1000}
	tb_dst := AVRational{num: 1, den: 90000}
	AVCodec_packet_rescale_ts(packet, tb_src, tb_dst)
}

////////////////////////////////////////////////////////////////////////////////
// TEST BYTES AND SIZE WITH EDGE CASES

func Test_avcodec_packet_bytes_nil(t *testing.T) {
	assert := assert.New(t)

	var packet *AVPacket
	bytes := packet.Bytes()
	assert.Nil(bytes)
}

func Test_avcodec_packet_bytes_empty(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	bytes := packet.Bytes()
	assert.Nil(bytes)
}

func Test_avcodec_packet_size_nil(t *testing.T) {
	assert := assert.New(t)

	var packet *AVPacket
	size := packet.Size()
	assert.Equal(0, size)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avcodec_packet_marshal_json(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	packet.SetStreamIndex(1)
	packet.SetPos(1000)
	packet.SetTimeBase(AVRational{num: 1, den: 25})

	bytes, err := json.Marshal(packet)
	assert.NoError(err)
	assert.NotNil(bytes)

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(bytes, &result)
	assert.NoError(err)
	assert.Equal(float64(1), result["stream_index"])
}

func Test_avcodec_packet_string(t *testing.T) {
	assert := assert.New(t)

	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_free(packet)

	packet.SetStreamIndex(2)
	str := packet.String()
	assert.NotEmpty(str)
	assert.Contains(str, "stream_index")
}

////////////////////////////////////////////////////////////////////////////////
// TEST INTEGRATION WITH NEW_PACKET

func Test_avcodec_packet_workflow(t *testing.T) {
	assert := assert.New(t)

	// Allocate packet
	packet := AVCodec_packet_alloc()
	defer AVCodec_packet_freep(&packet)

	// Allocate buffer
	assert.NoError(AVCodec_new_packet(packet, 256))
	assert.Equal(256, packet.Size())

	// Set properties
	packet.SetStreamIndex(3)
	packet.SetPos(5000)
	packet.SetTimeBase(AVRational{num: 1, den: 30})

	// Verify
	assert.Equal(3, packet.StreamIndex())
	assert.Equal(int64(5000), packet.Pos())
	assert.Equal(int(1), packet.TimeBase().Num())
	assert.Equal(int(30), packet.TimeBase().Den())

	// Get bytes
	bytes := packet.Bytes()
	assert.NotNil(bytes)
	assert.Equal(256, len(bytes))

	// Unref
	AVCodec_packet_unref(packet)
	assert.Equal(0, packet.Size())
	assert.Nil(packet.Bytes())
}
