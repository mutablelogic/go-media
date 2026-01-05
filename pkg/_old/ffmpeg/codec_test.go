package ffmpeg_test

import (
	"testing"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_codec_001(t *testing.T) {
	assert := assert.New(t)

	manager, err := ffmpeg.NewManager()
	if !assert.NoError(err) {
		t.FailNow()
	}

	for _, codec := range manager.Codecs(media.ANY) {
		t.Logf("%v", codec)
	}
}

func Test_codec_002(t *testing.T) {
	assert := assert.New(t)

	manager, err := ffmpeg.NewManager()
	if !assert.NoError(err) {
		t.FailNow()
	}

	for _, codec := range manager.Codecs(media.VIDEO) {
		assert.Equal(media.VIDEO, codec.Any().(*ffmpeg.Codec).Type())
		t.Logf("%v", codec)
	}

	for _, codec := range manager.Codecs(media.AUDIO) {
		assert.Equal(media.AUDIO, codec.Any().(*ffmpeg.Codec).Type())
		t.Logf("%v", codec)
	}

	for _, codec := range manager.Codecs(media.SUBTITLE) {
		assert.Equal(media.SUBTITLE, codec.Any().(*ffmpeg.Codec).Type())
		t.Logf("%v", codec)
	}
}

func Test_codec_003(t *testing.T) {
	assert := assert.New(t)

	manager, err := ffmpeg.NewManager()
	if !assert.NoError(err) {
		t.FailNow()
	}

	for _, codec := range manager.Codecs(media.ANY, "h264") {
		assert.Equal("h264", codec.Any().(*ffmpeg.Codec).Name())
		t.Logf("%v", codec)
	}
}

func Test_codec_004(t *testing.T) {
	assert := assert.New(t)

	manager, err := ffmpeg.NewManager()
	if !assert.NoError(err) {
		t.FailNow()
	}

	for _, meta := range manager.Codecs(media.ANY) {
		codec := meta.Any().(*ffmpeg.Codec)
		t.Logf("NAME %q", codec.Name())
		t.Logf("  TYPE           %q", codec.Type())
		t.Logf("  DESCRIPTION    %q", codec.Description())
		switch codec.Type() {
		case media.VIDEO:
			t.Logf("  PIXEL FORMATS  %q", codec.PixelFormats())
			t.Logf("  PROFILES       %q", codec.Profiles())
		case media.AUDIO:
			t.Logf("  SAMPLE FORMATS %q", codec.SampleFormats())
			t.Logf("  SAMPLE RATES   %q", codec.SampleRates())
			t.Logf("  CH LAYOUTS     %q", codec.ChannelLayouts())
		}
	}
}
