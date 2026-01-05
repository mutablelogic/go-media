package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_avcodec_core_000(t *testing.T) {
	assert := assert.New(t)

	// Iterate over all codecs
	var opaque uintptr
	for {
		codec := AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}

		t.Log("codec.name=", codec.Name())
		t.Log("  .longname=", codec.LongName())
		t.Log("  .type=", codec.Type())
		t.Log("  .id=", codec.ID())
		t.Log("  .encoder=", AVCodec_is_encoder(codec))
		t.Log("  .decoder=", AVCodec_is_decoder(codec))
		if AVCodec_is_encoder(codec) {
			codec_ := AVCodec_find_encoder(codec.ID())
			assert.NotNil(codec_)
		} else if AVCodec_is_decoder(codec) {
			codec_ := AVCodec_find_decoder(codec.ID())
			assert.NotNil(codec_)
		}
		if codec.Type().Is(AVMEDIA_TYPE_VIDEO) {
			t.Log("  .framerates=", codec.SupportedFramerates())
			t.Log("  .pixel_formats=", codec.PixelFormats())
		}
		if codec.Type().Is(AVMEDIA_TYPE_AUDIO) {
			t.Log("  .samplerates=", codec.SupportedSamplerates())
			t.Log("  .sample_formats=", codec.SampleFormats())
		}
		t.Log("  .profile=", codec.Profiles())
	}
}
