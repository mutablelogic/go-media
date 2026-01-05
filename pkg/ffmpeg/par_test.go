package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	media "github.com/mutablelogic/go-media"
	assert "github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST AUDIO PARAMETERS

func Test_par_audio_create(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "mono", 22050)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotNil(par)
	assert.Equal(media.AUDIO, par.Type())
	assert.Equal(22050, par.SampleRate())
	t.Log(par)
}

func Test_par_audio_create_stereo(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("s16", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotNil(par)
	assert.Equal(media.AUDIO, par.Type())
	assert.Equal(44100, par.SampleRate())
	assert.Equal(2, par.ChannelLayout().NumChannels())
	t.Log(par)
}

func Test_par_audio_invalid_format(t *testing.T) {
	assert := assert.New(t)

	_, err := NewAudioPar("invalid_format", "mono", 22050)
	assert.Error(err)
	t.Log("Expected error:", err)
}

func Test_par_audio_invalid_layout(t *testing.T) {
	assert := assert.New(t)

	_, err := NewAudioPar("fltp", "invalid_layout", 22050)
	assert.Error(err)
	t.Log("Expected error:", err)
}

func Test_par_audio_invalid_samplerate(t *testing.T) {
	assert := assert.New(t)

	_, err := NewAudioPar("fltp", "mono", 0)
	assert.Error(err)
	t.Log("Expected error:", err)

	_, err = NewAudioPar("fltp", "mono", -100)
	assert.Error(err)
	t.Log("Expected error:", err)
}

func Test_par_audio_panic_constructor(t *testing.T) {
	assert := assert.New(t)

	// Valid parameters should not panic
	assert.NotPanics(func() {
		par := AudioPar("fltp", "mono", 22050)
		assert.NotNil(par)
	})

	// Invalid parameters should panic
	assert.Panics(func() {
		AudioPar("invalid_format", "mono", 22050)
	})
}

////////////////////////////////////////////////////////////////////////////////
// TEST VIDEO PARAMETERS

func Test_par_video_create(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1280x720", 25)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotNil(par)
	assert.Equal(media.VIDEO, par.Type())
	assert.Equal(1280, par.Width())
	assert.Equal(720, par.Height())
	assert.Equal("1280x720", par.WidthHeight())
	assert.InDelta(25.0, par.FrameRate(), 0.01)
	t.Log(par)
}

func Test_par_video_create_various_sizes(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		size   string
		width  int
		height int
	}{
		{"640x480", 640, 480},
		{"1920x1080", 1920, 1080},
		{"3840x2160", 3840, 2160},
		{"720x576", 720, 576},
	}

	for _, tc := range testCases {
		par, err := NewVideoPar("yuv420p", tc.size, 30)
		if !assert.NoError(err, "Failed for size %s", tc.size) {
			continue
		}
		assert.Equal(tc.width, par.Width())
		assert.Equal(tc.height, par.Height())
		t.Logf("%s -> %dx%d", tc.size, par.Width(), par.Height())
	}
}

func Test_par_video_zero_framerate(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1280x720", 0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotNil(par)
	assert.Equal(0.0, par.FrameRate())
	t.Log(par)
}

func Test_par_video_invalid_pixelformat(t *testing.T) {
	assert := assert.New(t)

	_, err := NewVideoPar("invalid_format", "1280x720", 25)
	assert.Error(err)
	t.Log("Expected error:", err)
}

func Test_par_video_invalid_size(t *testing.T) {
	assert := assert.New(t)

	invalidSizes := []string{
		"invalid",
		"1280",
		"1280x",
		"x720",
		"0x0",
		"-100x720",
	}

	for _, size := range invalidSizes {
		_, err := NewVideoPar("yuv420p", size, 25)
		assert.Error(err, "Expected error for size: %s", size)
		t.Logf("Size %q error: %v", size, err)
	}
}

func Test_par_video_invalid_framerate(t *testing.T) {
	assert := assert.New(t)

	_, err := NewVideoPar("yuv420p", "1280x720", -1)
	assert.Error(err)
	t.Log("Expected error:", err)
}

func Test_par_video_panic_constructor(t *testing.T) {
	assert := assert.New(t)

	// Valid parameters should not panic
	assert.NotPanics(func() {
		par := VideoPar("yuv420p", "1280x720", 25)
		assert.NotNil(par)
	})

	// Invalid parameters should panic
	assert.Panics(func() {
		VideoPar("invalid_format", "1280x720", 25)
	})
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_par_audio_json(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "stereo", 48000)
	if !assert.NoError(err) {
		t.FailNow()
	}

	data, err := json.Marshal(par)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotEmpty(data)
	t.Logf("JSON: %s", string(data))
}

func Test_par_video_json(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1920x1080", 30)
	if !assert.NoError(err) {
		t.FailNow()
	}

	data, err := json.Marshal(par)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotEmpty(data)
	t.Logf("JSON: %s", string(data))
}

func Test_par_nil_json(t *testing.T) {
	assert := assert.New(t)

	var par *Par
	data, err := json.Marshal(par)
	assert.NoError(err)
	assert.Equal("null", string(data))
}

////////////////////////////////////////////////////////////////////////////////
// TEST NIL HANDLING

func Test_par_nil_type(t *testing.T) {
	assert := assert.New(t)

	var par *Par
	assert.Equal(media.UNKNOWN, par.Type())
}

func Test_par_nil_widthheight(t *testing.T) {
	assert := assert.New(t)

	var par *Par
	assert.Equal("0x0", par.WidthHeight())
}

func Test_par_nil_framerate(t *testing.T) {
	assert := assert.New(t)

	var par *Par
	assert.Equal(0.0, par.FrameRate())
}

func Test_par_nil_string(t *testing.T) {
	assert := assert.New(t)

	var par *Par
	assert.Equal("<nil>", par.String())
}

func Test_par_nil_validate(t *testing.T) {
	assert := assert.New(t)

	var par *Par
	err := par.ValidateFromCodec(nil)
	assert.Error(err)
	t.Log("Expected error:", err)
}

func Test_par_nil_copy(t *testing.T) {
	assert := assert.New(t)

	var par *Par
	err := par.CopyToCodecContext(nil)
	assert.Error(err)
	t.Log("Expected error:", err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST METADATA OPTIONS

func Test_par_audio_with_metadata(t *testing.T) {
	assert := assert.New(t)

	opts := []media.Metadata{
		NewMetadata("bitrate", "128000"),
		NewMetadata("profile", "aac_low"),
	}

	par, err := NewAudioPar("fltp", "stereo", 44100, opts...)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotNil(par)
	t.Log(par)
}

func Test_par_video_with_metadata(t *testing.T) {
	assert := assert.New(t)

	opts := []media.Metadata{
		NewMetadata("bitrate", "2000000"),
		NewMetadata("preset", "fast"),
	}

	par, err := NewVideoPar("yuv420p", "1920x1080", 30, opts...)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotNil(par)
	t.Log(par)
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_par_audio_high_samplerate(t *testing.T) {
	assert := assert.New(t)

	par, err := NewAudioPar("fltp", "7.1", 192000)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotNil(par)
	assert.Equal(192000, par.SampleRate())
	assert.Equal(8, par.ChannelLayout().NumChannels())
	t.Log(par)
}

func Test_par_video_high_framerate(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1920x1080", 120)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotNil(par)
	assert.InDelta(120.0, par.FrameRate(), 0.01)
	t.Log(par)
}

func Test_par_video_fractional_framerate(t *testing.T) {
	assert := assert.New(t)

	par, err := NewVideoPar("yuv420p", "1920x1080", 29.97)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.NotNil(par)
	assert.InDelta(29.97, par.FrameRate(), 0.01)
	t.Log("Framerate:", par.FrameRate())
}
