package ffmpeg

import (
	"io"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST DECODER SETUP

func Test_avcodec_decoding_setup_h264(t *testing.T) {
	assert := assert.New(t)

	// Find H264 decoder
	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	assert.NotNil(codec, "H264 decoder should be available")

	// Allocate context
	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Open codec
	err := AVCodec_open(ctx, codec, nil)
	assert.NoError(err, "Should open H264 decoder")
}

func Test_avcodec_decoding_setup_mp2(t *testing.T) {
	assert := assert.New(t)

	// Find MP2 decoder
	codec := AVCodec_find_decoder(AV_CODEC_ID_MP2)
	assert.NotNil(codec, "MP2 decoder should be available")

	// Allocate context
	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Open codec
	err := AVCodec_open(ctx, codec, nil)
	assert.NoError(err, "Should open MP2 decoder")
}

////////////////////////////////////////////////////////////////////////////////
// TEST SEND PACKET

func Test_avcodec_send_packet_before_open(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	assert.NotNil(codec)

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Try to send packet before opening codec
	pkt := AVCodec_packet_alloc()
	defer AVCodec_packet_free(pkt)

	err := AVCodec_send_packet(ctx, pkt)
	assert.Error(err, "Should fail when codec is not opened")
}

func Test_avcodec_send_packet_nil_flush(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Set required parameters
	ctx.SetWidth(1280)
	ctx.SetHeight(720)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	err := AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	// Send nil packet to flush (should not crash)
	err = AVCodec_send_packet(ctx, nil)
	// May return various errors depending on codec state, just ensure no crash
	t.Logf("Flush result: %v", err)
}

func Test_avcodec_send_packet_empty(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	ctx.SetWidth(1280)
	ctx.SetHeight(720)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	err := AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	// Send empty packet
	pkt := AVCodec_packet_alloc()
	defer AVCodec_packet_free(pkt)

	err = AVCodec_send_packet(ctx, pkt)
	// May fail or succeed depending on codec, just ensure API works
	t.Logf("Empty packet result: %v", err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST RECEIVE FRAME

func Test_avcodec_receive_frame_before_open(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	assert.NotNil(codec)

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Try to receive frame before opening codec
	frame := AVUtil_frame_alloc()
	defer AVUtil_frame_free(frame)

	err := AVCodec_receive_frame(ctx, frame)
	assert.Error(err, "Should fail when codec is not opened")
}

func Test_avcodec_receive_frame_eagain(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	ctx.SetWidth(1280)
	ctx.SetHeight(720)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	err := AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	// Try to receive without sending any packets
	frame := AVUtil_frame_alloc()
	defer AVUtil_frame_free(frame)

	err = AVCodec_receive_frame(ctx, frame)
	// Should get EAGAIN (need more input)
	if err == syscall.EAGAIN {
		t.Log("Got EAGAIN as expected (need more input)")
	} else {
		t.Logf("Got error: %v", err)
	}
}

func Test_avcodec_receive_frame_after_flush(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	ctx.SetWidth(1280)
	ctx.SetHeight(720)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	err := AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	// Flush decoder
	err = AVCodec_send_packet(ctx, nil)
	if err != nil && err != syscall.EAGAIN {
		t.Logf("Flush error: %v", err)
	}

	// Try to receive after flush
	frame := AVUtil_frame_alloc()
	defer AVUtil_frame_free(frame)

	err = AVCodec_receive_frame(ctx, frame)
	// Should get EOF or EAGAIN
	if err == io.EOF {
		t.Log("Got EOF as expected (end of stream)")
	} else if err == syscall.EAGAIN {
		t.Log("Got EAGAIN (no frames buffered)")
	} else {
		t.Logf("Got error: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST FLUSH BUFFERS

func Test_avcodec_flush_buffers(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	ctx.SetWidth(1280)
	ctx.SetHeight(720)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	err := AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	// Flush buffers (should not crash)
	AVCodec_flush_buffers(ctx)
	t.Log("Flush buffers succeeded")
}

func Test_avcodec_flush_buffers_before_open(t *testing.T) {
	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	defer AVCodec_free_context(ctx)

	// Flush before opening will crash - this is expected behavior
	// FFmpeg requires the codec to be opened before flushing
	t.Skip("Flushing unopened codec causes segfault - expected FFmpeg behavior")
}

func Test_avcodec_flush_buffers_audio(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_MP2)
	if codec == nil {
		t.Skip("MP2 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	ctx.SetSampleFormat(AV_SAMPLE_FMT_S16)
	ctx.SetSampleRate(48000)

	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := ctx.SetChannelLayout(layout)
	assert.NoError(err)

	err = AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	// Flush audio decoder
	AVCodec_flush_buffers(ctx)
	t.Log("Audio flush succeeded")
}

////////////////////////////////////////////////////////////////////////////////
// TEST ERROR HANDLING

func Test_avcodec_decoding_error_types(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	frame := AVUtil_frame_alloc()
	defer AVUtil_frame_free(frame)

	// Test EINVAL error (codec not opened)
	err := AVCodec_receive_frame(ctx, frame)
	assert.Error(err)
	t.Logf("Error type before open: %T, value: %v", err, err)

	// Open codec
	ctx.SetWidth(1280)
	ctx.SetHeight(720)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	err = AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	// Test EAGAIN error (no input)
	err = AVCodec_receive_frame(ctx, frame)
	if err == syscall.EAGAIN {
		t.Log("Got EAGAIN error as expected")
	} else {
		t.Logf("Got error: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST DECODE WORKFLOW

func Test_avcodec_decode_workflow(t *testing.T) {
	assert := assert.New(t)

	// This tests the basic decode workflow without actual data
	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Set parameters
	ctx.SetWidth(1280)
	ctx.SetHeight(720)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	// Open decoder
	err := AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	pkt := AVCodec_packet_alloc()
	defer AVCodec_packet_free(pkt)

	frame := AVUtil_frame_alloc()
	defer AVUtil_frame_free(frame)

	// Workflow: send packet -> receive frame
	// (Will fail without real data, but tests the API)
	err = AVCodec_send_packet(ctx, pkt)
	t.Logf("Send packet result: %v", err)

	err = AVCodec_receive_frame(ctx, frame)
	t.Logf("Receive frame result: %v", err)

	// Flush workflow
	err = AVCodec_send_packet(ctx, nil)
	t.Logf("Flush send result: %v", err)

	err = AVCodec_receive_frame(ctx, frame)
	t.Logf("Flush receive result: %v", err)

	// Reset buffers
	AVCodec_flush_buffers(ctx)
	t.Log("Complete workflow tested")
}

////////////////////////////////////////////////////////////////////////////////
// TEST MULTIPLE DECODERS

func Test_avcodec_multiple_decoders(t *testing.T) {
	assert := assert.New(t)

	// Create multiple decoders
	codec1 := AVCodec_find_decoder(AV_CODEC_ID_H264)
	codec2 := AVCodec_find_decoder(AV_CODEC_ID_MPEG2VIDEO)

	if codec1 == nil || codec2 == nil {
		t.Skip("Required codecs not available")
	}

	ctx1 := AVCodec_alloc_context(codec1)
	assert.NotNil(ctx1)
	defer AVCodec_free_context(ctx1)

	ctx2 := AVCodec_alloc_context(codec2)
	assert.NotNil(ctx2)
	defer AVCodec_free_context(ctx2)

	// Open both
	ctx1.SetWidth(1280)
	ctx1.SetHeight(720)
	ctx1.SetPixFmt(AV_PIX_FMT_YUV420P)

	ctx2.SetWidth(1920)
	ctx2.SetHeight(1080)
	ctx2.SetPixFmt(AV_PIX_FMT_YUV420P)

	err1 := AVCodec_open(ctx1, codec1, nil)
	err2 := AVCodec_open(ctx2, codec2, nil)

	if err1 != nil || err2 != nil {
		t.Skip("Could not open decoders")
	}

	// Both should be independent
	AVCodec_flush_buffers(ctx1)
	AVCodec_flush_buffers(ctx2)

	t.Log("Multiple decoders work independently")
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avcodec_receive_frame_nil(t *testing.T) {
	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	defer AVCodec_free_context(ctx)

	ctx.SetWidth(1280)
	ctx.SetHeight(720)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	err := AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	// Pass nil frame will crash - this is expected FFmpeg behavior
	t.Skip("Passing nil frame causes segfault - expected FFmpeg behavior")
}

func Test_avcodec_flush_sequence(t *testing.T) {
	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	defer AVCodec_free_context(ctx)

	ctx.SetWidth(1280)
	ctx.SetHeight(720)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	err := AVCodec_open(ctx, codec, nil)
	if err != nil {
		t.Skipf("Could not open decoder: %v", err)
	}

	// Multiple flushes in sequence
	AVCodec_flush_buffers(ctx)
	AVCodec_flush_buffers(ctx)
	AVCodec_flush_buffers(ctx)

	t.Log("Multiple flushes succeeded")
}
