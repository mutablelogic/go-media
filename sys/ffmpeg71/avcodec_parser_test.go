package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avcodec_parser_000(t *testing.T) {
	//assert := assert.New(t)

	// Iterate over all codecs
	var opaque uintptr
	for {
		parser := AVCodec_parser_iterate(&opaque)
		if parser == nil {
			break
		}

		t.Log("codec_parser=", parser)
	}
}
