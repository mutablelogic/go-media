package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST ALLOCATION AND DEALLOCATION

func Test_avcodec_parameters_alloc(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par, "Parameters allocation should succeed")

	AVCodec_parameters_free(par)
}

func Test_avcodec_parameters_free_nil(t *testing.T) {
	// Should not crash with nil parameters
	var par *AVCodecParameters
	AVCodec_parameters_free(par)
}

func Test_avcodec_parameters_multiple_alloc_free(t *testing.T) {
	assert := assert.New(t)

	// Allocate and free multiple parameters
	for i := 0; i < 100; i++ {
		par := AVCodec_parameters_alloc()
		assert.NotNil(par)
		AVCodec_parameters_free(par)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST CODEC TYPE

func Test_avcodec_parameters_codec_type(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Default should be unknown
	codecType := par.CodecType()
	t.Logf("Default codec type: %s", codecType)

	// Set video type
	par.SetCodecType(AVMEDIA_TYPE_VIDEO)
	assert.Equal(AVMEDIA_TYPE_VIDEO, par.CodecType())

	// Set audio type
	par.SetCodecType(AVMEDIA_TYPE_AUDIO)
	assert.Equal(AVMEDIA_TYPE_AUDIO, par.CodecType())
}

////////////////////////////////////////////////////////////////////////////////
// TEST CODEC ID

func Test_avcodec_parameters_codec_id(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Set codec ID
	par.SetCodecID(AV_CODEC_ID_H264)
	assert.Equal(AV_CODEC_ID_H264, par.CodecID())

	// Set different codec ID
	par.SetCodecID(AV_CODEC_ID_MP2)
	assert.Equal(AV_CODEC_ID_MP2, par.CodecID())
}

////////////////////////////////////////////////////////////////////////////////
// TEST CODEC TAG

func Test_avcodec_parameters_codec_tag(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Set codec tag
	par.SetCodecTag(0x64636F76) // 'dvco'
	assert.Equal(uint32(0x64636F76), par.CodecTag())
	t.Logf("Codec tag: 0x%08X", par.CodecTag())
}

////////////////////////////////////////////////////////////////////////////////
// TEST VIDEO PROPERTIES

func Test_avcodec_parameters_video_properties(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Set codec type to video
	par.SetCodecType(AVMEDIA_TYPE_VIDEO)

	// Set width
	par.SetWidth(1920)
	assert.Equal(1920, par.Width())

	// Set height
	par.SetHeight(1080)
	assert.Equal(1080, par.Height())

	// Set pixel format
	par.SetPixelFormat(AV_PIX_FMT_YUV420P)
	assert.Equal(AV_PIX_FMT_YUV420P, par.PixelFormat())
	assert.Equal(int(AV_PIX_FMT_YUV420P), par.Format())

	// Set sample aspect ratio
	sar := AVUtil_rational(16, 9)
	par.SetSampleAspectRatio(sar)
	retrievedSar := par.SampleAspectRatio()
	assert.True(AVUtil_rational_equal(sar, retrievedSar))
}

func Test_avcodec_parameters_pixel_format_audio_type(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Set codec type to audio
	par.SetCodecType(AVMEDIA_TYPE_AUDIO)

	// PixelFormat should return NONE for audio
	pixFmt := par.PixelFormat()
	assert.Equal(AV_PIX_FMT_NONE, pixFmt)
}

////////////////////////////////////////////////////////////////////////////////
// TEST AUDIO PROPERTIES

func Test_avcodec_parameters_audio_properties(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Set codec type to audio
	par.SetCodecType(AVMEDIA_TYPE_AUDIO)

	// Set sample format
	par.SetSampleFormat(AV_SAMPLE_FMT_S16)
	assert.Equal(AV_SAMPLE_FMT_S16, par.SampleFormat())
	assert.Equal(int(AV_SAMPLE_FMT_S16), par.Format())

	// Set sample rate
	par.SetSampleRate(48000)
	assert.Equal(48000, par.SampleRate())

	// Set frame size
	par.SetFrameSize(1024)
	assert.Equal(1024, par.FrameSize())
}

func Test_avcodec_parameters_sample_format_video_type(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Set codec type to video
	par.SetCodecType(AVMEDIA_TYPE_VIDEO)

	// SampleFormat should return NONE for video
	sampleFmt := par.SampleFormat()
	assert.Equal(AV_SAMPLE_FMT_NONE, sampleFmt)
}

func Test_avcodec_parameters_channel_layout(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Set codec type to audio
	par.SetCodecType(AVMEDIA_TYPE_AUDIO)

	// Set stereo channel layout
	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := par.SetChannelLayout(layout)
	assert.NoError(err)

	retrievedLayout := par.ChannelLayout()
	assert.Equal(2, retrievedLayout.NumChannels())
}

func Test_avcodec_parameters_channel_layout_multiple_sets(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	par.SetCodecType(AVMEDIA_TYPE_AUDIO)

	// Set stereo first
	var layout1 AVChannelLayout
	AVUtil_channel_layout_default(&layout1, 2)
	err := par.SetChannelLayout(layout1)
	assert.NoError(err)
	assert.Equal(2, par.ChannelLayout().NumChannels())

	// Set 5.1 surround (should free previous layout)
	var layout2 AVChannelLayout
	AVUtil_channel_layout_default(&layout2, 6)
	err = par.SetChannelLayout(layout2)
	assert.NoError(err)
	assert.Equal(6, par.ChannelLayout().NumChannels())

	// Set mono (should free previous layout)
	var layout3 AVChannelLayout
	AVUtil_channel_layout_default(&layout3, 1)
	err = par.SetChannelLayout(layout3)
	assert.NoError(err)
	assert.Equal(1, par.ChannelLayout().NumChannels())
}

func Test_avcodec_parameters_channel_layout_invalid(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	par.SetCodecType(AVMEDIA_TYPE_AUDIO)

	// Try to set invalid channel layout
	var invalidLayout AVChannelLayout
	err := par.SetChannelLayout(invalidLayout)
	assert.Error(err, "Should fail with invalid channel layout")
}

////////////////////////////////////////////////////////////////////////////////
// TEST BITRATE

func Test_avcodec_parameters_bitrate(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Set bitrate
	par.SetBitRate(5000000)
	assert.Equal(int64(5000000), par.BitRate())

	// Set different bitrate
	par.SetBitRate(128000)
	assert.Equal(int64(128000), par.BitRate())
}

////////////////////////////////////////////////////////////////////////////////
// TEST COPY

func Test_avcodec_parameters_copy(t *testing.T) {
	assert := assert.New(t)

	src := AVCodec_parameters_alloc()
	assert.NotNil(src)
	defer AVCodec_parameters_free(src)

	dst := AVCodec_parameters_alloc()
	assert.NotNil(dst)
	defer AVCodec_parameters_free(dst)

	// Set up source parameters
	src.SetCodecType(AVMEDIA_TYPE_VIDEO)
	src.SetCodecID(AV_CODEC_ID_H264)
	src.SetWidth(1920)
	src.SetHeight(1080)
	src.SetPixelFormat(AV_PIX_FMT_YUV420P)
	src.SetBitRate(5000000)

	// Copy parameters
	err := AVCodec_parameters_copy(dst, src)
	assert.NoError(err)

	// Verify copy
	assert.Equal(src.CodecType(), dst.CodecType())
	assert.Equal(src.CodecID(), dst.CodecID())
	assert.Equal(src.Width(), dst.Width())
	assert.Equal(src.Height(), dst.Height())
	assert.Equal(src.PixelFormat(), dst.PixelFormat())
	assert.Equal(src.BitRate(), dst.BitRate())
}

func Test_avcodec_parameters_copy_audio(t *testing.T) {
	assert := assert.New(t)

	src := AVCodec_parameters_alloc()
	assert.NotNil(src)
	defer AVCodec_parameters_free(src)

	dst := AVCodec_parameters_alloc()
	assert.NotNil(dst)
	defer AVCodec_parameters_free(dst)

	// Set up source parameters
	src.SetCodecType(AVMEDIA_TYPE_AUDIO)
	src.SetCodecID(AV_CODEC_ID_MP2)
	src.SetSampleFormat(AV_SAMPLE_FMT_S16)
	src.SetSampleRate(48000)
	src.SetFrameSize(1024)
	src.SetBitRate(128000)

	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := src.SetChannelLayout(layout)
	assert.NoError(err)

	// Copy parameters
	err = AVCodec_parameters_copy(dst, src)
	assert.NoError(err)

	// Verify copy
	assert.Equal(src.CodecType(), dst.CodecType())
	assert.Equal(src.CodecID(), dst.CodecID())
	assert.Equal(src.SampleFormat(), dst.SampleFormat())
	assert.Equal(src.SampleRate(), dst.SampleRate())
	assert.Equal(src.FrameSize(), dst.FrameSize())
	assert.Equal(src.BitRate(), dst.BitRate())
	assert.Equal(src.ChannelLayout().NumChannels(), dst.ChannelLayout().NumChannels())
}

////////////////////////////////////////////////////////////////////////////////
// TEST CONTEXT CONVERSION

func Test_avcodec_parameters_from_context_video(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_encoder(AV_CODEC_ID_MPEG2VIDEO)
	if codec == nil {
		t.Skip("MPEG2VIDEO encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Set up context
	ctx.SetWidth(1920)
	ctx.SetHeight(1080)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)
	ctx.SetBitRate(5000000)

	// Create parameters and copy from context
	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	err := AVCodec_parameters_from_context(par, ctx)
	assert.NoError(err)

	// Verify parameters match context
	assert.Equal(AVMEDIA_TYPE_VIDEO, par.CodecType())
	assert.Equal(ctx.Width(), par.Width())
	assert.Equal(ctx.Height(), par.Height())
	assert.Equal(ctx.PixFmt(), par.PixelFormat())
	assert.Equal(ctx.BitRate(), par.BitRate())
}

func Test_avcodec_parameters_to_context_video(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Create parameters and set values
	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	par.SetCodecType(AVMEDIA_TYPE_VIDEO)
	par.SetCodecID(AV_CODEC_ID_H264)
	par.SetWidth(1280)
	par.SetHeight(720)
	par.SetPixelFormat(AV_PIX_FMT_YUV420P)

	// Copy to context
	err := AVCodec_parameters_to_context(ctx, par)
	assert.NoError(err)

	// Verify context matches parameters
	assert.Equal(par.Width(), ctx.Width())
	assert.Equal(par.Height(), ctx.Height())
	assert.Equal(par.PixelFormat(), ctx.PixFmt())
}

func Test_avcodec_parameters_from_context_audio(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_encoder(AV_CODEC_ID_MP2)
	if codec == nil {
		t.Skip("MP2 encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Set up context
	ctx.SetSampleFormat(AV_SAMPLE_FMT_S16)
	ctx.SetSampleRate(48000)
	ctx.SetBitRate(128000)

	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := ctx.SetChannelLayout(layout)
	assert.NoError(err)

	// Create parameters and copy from context
	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	err = AVCodec_parameters_from_context(par, ctx)
	assert.NoError(err)

	// Verify parameters match context
	assert.Equal(AVMEDIA_TYPE_AUDIO, par.CodecType())
	assert.Equal(ctx.SampleFormat(), par.SampleFormat())
	assert.Equal(ctx.SampleRate(), par.SampleRate())
	assert.Equal(ctx.BitRate(), par.BitRate())
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avcodec_parameters_json_video(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	par.SetCodecType(AVMEDIA_TYPE_VIDEO)
	par.SetCodecID(AV_CODEC_ID_H264)
	par.SetWidth(1920)
	par.SetHeight(1080)
	par.SetPixelFormat(AV_PIX_FMT_YUV420P)
	par.SetBitRate(5000000)

	data, err := json.Marshal(par)
	assert.NoError(err)
	assert.NotEmpty(data)

	jsonStr := string(data)
	assert.Contains(jsonStr, "codec_type")
	assert.Contains(jsonStr, "codec_id")
	assert.Contains(jsonStr, "pixel_format")
	assert.Contains(jsonStr, "width")
	assert.Contains(jsonStr, "height")
	assert.Contains(jsonStr, "1920")
	assert.Contains(jsonStr, "1080")
	t.Logf("Video parameters JSON: %s", jsonStr)
}

func Test_avcodec_parameters_json_audio(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	par.SetCodecType(AVMEDIA_TYPE_AUDIO)
	par.SetCodecID(AV_CODEC_ID_MP2)
	par.SetSampleFormat(AV_SAMPLE_FMT_S16)
	par.SetSampleRate(48000)
	par.SetFrameSize(1024)
	par.SetBitRate(128000)

	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := par.SetChannelLayout(layout)
	assert.NoError(err)

	data, err := json.Marshal(par)
	assert.NoError(err)
	assert.NotEmpty(data)

	jsonStr := string(data)
	assert.Contains(jsonStr, "codec_type")
	assert.Contains(jsonStr, "codec_id")
	assert.Contains(jsonStr, "sample_format")
	assert.Contains(jsonStr, "sample_rate")
	assert.Contains(jsonStr, "48000")
	assert.Contains(jsonStr, "frame_size")
	assert.Contains(jsonStr, "1024")
	t.Logf("Audio parameters JSON: %s", jsonStr)
}

////////////////////////////////////////////////////////////////////////////////
// TEST STRING

func Test_avcodec_parameters_string(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	par.SetCodecType(AVMEDIA_TYPE_VIDEO)
	par.SetCodecID(AV_CODEC_ID_H264)
	par.SetWidth(1920)
	par.SetHeight(1080)
	par.SetPixelFormat(AV_PIX_FMT_YUV420P)

	str := par.String()
	assert.NotEmpty(str)
	t.Logf("Parameters string:\n%s", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST FORMAT

func Test_avcodec_parameters_format(t *testing.T) {
	assert := assert.New(t)

	// Test video format
	parVideo := AVCodec_parameters_alloc()
	assert.NotNil(parVideo)
	defer AVCodec_parameters_free(parVideo)

	parVideo.SetCodecType(AVMEDIA_TYPE_VIDEO)
	parVideo.SetPixelFormat(AV_PIX_FMT_YUV420P)
	assert.Equal(int(AV_PIX_FMT_YUV420P), parVideo.Format())

	// Test audio format
	parAudio := AVCodec_parameters_alloc()
	assert.NotNil(parAudio)
	defer AVCodec_parameters_free(parAudio)

	parAudio.SetCodecType(AVMEDIA_TYPE_AUDIO)
	parAudio.SetSampleFormat(AV_SAMPLE_FMT_S16)
	assert.Equal(int(AV_SAMPLE_FMT_S16), parAudio.Format())
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avcodec_parameters_default_values(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Check default values
	codecType := par.CodecType()
	t.Logf("Default codec type: %s", codecType)

	codecID := par.CodecID()
	t.Logf("Default codec ID: %s", codecID)

	bitrate := par.BitRate()
	assert.Equal(int64(0), bitrate)

	width := par.Width()
	assert.Equal(0, width)

	height := par.Height()
	assert.Equal(0, height)
}

func Test_avcodec_parameters_zero_dimensions(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	par.SetCodecType(AVMEDIA_TYPE_VIDEO)

	// Set zero dimensions (allowed in parameters, validation happens elsewhere)
	par.SetWidth(0)
	par.SetHeight(0)
	assert.Equal(0, par.Width())
	assert.Equal(0, par.Height())
}

func Test_avcodec_parameters_negative_values(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Negative bitrate
	par.SetBitRate(-1000)
	assert.Equal(int64(-1000), par.BitRate())

	// Negative dimensions (shouldn't happen in practice, but parameters allow it)
	par.SetWidth(-100)
	assert.Equal(-100, par.Width())
}

func Test_avcodec_parameters_large_values(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Very large bitrate
	par.SetBitRate(1000000000)
	assert.Equal(int64(1000000000), par.BitRate())

	// Large dimensions (8K)
	par.SetWidth(7680)
	par.SetHeight(4320)
	assert.Equal(7680, par.Width())
	assert.Equal(4320, par.Height())

	// Very high sample rate
	par.SetSampleRate(192000)
	assert.Equal(192000, par.SampleRate())
}

func Test_avcodec_parameters_multiple_type_changes(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	// Start as video
	par.SetCodecType(AVMEDIA_TYPE_VIDEO)
	par.SetPixelFormat(AV_PIX_FMT_YUV420P)
	assert.Equal(AVMEDIA_TYPE_VIDEO, par.CodecType())
	assert.Equal(AV_PIX_FMT_YUV420P, par.PixelFormat())

	// Change to audio
	par.SetCodecType(AVMEDIA_TYPE_AUDIO)
	assert.Equal(AVMEDIA_TYPE_AUDIO, par.CodecType())
	// PixelFormat should now return NONE
	assert.Equal(AV_PIX_FMT_NONE, par.PixelFormat())

	// Change back to video
	par.SetCodecType(AVMEDIA_TYPE_VIDEO)
	assert.Equal(AVMEDIA_TYPE_VIDEO, par.CodecType())
}

func Test_avcodec_parameters_sample_aspect_ratio(t *testing.T) {
	assert := assert.New(t)

	par := AVCodec_parameters_alloc()
	assert.NotNil(par)
	defer AVCodec_parameters_free(par)

	par.SetCodecType(AVMEDIA_TYPE_VIDEO)

	tests := []struct {
		num int
		den int
	}{
		{1, 1},   // Square pixels
		{4, 3},   // 4:3
		{16, 9},  // 16:9
		{16, 15}, // PAL DV
		{40, 33}, // NTSC DV
	}

	for _, tc := range tests {
		sar := AVUtil_rational(tc.num, tc.den)
		par.SetSampleAspectRatio(sar)
		retrieved := par.SampleAspectRatio()
		assert.True(AVUtil_rational_equal(sar, retrieved))
		t.Logf("SAR %d/%d: %s", tc.num, tc.den, retrieved.String())
	}
}
