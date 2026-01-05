package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME ALLOCATION

func Test_avutil_frame_alloc(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame, "Frame allocation should succeed")

	// Verify frame is not allocated yet
	assert.False(AVUtil_frame_is_allocated(frame))

	AVUtil_frame_free(frame)
}

func Test_avutil_frame_free_nil(t *testing.T) {
	// Should not crash with nil frame
	var frame *AVFrame
	AVUtil_frame_free(frame)
}

func Test_avutil_frame_multiple_alloc_free(t *testing.T) {
	assert := assert.New(t)

	// Allocate and free multiple frames
	for i := 0; i < 100; i++ {
		frame := AVUtil_frame_alloc()
		assert.NotNil(frame)
		AVUtil_frame_free(frame)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST VIDEO FRAME BUFFER ALLOCATION

func Test_avutil_frame_get_buffer_video(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Set up video frame parameters
	frame.SetWidth(1920)
	frame.SetHeight(1080)
	frame.SetPixFmt(AV_PIX_FMT_YUV420P)

	// Allocate buffer
	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err, "Buffer allocation should succeed")
	assert.True(AVUtil_frame_is_allocated(frame), "Frame should be allocated")

	// Verify properties
	assert.Equal(1920, frame.Width())
	assert.Equal(1080, frame.Height())
	assert.Equal(AV_PIX_FMT_YUV420P, frame.PixFmt())

	// Verify planes
	numPlanes := AVUtil_frame_get_num_planes(frame)
	assert.Equal(3, numPlanes, "YUV420P should have 3 planes")
}

func Test_avutil_frame_get_buffer_video_formats(t *testing.T) {
	assert := assert.New(t)

	formats := []struct {
		format      AVPixelFormat
		width       int
		height      int
		description string
	}{
		{AV_PIX_FMT_RGB24, 640, 480, "RGB24"},
		{AV_PIX_FMT_RGBA, 640, 480, "RGBA"},
		{AV_PIX_FMT_YUV420P, 1280, 720, "YUV420P"},
		{AV_PIX_FMT_YUV422P, 1920, 1080, "YUV422P"},
		{AV_PIX_FMT_GRAY8, 320, 240, "GRAY8"},
	}

	for _, tc := range formats {
		frame := AVUtil_frame_alloc()
		assert.NotNil(frame)

		frame.SetWidth(tc.width)
		frame.SetHeight(tc.height)
		frame.SetPixFmt(tc.format)

		err := AVUtil_frame_get_buffer(frame, false)
		assert.NoError(err, "Buffer allocation should succeed for %s", tc.description)
		assert.True(AVUtil_frame_is_allocated(frame))

		t.Logf("Format %s (%dx%d): %d planes", tc.description, tc.width, tc.height,
			AVUtil_frame_get_num_planes(frame))

		AVUtil_frame_free(frame)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST AUDIO FRAME BUFFER ALLOCATION

func Test_avutil_frame_get_buffer_audio(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Set up audio frame parameters
	frame.SetSampleFormat(AV_SAMPLE_FMT_S16)
	frame.SetNumSamples(1024)
	frame.SetSampleRate(48000)

	// Set stereo channel layout
	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := frame.SetChannelLayout(layout)
	assert.NoError(err)

	// Allocate buffer
	err = AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err, "Buffer allocation should succeed")
	assert.True(AVUtil_frame_is_allocated(frame))

	// Verify properties
	assert.Equal(1024, frame.NumSamples())
	assert.Equal(AV_SAMPLE_FMT_S16, frame.SampleFormat())
	assert.Equal(48000, frame.SampleRate())
}

func Test_avutil_frame_get_buffer_audio_formats(t *testing.T) {
	assert := assert.New(t)

	formats := []struct {
		format      AVSampleFormat
		numSamples  int
		sampleRate  int
		channels    int
		description string
	}{
		{AV_SAMPLE_FMT_U8, 512, 44100, 2, "U8 Stereo"},
		{AV_SAMPLE_FMT_S16, 1024, 48000, 2, "S16 Stereo"},
		{AV_SAMPLE_FMT_S32, 2048, 48000, 2, "S32 Stereo"},
		{AV_SAMPLE_FMT_FLT, 1024, 48000, 2, "Float Stereo"},
		{AV_SAMPLE_FMT_FLTP, 1024, 48000, 2, "Float Planar Stereo"},
		{AV_SAMPLE_FMT_S16P, 1024, 48000, 6, "S16 Planar 5.1"},
	}

	for _, tc := range formats {
		frame := AVUtil_frame_alloc()
		assert.NotNil(frame)

		frame.SetSampleFormat(tc.format)
		frame.SetNumSamples(tc.numSamples)
		frame.SetSampleRate(tc.sampleRate)

		var layout AVChannelLayout
		AVUtil_channel_layout_default(&layout, tc.channels)
		err := frame.SetChannelLayout(layout)
		assert.NoError(err)

		err = AVUtil_frame_get_buffer(frame, false)
		assert.NoError(err, "Buffer allocation should succeed for %s", tc.description)
		assert.True(AVUtil_frame_is_allocated(frame))

		numPlanes := AVUtil_frame_get_num_planes(frame)
		t.Logf("Format %s: %d samples, %d Hz, %d channels, %d planes",
			tc.description, tc.numSamples, tc.sampleRate, tc.channels, numPlanes)

		AVUtil_frame_free(frame)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME PROPERTIES

func Test_avutil_frame_video_properties(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Set and get width
	frame.SetWidth(1920)
	assert.Equal(1920, frame.Width())

	// Set and get height
	frame.SetHeight(1080)
	assert.Equal(1080, frame.Height())

	// Set and get pixel format
	frame.SetPixFmt(AV_PIX_FMT_YUV420P)
	assert.Equal(AV_PIX_FMT_YUV420P, frame.PixFmt())

	// Set and get PTS
	frame.SetPts(12345)
	assert.Equal(int64(12345), frame.Pts())

	// Set and get time base
	tb := AVRational{num: 1, den: 1000}
	frame.SetTimeBase(tb)
	retrievedTb := frame.TimeBase()
	assert.Equal(tb.num, retrievedTb.num)
	assert.Equal(tb.den, retrievedTb.den)

	// Set and get sample aspect ratio
	sar := AVRational{num: 1, den: 1}
	frame.SetSampleAspectRatio(sar)
	retrievedSar := frame.SampleAspectRatio()
	assert.Equal(sar.num, retrievedSar.num)
	assert.Equal(sar.den, retrievedSar.den)
}

func Test_avutil_frame_audio_properties(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Set and get sample format
	frame.SetSampleFormat(AV_SAMPLE_FMT_S16)
	assert.Equal(AV_SAMPLE_FMT_S16, frame.SampleFormat())

	// Set and get number of samples
	frame.SetNumSamples(1024)
	assert.Equal(1024, frame.NumSamples())

	// Set and get sample rate
	frame.SetSampleRate(48000)
	assert.Equal(48000, frame.SampleRate())

	// Set and get channel layout
	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := frame.SetChannelLayout(layout)
	assert.NoError(err)
	retrievedLayout := frame.ChannelLayout()
	assert.Equal(2, retrievedLayout.NumChannels())
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME UNREF

func Test_avutil_frame_unref(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Allocate buffer
	frame.SetWidth(640)
	frame.SetHeight(480)
	frame.SetPixFmt(AV_PIX_FMT_RGB24)
	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)
	assert.True(AVUtil_frame_is_allocated(frame))

	// Unref should deallocate
	AVUtil_frame_unref(frame)
	assert.False(AVUtil_frame_is_allocated(frame))
}

func Test_avutil_frame_multiple_unref(t *testing.T) {
	frame := AVUtil_frame_alloc()
	defer AVUtil_frame_free(frame)

	// Multiple unref calls should be safe
	AVUtil_frame_unref(frame)
	AVUtil_frame_unref(frame)
	AVUtil_frame_unref(frame)
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME COPY

func Test_avutil_frame_copy_video(t *testing.T) {
	assert := assert.New(t)

	src := AVUtil_frame_alloc()
	assert.NotNil(src)
	defer AVUtil_frame_free(src)

	dst := AVUtil_frame_alloc()
	assert.NotNil(dst)
	defer AVUtil_frame_free(dst)

	// Allocate source frame
	src.SetWidth(320)
	src.SetHeight(240)
	src.SetPixFmt(AV_PIX_FMT_RGB24)
	err := AVUtil_frame_get_buffer(src, false)
	assert.NoError(err)

	// Allocate destination frame with same parameters
	dst.SetWidth(320)
	dst.SetHeight(240)
	dst.SetPixFmt(AV_PIX_FMT_RGB24)
	err = AVUtil_frame_get_buffer(dst, false)
	assert.NoError(err)

	// Copy frame data
	err = AVUtil_frame_copy(dst, src)
	assert.NoError(err)

	// Verify dimensions match
	assert.Equal(src.Width(), dst.Width())
	assert.Equal(src.Height(), dst.Height())
	assert.Equal(src.PixFmt(), dst.PixFmt())
}

func Test_avutil_frame_copy_props(t *testing.T) {
	assert := assert.New(t)

	src := AVUtil_frame_alloc()
	assert.NotNil(src)
	defer AVUtil_frame_free(src)

	dst := AVUtil_frame_alloc()
	assert.NotNil(dst)
	defer AVUtil_frame_free(dst)

	// Set properties on source
	src.SetWidth(1920)
	src.SetHeight(1080)
	src.SetPixFmt(AV_PIX_FMT_YUV420P)
	src.SetPts(12345)
	tb := AVRational{num: 1, den: 1000}
	src.SetTimeBase(tb)

	// Copy properties
	err := AVUtil_frame_copy_props(dst, src)
	assert.NoError(err)

	// Verify properties copied (except dimensions which affect data layout)
	assert.Equal(src.Pts(), dst.Pts())
	assert.Equal(src.TimeBase().num, dst.TimeBase().num)
	assert.Equal(src.TimeBase().den, dst.TimeBase().den)
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME MAKE WRITABLE

func Test_avutil_frame_make_writable(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Allocate buffer
	frame.SetWidth(320)
	frame.SetHeight(240)
	frame.SetPixFmt(AV_PIX_FMT_RGB24)
	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Make writable should succeed
	err = AVUtil_frame_make_writable(frame)
	assert.NoError(err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST PLANE OPERATIONS

func Test_avutil_frame_linesize(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	frame.SetWidth(640)
	frame.SetHeight(480)
	frame.SetPixFmt(AV_PIX_FMT_RGB24)
	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Get linesize for plane 0
	linesize := frame.Linesize(0)
	assert.Greater(linesize, 0, "Linesize should be positive")
	assert.GreaterOrEqual(linesize, 640*3, "RGB24 linesize should be at least width*3")

	t.Logf("RGB24 640x480 linesize: %d", linesize)
}

func Test_avutil_frame_planesize_video(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	frame.SetWidth(640)
	frame.SetHeight(480)
	frame.SetPixFmt(AV_PIX_FMT_YUV420P)
	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	numPlanes := AVUtil_frame_get_num_planes(frame)
	assert.Equal(3, numPlanes)

	for i := 0; i < numPlanes; i++ {
		planesize := frame.Planesize(i)
		assert.Greater(planesize, 0, "Plane %d size should be positive", i)
		t.Logf("YUV420P plane %d size: %d bytes", i, planesize)
	}
}

func Test_avutil_frame_planesize_audio_packed(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Packed audio format
	frame.SetSampleFormat(AV_SAMPLE_FMT_S16)
	frame.SetNumSamples(1024)
	frame.SetSampleRate(48000)

	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := frame.SetChannelLayout(layout)
	assert.NoError(err)

	err = AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Packed format should have 1 plane
	numPlanes := AVUtil_frame_get_num_planes(frame)
	assert.Equal(1, numPlanes)

	// Plane 0 should contain all channels
	planesize := frame.Planesize(0)
	expectedSize := 2 * 1024 * 2 // 2 bytes per sample * 1024 samples * 2 channels
	assert.Equal(expectedSize, planesize)
	t.Logf("Packed S16 stereo plane size: %d bytes (expected %d)", planesize, expectedSize)

	// Other planes should be 0
	assert.Equal(0, frame.Planesize(1))
}

func Test_avutil_frame_planesize_audio_planar(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Planar audio format
	frame.SetSampleFormat(AV_SAMPLE_FMT_FLTP)
	frame.SetNumSamples(1024)
	frame.SetSampleRate(48000)

	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := frame.SetChannelLayout(layout)
	assert.NoError(err)

	err = AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Planar format should have one plane per channel
	numPlanes := AVUtil_frame_get_num_planes(frame)
	assert.Equal(2, numPlanes)

	// Each plane should contain one channel
	expectedSize := 4 * 1024 // 4 bytes per float sample * 1024 samples
	for i := 0; i < numPlanes; i++ {
		planesize := frame.Planesize(i)
		assert.Equal(expectedSize, planesize, "Plane %d size mismatch", i)
		t.Logf("Planar FLTP channel %d plane size: %d bytes (expected %d)", i, planesize, expectedSize)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST DATA ACCESS

func Test_avutil_frame_bytes_video(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	frame.SetWidth(320)
	frame.SetHeight(240)
	frame.SetPixFmt(AV_PIX_FMT_RGB24)
	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Get bytes for plane 0
	data := frame.Bytes(0)
	assert.NotNil(data)
	assert.Greater(len(data), 0)
	t.Logf("RGB24 320x240 plane 0: %d bytes", len(data))

	// Verify we can write to the data
	if len(data) > 0 {
		data[0] = 255
		assert.Equal(uint8(255), data[0])
	}
}

func Test_avutil_frame_typed_access(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	frame.SetWidth(320)
	frame.SetHeight(240)
	frame.SetPixFmt(AV_PIX_FMT_RGB24)
	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Test different typed access methods
	uint8Data := frame.Uint8(0)
	assert.NotNil(uint8Data)

	uint16Data := frame.Uint16(0)
	assert.NotNil(uint16Data)

	int16Data := frame.Int16(0)
	assert.NotNil(int16Data)
}

func Test_avutil_frame_data(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	frame.SetWidth(320)
	frame.SetHeight(240)
	frame.SetPixFmt(AV_PIX_FMT_YUV420P)
	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Get all planes and strides
	planes, strides := frame.Data()
	assert.NotNil(planes)
	assert.NotNil(strides)
	assert.Len(planes, 8) // AV_NUM_DATA_POINTERS
	assert.Len(strides, 8)

	// First 3 planes should have data for YUV420P
	for i := 0; i < 3; i++ {
		assert.NotNil(planes[i], "Plane %d should have data", i)
		assert.Greater(strides[i], 0, "Plane %d stride should be positive", i)
		t.Logf("Plane %d: %d bytes, stride %d", i, len(planes[i]), strides[i])
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avutil_frame_json_video(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	frame.SetWidth(1920)
	frame.SetHeight(1080)
	frame.SetPixFmt(AV_PIX_FMT_YUV420P)
	frame.SetPts(12345)

	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Marshal to JSON
	data, err := json.Marshal(frame)
	assert.NoError(err)
	assert.NotEmpty(data)

	// Verify JSON contains expected fields
	jsonStr := string(data)
	assert.Contains(jsonStr, "pixel_format")
	assert.Contains(jsonStr, "width")
	assert.Contains(jsonStr, "height")
	assert.Contains(jsonStr, "1920")
	assert.Contains(jsonStr, "1080")

	t.Logf("Video frame JSON: %s", jsonStr)
}

func Test_avutil_frame_json_audio(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	frame.SetSampleFormat(AV_SAMPLE_FMT_S16)
	frame.SetNumSamples(1024)
	frame.SetSampleRate(48000)

	var layout AVChannelLayout
	AVUtil_channel_layout_default(&layout, 2)
	err := frame.SetChannelLayout(layout)
	assert.NoError(err)

	err = AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	// Marshal to JSON
	data, err := json.Marshal(frame)
	assert.NoError(err)
	assert.NotEmpty(data)

	// Verify JSON contains expected fields
	jsonStr := string(data)
	assert.Contains(jsonStr, "sample_format")
	assert.Contains(jsonStr, "num_samples")
	assert.Contains(jsonStr, "sample_rate")
	assert.Contains(jsonStr, "1024")
	assert.Contains(jsonStr, "48000")

	t.Logf("Audio frame JSON: %s", jsonStr)
}

func Test_avutil_frame_string(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	frame.SetWidth(640)
	frame.SetHeight(480)
	frame.SetPixFmt(AV_PIX_FMT_RGB24)

	err := AVUtil_frame_get_buffer(frame, false)
	assert.NoError(err)

	str := frame.String()
	assert.NotEmpty(str)
	t.Logf("Frame string:\n%s", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avutil_frame_invalid_plane_access(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Negative plane index
	assert.Equal(0, frame.Linesize(-1))
	assert.Equal(0, frame.Planesize(-1))
	assert.Nil(frame.Bytes(-1))

	// Large plane index
	assert.Equal(0, frame.Linesize(100))
	assert.Equal(0, frame.Planesize(100))
}

func Test_avutil_frame_unallocated_access(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Accessing unallocated frame should return nil
	assert.Nil(frame.Bytes(0))
	assert.Nil(frame.Uint8(0))
	assert.Nil(frame.Int16(0))
	assert.Nil(frame.Float32(0))
}

func Test_avutil_frame_zero_dimensions(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	assert.NotNil(frame)
	defer AVUtil_frame_free(frame)

	// Try to allocate with zero dimensions (should fail)
	frame.SetWidth(0)
	frame.SetHeight(0)
	frame.SetPixFmt(AV_PIX_FMT_RGB24)

	err := AVUtil_frame_get_buffer(frame, false)
	assert.Error(err, "Should fail with zero dimensions")
}

func Test_avutil_frame_get_num_planes_unallocated(t *testing.T) {
	frame := AVUtil_frame_alloc()
	defer AVUtil_frame_free(frame)

	// Unallocated frame should return 0 planes
	numPlanes := AVUtil_frame_get_num_planes(frame)
	// Result depends on frame state, just verify it doesn't crash
	t.Logf("Unallocated frame num planes: %d", numPlanes)
}
