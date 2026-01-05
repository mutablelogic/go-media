package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME CREATION

func Test_frame_create_empty(t *testing.T) {
	assert := assert.New(t)

	frame, err := NewFrame(nil)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	assert.NotNil(frame)
	assert.Equal(media.UNKNOWN, frame.Type())
	t.Log(frame)
}

func Test_frame_create_audio(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	assert.NotNil(frame)
	assert.Equal(media.AUDIO, frame.Type())
	assert.Equal(44100, frame.SampleRate())
	assert.Equal(ff.AV_SAMPLE_FMT_FLTP, frame.SampleFormat())
	assert.Equal(2, frame.ChannelLayout().NumChannels())
	t.Log(frame)
}

func Test_frame_create_video(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1920x1080", 30)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	assert.NotNil(frame)
	assert.Equal(media.VIDEO, frame.Type())
	assert.Equal(1920, frame.Width())
	assert.Equal(1080, frame.Height())
	assert.Equal(ff.AV_PIX_FMT_YUV420P, frame.PixelFormat())
	t.Log(frame)
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME ALLOCATION

func Test_frame_allocate_audio(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	assert.False(frame.IsAllocated())

	// Set number of samples and allocate
	(*ff.AVFrame)(frame).SetNumSamples(1024)
	err = frame.AllocateBuffers()
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.True(frame.IsAllocated())
	assert.Equal(1024, frame.NumSamples())
	t.Log(frame)
}

func Test_frame_allocate_video(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "640x480", 25)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	assert.False(frame.IsAllocated())

	err = frame.AllocateBuffers()
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.True(frame.IsAllocated())
	t.Log(frame)
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME COPY

func Test_frame_copy_audio(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	(*ff.AVFrame)(frame).SetNumSamples(512)
	if err := frame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	copy, err := frame.Copy()
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer copy.Close()

	assert.Equal(frame.Type(), copy.Type())
	assert.Equal(frame.SampleRate(), copy.SampleRate())
	assert.Equal(frame.NumSamples(), copy.NumSamples())
	assert.True(copy.IsAllocated())
	t.Log("Original:", frame)
	t.Log("Copy:", copy)
}

func Test_frame_copy_video(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "320x240", 25)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	if err := frame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	copy, err := frame.Copy()
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer copy.Close()

	assert.Equal(frame.Type(), copy.Type())
	assert.Equal(frame.Width(), copy.Width())
	assert.Equal(frame.Height(), copy.Height())
	assert.True(copy.IsAllocated())
	t.Log("Original:", frame)
	t.Log("Copy:", copy)
}

func Test_frame_copy_nil(t *testing.T) {
	assert := assert.New(t)

	var frame *Frame
	_, err := frame.Copy()
	assert.Error(err)
	t.Log("Expected error:", err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME TYPE DETECTION

func Test_frame_type_audio(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("s16", "mono", 22050)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	assert.Equal(media.AUDIO, frame.Type())
}

func Test_frame_type_video(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("rgb24", "800x600", 60)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	assert.Equal(media.VIDEO, frame.Type())
}

func Test_frame_type_nil(t *testing.T) {
	assert := assert.New(t)

	var frame *Frame
	assert.Equal(media.UNKNOWN, frame.Type())
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME TIME PARAMETERS

func Test_frame_pts(t *testing.T) {
	assert := assert.New(t)

	frame, err := NewFrame(nil)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Initial PTS should be undefined
	assert.Equal(int64(ff.AV_NOPTS_VALUE), frame.Pts())

	// Set PTS
	frame.SetPts(100)
	assert.Equal(int64(100), frame.Pts())

	// Increment PTS
	frame.IncPts(50)
	assert.Equal(int64(150), frame.Pts())

	t.Logf("PTS: %d", frame.Pts())
}

func Test_frame_timestamp(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1920x1080", 25)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Initial timestamp should be undefined
	assert.Equal(TS_UNDEFINED, frame.Ts())

	// Set timestamp in seconds
	frame.SetTs(1.5)
	assert.InDelta(1.5, frame.Ts(), 0.1)

	// Set PTS to 0 should give 0 timestamp
	frame.SetPts(0)
	assert.InDelta(0.0, frame.Ts(), 0.001)

	t.Logf("Timestamp: %f", frame.Ts())
}

func Test_frame_timebase(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1920x1080", 30)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	tb := frame.TimeBase()
	assert.NotEqual(0, tb.Num())
	assert.NotEqual(0, tb.Den())

	// Framerate should be inverse of timebase
	framerate := ff.AVUtil_rational_q2d(ff.AVUtil_rational_invert(tb))
	assert.InDelta(30.0, framerate, 0.1)

	t.Logf("Timebase: %d/%d (framerate: %f)", tb.Num(), tb.Den(), framerate)
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME UNREF

func Test_frame_unref(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "640x480", 25)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	if err := frame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	assert.True(frame.IsAllocated())

	frame.Unref()
	assert.False(frame.IsAllocated())
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME MAKE WRITABLE

func Test_frame_make_writable(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "320x240", 25)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	if err := frame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	err = frame.MakeWritable()
	assert.NoError(err)
	t.Log("Frame is writable")
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME DATA ACCESS

func Test_frame_bytes(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "320x240", 25)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	if err := frame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	// YUV420P has 3 planes
	plane0 := frame.Bytes(0)
	assert.NotNil(plane0)
	assert.NotEmpty(plane0)

	stride0 := frame.Stride(0)
	assert.Greater(stride0, 0)

	t.Logf("Plane 0: %d bytes, stride: %d", len(plane0), stride0)
}

func Test_frame_float32(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	(*ff.AVFrame)(frame).SetNumSamples(256)
	if err := frame.AllocateBuffers(); !assert.NoError(err) {
		t.FailNow()
	}

	// FLTP has planar float samples
	plane0 := frame.Float32(0)
	assert.NotNil(plane0)
	assert.Equal(256, len(plane0))

	t.Logf("Plane 0: %d samples", len(plane0))
}

func Test_frame_set_float32(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "mono", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Set samples
	samples := make([]float32, 128)
	for i := range samples {
		samples[i] = float32(i) * 0.1
	}

	err = frame.SetFloat32(0, samples)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.Equal(128, frame.NumSamples())

	// Verify data was set
	plane0 := frame.Float32(0)
	assert.Equal(len(samples), len(plane0))
	assert.InDelta(samples[0], plane0[0], 0.001)
	assert.InDelta(samples[127], plane0[127], 0.001)

	t.Logf("Set %d samples", len(samples))
}

////////////////////////////////////////////////////////////////////////////////
// TEST FRAME MATCH FORMAT

func Test_frame_matches_format_audio(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame1, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame1.Close()

	frame2, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame2.Close()

	// Same format should match
	assert.True(frame1.MatchesFormat(frame2))
	assert.True(frame2.MatchesFormat(frame1))
}

func Test_frame_matches_format_audio_different(t *testing.T) {
	assert := assert.New(t)

	par1, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	par2, err := NewAudioPar("s16", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame1, err := NewFrame(par1)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame1.Close()

	frame2, err := NewFrame(par2)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame2.Close()

	// Different sample format should not match
	assert.False(frame1.MatchesFormat(frame2))
}

func Test_frame_matches_format_video(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1920x1080", 30)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame1, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame1.Close()

	frame2, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame2.Close()

	// Same format should match
	assert.True(frame1.MatchesFormat(frame2))
}

func Test_frame_matches_format_video_different_size(t *testing.T) {
	assert := assert.New(t)

	par1, err := NewVideoPar("yuv420p", "1920x1080", 30)
	if !assert.NoError(err) {
		t.FailNow()
	}

	par2, err := NewVideoPar("yuv420p", "1280x720", 30)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame1, err := NewFrame(par1)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame1.Close()

	frame2, err := NewFrame(par2)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame2.Close()

	// Different size should not match
	assert.False(frame1.MatchesFormat(frame2))
}

func Test_frame_matches_format_different_types(t *testing.T) {
	assert := assert.New(t)

	parAudio, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	parVideo, err := NewVideoPar("yuv420p", "1920x1080", 30)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frameAudio, err := NewFrame(parAudio)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frameAudio.Close()

	frameVideo, err := NewFrame(parVideo)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frameVideo.Close()

	// Different types should not match
	assert.False(frameAudio.MatchesFormat(frameVideo))
	assert.False(frameVideo.MatchesFormat(frameAudio))
}

func Test_frame_matches_format_nil(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	var nilFrame *Frame
	assert.False(frame.MatchesFormat(nilFrame))
	assert.False(nilFrame.MatchesFormat(frame))
	assert.False(nilFrame.MatchesFormat(nilFrame))
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_frame_json(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1920x1080", 30)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	data, err := json.Marshal((*ff.AVFrame)(frame))
	assert.NoError(err)
	assert.NotEmpty(data)

	t.Logf("JSON: %s", string(data))
}

////////////////////////////////////////////////////////////////////////////////
// TEST COPY PROPS

func Test_frame_copy_props(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "640x480", 25)
	if !assert.NoError(err) {
		t.FailNow()
	}

	frame1, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame1.Close()

	frame1.SetPts(1000)

	frame2, err := NewFrame(par)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame2.Close()

	err = frame2.CopyPropsFromFrame(frame1)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.Equal(frame1.Pts(), frame2.Pts())
	t.Logf("Copied PTS: %d", frame2.Pts())
}
