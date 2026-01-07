package ffmpeg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST AVCodec PROPERTIES

func Test_avcodec_codec_properties(t *testing.T) {
	assert := assert.New(t)

	// Find a known codec
	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	assert.NotNil(codec, "H264 decoder should exist")

	// Test basic properties
	name := codec.Name()
	assert.NotEmpty(name)
	assert.Contains(name, "264")
	t.Logf("Codec name: %s", name)

	longName := codec.LongName()
	assert.NotEmpty(longName)
	t.Logf("Codec long name: %s", longName)

	codecType := codec.Type()
	assert.True(codecType.Is(AVMEDIA_TYPE_VIDEO))
	t.Logf("Codec type: %s", codecType)

	id := codec.ID()
	assert.Equal(AV_CODEC_ID_H264, id)
	t.Logf("Codec ID: %s", id)

	capabilities := codec.Capabilities()
	t.Logf("Codec capabilities: %s", capabilities)
}

func Test_avcodec_codec_audio_properties(t *testing.T) {
	assert := assert.New(t)

	// Find an audio codec
	codec := AVCodec_find_decoder(AV_CODEC_ID_MP2)
	assert.NotNil(codec, "MP2 decoder should exist")

	name := codec.Name()
	assert.NotEmpty(name)
	t.Logf("Audio codec name: %s", name)

	codecType := codec.Type()
	assert.True(codecType.Is(AVMEDIA_TYPE_AUDIO))
	t.Logf("Audio codec type: %s", codecType)
}

////////////////////////////////////////////////////////////////////////////////
// TEST AVCodec SUPPORTED FORMATS

func Test_avcodec_codec_supported_framerates(t *testing.T) {
	assert := assert.New(t)

	// Find a video codec
	codec := AVCodec_find_encoder(AV_CODEC_ID_MPEG1VIDEO)
	if codec == nil {
		t.Skip("MPEG1VIDEO encoder not available")
	}

	framerates := codec.SupportedFramerates()
	if framerates != nil {
		assert.Greater(len(framerates), 0)
		for _, fr := range framerates {
			assert.False(fr.IsZero())
			t.Logf("Supported framerate: %s", fr)
		}
	} else {
		t.Log("No specific framerates restriction (all supported)")
	}
}

func Test_avcodec_codec_pixel_formats(t *testing.T) {
	assert := assert.New(t)

	// Find a video codec
	codec := AVCodec_find_encoder(AV_CODEC_ID_MPEG2VIDEO)
	if codec == nil {
		t.Skip("MPEG2VIDEO encoder not available")
	}

	pixelFormats := codec.PixelFormats()
	if pixelFormats != nil {
		assert.Greater(len(pixelFormats), 0)
		for _, fmt := range pixelFormats {
			assert.NotEqual(AV_PIX_FMT_NONE, fmt)
			t.Logf("Supported pixel format: %s", fmt)
		}
	} else {
		t.Log("No specific pixel formats restriction")
	}
}

func Test_avcodec_codec_sample_formats(t *testing.T) {
	// Find an audio encoder
	var opaque uintptr
	for {
		codec := AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}

		if !codec.Type().Is(AVMEDIA_TYPE_AUDIO) {
			continue
		}

		if !AVCodec_is_encoder(codec) {
			continue
		}

		sampleFormats := codec.SampleFormats()
		if sampleFormats != nil && len(sampleFormats) > 0 {
			t.Logf("Codec %s sample formats:", codec.Name())
			for _, fmt := range sampleFormats {
				t.Logf("  %s", fmt)
			}
			return // Found at least one
		}
	}
}

func Test_avcodec_codec_supported_samplerates(t *testing.T) {
	// Find an audio encoder
	var opaque uintptr
	for {
		codec := AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}

		if !codec.Type().Is(AVMEDIA_TYPE_AUDIO) {
			continue
		}

		if !AVCodec_is_encoder(codec) {
			continue
		}

		samplerates := codec.SupportedSamplerates()
		if samplerates != nil && len(samplerates) > 0 {
			t.Logf("Codec %s samplerates:", codec.Name())
			for _, sr := range samplerates {
				t.Logf("  %d Hz", sr)
			}
			return // Found at least one
		}
	}
}

func Test_avcodec_codec_profiles(t *testing.T) {
	assert := assert.New(t)

	// H264 has profiles
	codec := AVCodec_find_encoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not available")
	}

	profiles := codec.Profiles()
	if profiles != nil && len(profiles) > 0 {
		assert.Greater(len(profiles), 0)
		for _, profile := range profiles {
			id := profile.ID()
			name := profile.Name()
			assert.NotEmpty(name)
			t.Logf("Profile ID=%d, Name=%s", id, name)
		}
	} else {
		t.Log("No profiles available")
	}
}

func Test_avcodec_codec_channel_layouts(t *testing.T) {
	// Find an audio encoder with channel layouts
	var opaque uintptr
	for {
		codec := AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}

		if !codec.Type().Is(AVMEDIA_TYPE_AUDIO) {
			continue
		}

		if !AVCodec_is_encoder(codec) {
			continue
		}

		channelLayouts := codec.ChannelLayouts()
		if channelLayouts != nil && len(channelLayouts) > 0 {
			t.Logf("Codec %s channel layouts:", codec.Name())
			for _, layout := range channelLayouts {
				t.Logf("  %d channels", layout.NumChannels())
			}
			return // Found at least one
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST AVCodecContext PROPERTIES

func Test_avcodec_context_video_properties(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_encoder(AV_CODEC_ID_MPEG2VIDEO)
	if codec == nil {
		t.Skip("MPEG2VIDEO encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Test video properties
	ctx.SetWidth(1920)
	assert.Equal(1920, ctx.Width())

	ctx.SetHeight(1080)
	assert.Equal(1080, ctx.Height())

	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)
	assert.Equal(AV_PIX_FMT_YUV420P, ctx.PixFmt())

	ctx.SetBitRate(5000000)
	assert.Equal(int64(5000000), ctx.BitRate())

	ctx.SetGopSize(30)
	assert.Equal(30, ctx.GopSize())

	ctx.SetMaxBFrames(2)
	assert.Equal(2, ctx.MaxBFrames())

	framerate := AVUtil_rational(30, 1)
	ctx.SetFramerate(framerate)
	retrievedFr := ctx.Framerate()
	assert.True(AVUtil_rational_equal(framerate, retrievedFr))

	timebase := AVUtil_rational(1, 30)
	ctx.SetTimeBase(timebase)
	retrievedTb := ctx.TimeBase()
	assert.True(AVUtil_rational_equal(timebase, retrievedTb))

	sar := AVUtil_rational(1, 1)
	ctx.SetSampleAspectRatio(sar)
	retrievedSar := ctx.SampleAspectRatio()
	assert.True(AVUtil_rational_equal(sar, retrievedSar))
}

func Test_avcodec_context_audio_properties(t *testing.T) {
	assert := assert.New(t)

	// Find an audio encoder
	codec := AVCodec_find_encoder(AV_CODEC_ID_MP2)
	if codec == nil {
		t.Skip("MP2 encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Test audio properties
	ctx.SetSampleFormat(AV_SAMPLE_FMT_S16)
	assert.Equal(AV_SAMPLE_FMT_S16, ctx.SampleFormat())

	ctx.SetSampleRate(48000)
	assert.Equal(48000, ctx.SampleRate())

	ctx.SetBitRate(128000)
	assert.Equal(int64(128000), ctx.BitRate())

	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := ctx.SetChannelLayout(layout)
	assert.NoError(err)

	retrievedLayout := ctx.ChannelLayout()
	assert.Equal(2, retrievedLayout.NumChannels())
}

func Test_avcodec_context_codec_properties(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	assert.NotNil(codec)

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Test codec-related properties
	retrievedCodec := ctx.Codec()
	assert.NotNil(retrievedCodec)
	assert.Equal(codec.Name(), retrievedCodec.Name())

	codecType := ctx.CodecType()
	assert.True(codecType.Is(AVMEDIA_TYPE_VIDEO))

	codecID := ctx.CodecID()
	assert.Equal(AV_CODEC_ID_H264, codecID)

	// Codec tag
	tag := ctx.CodecTag()
	t.Logf("Codec tag: 0x%08X", tag)
}

func Test_avcodec_context_flags(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_encoder(AV_CODEC_ID_MPEG2VIDEO)
	if codec == nil {
		t.Skip("MPEG2VIDEO encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	// Test flags
	ctx.SetFlags(AV_CODEC_FLAG_GLOBAL_HEADER | AV_CODEC_FLAG_LOW_DELAY)
	flags := ctx.Flags()
	t.Logf("Flags: 0x%08X", flags)

	ctx.SetFlags2(AV_CODEC_FLAG2_FAST | AV_CODEC_FLAG2_SHOW_ALL)
	flags2 := ctx.Flags2()
	t.Logf("Flags2: 0x%08X", flags2)
}

func Test_avcodec_context_priv_data(t *testing.T) {
	codec := AVCodec_find_encoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Skip("Failed to allocate context")
	}
	defer AVCodec_free_context(ctx)

	// Try setting a private data option
	// Note: This may fail if the option doesn't exist, but shouldn't crash
	err := ctx.SetPrivDataKV("preset", "ultrafast")
	if err != nil {
		t.Logf("SetPrivDataKV failed (expected if option not available): %v", err)
	} else {
		t.Log("SetPrivDataKV succeeded")
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST AVCodecID

func Test_avcodec_id_name(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		id   AVCodecID
		name string
	}{
		{AV_CODEC_ID_NONE, "none"},
		{AV_CODEC_ID_H264, "h264"},
		{AV_CODEC_ID_MP2, "mp2"},
		{AV_CODEC_ID_MPEG1VIDEO, "mpeg1video"},
		{AV_CODEC_ID_MPEG2VIDEO, "mpeg2video"},
	}

	for _, tc := range tests {
		name := tc.id.Name()
		assert.NotEmpty(name)
		t.Logf("Codec ID %d name: %s", tc.id, name)
		if tc.id != AV_CODEC_ID_NONE {
			assert.Contains(name, tc.name)
		}
	}
}

func Test_avcodec_id_type(t *testing.T) {
	assert := assert.New(t)

	videoType := AV_CODEC_ID_H264.Type()
	assert.True(videoType.Is(AVMEDIA_TYPE_VIDEO))

	audioType := AV_CODEC_ID_MP2.Type()
	assert.True(audioType.Is(AVMEDIA_TYPE_AUDIO))

	noneType := AV_CODEC_ID_NONE.Type()
	assert.True(noneType.Is(AVMEDIA_TYPE_UNKNOWN))
}

func Test_avcodec_id_string(t *testing.T) {
	assert := assert.New(t)

	str := AV_CODEC_ID_H264.String()
	assert.NotEmpty(str)
	assert.Contains(str, "264")
	t.Logf("H264 ID string: %s", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST AVCodecCap

func Test_avcodec_cap_is(t *testing.T) {
	assert := assert.New(t)

	caps := AV_CODEC_CAP_DR1 | AV_CODEC_CAP_DELAY | AV_CODEC_CAP_FRAME_THREADS

	assert.True(caps.Is(AV_CODEC_CAP_DR1))
	assert.True(caps.Is(AV_CODEC_CAP_DELAY))
	assert.True(caps.Is(AV_CODEC_CAP_FRAME_THREADS))
	assert.False(caps.Is(AV_CODEC_CAP_HARDWARE))
	assert.False(caps.Is(AV_CODEC_CAP_EXPERIMENTAL))
}

func Test_avcodec_cap_none(t *testing.T) {
	assert := assert.New(t)

	caps := AV_CODEC_CAP_NONE
	assert.False(caps.Is(AV_CODEC_CAP_DR1))
	assert.False(caps.Is(AV_CODEC_CAP_DELAY))
}

func Test_avcodec_cap_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		cap      AVCodecCap
		contains string
	}{
		{AV_CODEC_CAP_NONE, "AV_CODEC_CAP_NONE"},
		{AV_CODEC_CAP_DR1, "AV_CODEC_CAP_DR1"},
		{AV_CODEC_CAP_DELAY, "AV_CODEC_CAP_DELAY"},
		{AV_CODEC_CAP_FRAME_THREADS, "AV_CODEC_CAP_FRAME_THREADS"},
	}

	for _, tc := range tests {
		str := tc.cap.String()
		assert.Contains(str, tc.contains)
		t.Logf("Cap %d string: %s", tc.cap, str)
	}
}

func Test_avcodec_cap_multiple_flags(t *testing.T) {
	assert := assert.New(t)

	caps := AV_CODEC_CAP_DR1 | AV_CODEC_CAP_DELAY
	str := caps.String()
	assert.NotEmpty(str)
	assert.Contains(str, "AV_CODEC_CAP_DR1")
	assert.Contains(str, "AV_CODEC_CAP_DELAY")
	t.Logf("Multiple caps string: %s", str)
}

func Test_avcodec_cap_flag_string(t *testing.T) {
	assert := assert.New(t)

	str := AV_CODEC_CAP_EXPERIMENTAL.FlagString()
	assert.Equal("AV_CODEC_CAP_EXPERIMENTAL", str)

	str = AV_CODEC_CAP_HARDWARE.FlagString()
	assert.Equal("AV_CODEC_CAP_HARDWARE", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST AVProfile

func Test_avcodec_profile_properties(t *testing.T) {
	codec := AVCodec_find_encoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not available")
	}

	profiles := codec.Profiles()
	if profiles == nil || len(profiles) == 0 {
		t.Skip("No profiles available")
	}

	for i, profile := range profiles {
		id := profile.ID()
		name := profile.Name()
		t.Logf("Profile %d: ID=%d, Name=%s", i, id, name)
	}
}

func Test_avcodec_profile_string(t *testing.T) {
	codec := AVCodec_find_encoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not available")
	}

	profiles := codec.Profiles()
	if profiles == nil || len(profiles) == 0 {
		t.Skip("No profiles available")
	}

	str := profiles[0].String()
	t.Logf("Profile string: %s", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avcodec_codec_json(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	assert.NotNil(codec)

	data, err := json.Marshal(codec)
	assert.NoError(err)
	assert.NotEmpty(data)

	jsonStr := string(data)
	assert.Contains(jsonStr, "type")
	assert.Contains(jsonStr, "name")
	assert.Contains(jsonStr, "id")
	t.Logf("Codec JSON: %s", jsonStr)
}

func Test_avcodec_context_video_json(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_encoder(AV_CODEC_ID_MPEG2VIDEO)
	if codec == nil {
		t.Skip("MPEG2VIDEO encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	ctx.SetWidth(1920)
	ctx.SetHeight(1080)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)
	ctx.SetBitRate(5000000)

	data, err := json.Marshal(ctx)
	assert.NoError(err)
	assert.NotEmpty(data)

	jsonStr := string(data)
	assert.Contains(jsonStr, "codec")
	assert.Contains(jsonStr, "width")
	assert.Contains(jsonStr, "height")
	assert.Contains(jsonStr, "1920")
	assert.Contains(jsonStr, "1080")
	t.Logf("Video context JSON: %s", jsonStr)
}

func Test_avcodec_context_audio_json(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_encoder(AV_CODEC_ID_MP2)
	if codec == nil {
		t.Skip("MP2 encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	ctx.SetSampleFormat(AV_SAMPLE_FMT_S16)
	ctx.SetSampleRate(48000)
	ctx.SetBitRate(128000)

	data, err := json.Marshal(ctx)
	assert.NoError(err)
	assert.NotEmpty(data)

	jsonStr := string(data)
	assert.Contains(jsonStr, "codec")
	assert.Contains(jsonStr, "sample_fmt")
	assert.Contains(jsonStr, "sample_rate")
	assert.Contains(jsonStr, "48000")
	t.Logf("Audio context JSON: %s", jsonStr)
}

func Test_avcodec_profile_json(t *testing.T) {
	codec := AVCodec_find_encoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not available")
	}

	profiles := codec.Profiles()
	if profiles == nil || len(profiles) == 0 {
		t.Skip("No profiles available")
	}

	data, err := json.Marshal(profiles[0])
	if err == nil {
		t.Logf("Profile JSON: %s", string(data))
	}
}

func Test_avcodec_cap_json(t *testing.T) {
	assert := assert.New(t)

	cap := AV_CODEC_CAP_DR1 | AV_CODEC_CAP_DELAY
	data, err := json.Marshal(cap)
	assert.NoError(err)
	assert.NotEmpty(data)
	t.Logf("Capabilities JSON: %s", string(data))
}

func Test_avcodec_id_json(t *testing.T) {
	assert := assert.New(t)

	id := AV_CODEC_ID_H264
	data, err := json.Marshal(id)
	assert.NoError(err)
	assert.NotEmpty(data)
	t.Logf("Codec ID JSON: %s", string(data))
}

////////////////////////////////////////////////////////////////////////////////
// TEST STRING

func Test_avcodec_codec_string(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	assert.NotNil(codec)

	str := codec.String()
	assert.NotEmpty(str)
	t.Logf("Codec string:\n%s", str)
}

func Test_avcodec_context_string(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_encoder(AV_CODEC_ID_MPEG2VIDEO)
	if codec == nil {
		t.Skip("MPEG2VIDEO encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	ctx.SetWidth(1920)
	ctx.SetHeight(1080)
	ctx.SetPixFmt(AV_PIX_FMT_YUV420P)

	str := ctx.String()
	assert.NotEmpty(str)
	t.Logf("Context string:\n%s", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST CONSTANTS

func Test_avcodec_codec_constants(t *testing.T) {
	assert := assert.New(t)

	// Test codec IDs
	assert.NotEqual(AV_CODEC_ID_NONE, AV_CODEC_ID_H264)
	assert.NotEqual(AV_CODEC_ID_NONE, AV_CODEC_ID_MP2)

	// Test padding size
	assert.Greater(AV_INPUT_BUFFER_PADDING_SIZE, 0)
	t.Logf("Input buffer padding size: %d", AV_INPUT_BUFFER_PADDING_SIZE)
}

func Test_avcodec_codec_flags(t *testing.T) {
	assert := assert.New(t)

	// Test that flags have values
	assert.NotEqual(AVCodecFlag(0), AV_CODEC_FLAG_GLOBAL_HEADER)
	assert.NotEqual(AVCodecFlag(0), AV_CODEC_FLAG_LOW_DELAY)
	assert.NotEqual(AVCodecFlag2(0), AV_CODEC_FLAG2_FAST)
	assert.NotEqual(AVCodecFlag2(0), AV_CODEC_FLAG2_SHOW_ALL)
}

func Test_avcodec_codec_capabilities(t *testing.T) {
	assert := assert.New(t)

	// Test that capabilities have values
	assert.Equal(AVCodecCap(0), AV_CODEC_CAP_NONE)
	assert.NotEqual(AVCodecCap(0), AV_CODEC_CAP_DR1)
	assert.NotEqual(AVCodecCap(0), AV_CODEC_CAP_DELAY)
	assert.NotEqual(AVCodecCap(0), AV_CODEC_CAP_FRAME_THREADS)
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avcodec_codec_nil_checks(t *testing.T) {
	// These functions should handle nil gracefully and not crash
	var codec *AVCodec

	// Reading from nil codec may panic in C, so we just verify the test doesn't crash
	// In a real scenario, these would be checked before use
	t.Log("Testing nil codec handling")

	// Allocating with nil codec
	ctx := AVCodec_alloc_context(codec)
	if ctx != nil {
		AVCodec_free_context(ctx)
	}
}

func Test_avcodec_context_delay(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	delay := ctx.Delay()
	t.Logf("Context delay: %d", delay)
}

func Test_avcodec_context_coded_dimensions(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_decoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 decoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	codedWidth := ctx.CodedWidth()
	codedHeight := ctx.CodedHeight()
	t.Logf("Coded dimensions: %dx%d", codedWidth, codedHeight)
}

func Test_avcodec_context_frame_num(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_encoder(AV_CODEC_ID_MPEG2VIDEO)
	if codec == nil {
		t.Skip("MPEG2VIDEO encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	frameNum := ctx.FrameNum()
	assert.Equal(int64(0), frameNum)
	t.Logf("Frame num: %d", frameNum)
}

func Test_avcodec_context_frame_size(t *testing.T) {
	assert := assert.New(t)

	codec := AVCodec_find_encoder(AV_CODEC_ID_MP2)
	if codec == nil {
		t.Skip("MP2 encoder not available")
	}

	ctx := AVCodec_alloc_context(codec)
	assert.NotNil(ctx)
	defer AVCodec_free_context(ctx)

	frameSize := ctx.FrameSize()
	t.Logf("Frame size: %d", frameSize)
}

////////////////////////////////////////////////////////////////////////////////
// TEST IsEncoder and IsDecoder

func Test_avcodec_codec_is_encoder(t *testing.T) {
	assert := assert.New(t)

	// Test encoder
	encoder := AVCodec_find_encoder(AV_CODEC_ID_H264)
	if encoder == nil {
		t.Skip("H264 encoder not found")
	}

	assert.True(encoder.IsEncoder(), "H264 encoder should return true for IsEncoder()")
	assert.False(encoder.IsDecoder(), "H264 encoder should return false for IsDecoder()")
	t.Logf("H264 encoder: IsEncoder=%v, IsDecoder=%v", encoder.IsEncoder(), encoder.IsDecoder())

	// Test decoder
	decoder := AVCodec_find_decoder(AV_CODEC_ID_H264)
	assert.NotNil(decoder, "H264 decoder should exist")

	assert.False(decoder.IsEncoder(), "H264 decoder should return false for IsEncoder()")
	assert.True(decoder.IsDecoder(), "H264 decoder should return true for IsDecoder()")
	t.Logf("H264 decoder: IsEncoder=%v, IsDecoder=%v", decoder.IsEncoder(), decoder.IsDecoder())
}

func Test_avcodec_codec_is_decoder(t *testing.T) {
	assert := assert.New(t)

	// Test audio decoder
	decoder := AVCodec_find_decoder(AV_CODEC_ID_MP2)
	if decoder == nil {
		t.Skip("MP2 decoder not found")
	}

	assert.True(decoder.IsDecoder(), "MP2 decoder should return true for IsDecoder()")
	assert.False(decoder.IsEncoder(), "MP2 decoder should return false for IsEncoder()")
	t.Logf("MP2 decoder: IsEncoder=%v, IsDecoder=%v", decoder.IsEncoder(), decoder.IsDecoder())

	// Test audio encoder
	encoder := AVCodec_find_encoder(AV_CODEC_ID_MP2)
	if encoder == nil {
		t.Skip("MP2 encoder not found")
	}

	assert.True(encoder.IsEncoder(), "MP2 encoder should return true for IsEncoder()")
	assert.False(encoder.IsDecoder(), "MP2 encoder should return false for IsDecoder()")
	t.Logf("MP2 encoder: IsEncoder=%v, IsDecoder=%v", encoder.IsEncoder(), encoder.IsDecoder())
}

////////////////////////////////////////////////////////////////////////////////
// TEST PrivClass

func Test_avcodec_codec_priv_class(t *testing.T) {
	assert := assert.New(t)

	// Test codec with priv_class
	codec := AVCodec_find_encoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	class := codec.PrivClass()
	if class == nil {
		t.Skip("H264 encoder has no priv_class (may not be compiled with options)")
	}

	assert.NotNil(class, "H264 encoder should have priv_class")

	// The class should have a valid class name
	className := class.Name()
	assert.NotEmpty(className, "AVClass should have a name")
	t.Logf("H264 encoder priv_class name: %s", className)
}

func Test_avcodec_codec_priv_class_options(t *testing.T) {
	assert := assert.New(t)

	// Test that we can enumerate options via PrivClass
	codec := AVCodec_find_encoder(AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	class := codec.PrivClass()
	if class == nil {
		t.Skip("H264 encoder has no priv_class")
	}

	// Use FAKE_OBJ trick to enumerate options
	options := AVUtil_opt_list_from_class(class)
	assert.NotEmpty(options, "H264 encoder should have options")

	t.Logf("Found %d options for H264 encoder via PrivClass", len(options))

	// Check first few options
	for i := 0; i < min(5, len(options)); i++ {
		opt := options[i]
		assert.NotEmpty(opt.Name(), "Option should have a name")
		t.Logf("Option %d: %s (%v)", i, opt.Name(), opt.Type())
	}
}
