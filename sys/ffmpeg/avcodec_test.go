package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/djthorpe/go-media/sys/ffmpeg"
)

////////////////////////////////////////////////////////////////////////////////
// TEST ENUMS

func Test_avcodec_000(t *testing.T) {
	t.Log("Test_avcodec_000")
}

func Test_avcodec_001(t *testing.T) {
	if codec := FindDecoderById(AV_CODEC_ID_NONE); codec != nil {
		t.Error("Unexpected codec returned")
	}
}
func Test_avcodec_002(t *testing.T) {
	if codec := FindDecoderById(AV_CODEC_ID_H265); codec == nil {
		t.Error("Unable to find codec")
	} else {
		t.Log(codec)
	}
}
func Test_avcodec_003(t *testing.T) {
	if codec := FindDecoderByName("hevc"); codec == nil {
		t.Error("Unable to find codec")
	} else {
		t.Log(codec)
	}
}
func Test_avcodec_004(t *testing.T) {
	if codec := FindEncoderById(AV_CODEC_ID_NONE); codec != nil {
		t.Error("Unexpected codec returned")
	}
}
func Test_avcodec_005(t *testing.T) {
	if codec := FindEncoderById(AV_CODEC_ID_H264); codec == nil {
		t.Error("Unable to find codec")
	} else if codec := FindEncoderByName(codec.Name()); codec == nil {
		t.Error("Unable to find codec")
	} else {
		t.Log(codec)
	}
}

func Test_avcodec_006(t *testing.T) {
	if params := NewAVCodecParameters(); params == nil {
		t.Error("Unexpected nil value returned")
	} else {
		t.Log(params)
		params.Free()
	}
}
func Test_avcodec_007(t *testing.T) {
	if codec := FindEncoderById(AV_CODEC_ID_H264); codec == nil {
		t.Error("Unable to find codec")
	} else if context := NewAVCodecContext(codec); context == nil {
		t.Error("Unexpected nil value returned")
	} else {
		t.Log(context)
		context.Free()
	}
}
func Test_avcodec_008(t *testing.T) {
	codecs := AllCodecs()
	for _, codec := range codecs {
		t.Log(codec)
	}
}
