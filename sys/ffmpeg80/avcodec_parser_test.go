package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AVCodecParser_iterate_001(t *testing.T) {
	assert := assert.New(t)

	// Iterate over all parsers
	var opaque uintptr
	count := 0
	for {
		parser := AVCodec_parser_iterate(&opaque)
		if parser == nil {
			break
		}
		count++
		assert.NotNil(parser)
	}

	t.Logf("Found %d parsers", count)
	assert.Greater(count, 0)
}

func Test_AVCodecParser_init_001(t *testing.T) {
	assert := assert.New(t)

	// Test parser initialization for H264
	parser := AVCodec_parser_init(AV_CODEC_ID_H264)
	assert.NotNil(parser)
	defer AVCodec_parser_close(parser)
}

func Test_AVCodecParser_init_002(t *testing.T) {
	// Test parser initialization for various codec types
	testCases := []struct {
		name    string
		codecID AVCodecID
	}{
		{"H264", AV_CODEC_ID_H264},
		{"MPEG1VIDEO", AV_CODEC_ID_MPEG1VIDEO},
		{"MPEG2VIDEO", AV_CODEC_ID_MPEG2VIDEO},
		{"MP2", AV_CODEC_ID_MP2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := AVCodec_parser_init(tc.codecID)
			if parser != nil {
				defer AVCodec_parser_close(parser)
				t.Logf("Parser for %s initialized successfully", tc.name)
			} else {
				t.Logf("No parser available for %s", tc.name)
			}
		})
	}
}
