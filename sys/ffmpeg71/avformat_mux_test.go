package ffmpeg_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avformat_mux_001(t *testing.T) {
	assert := assert.New(t)

	// Create the file
	filename := filepath.Join(os.TempDir(), "test.mp4")
	output, err := AVFormat_create_file(filename, nil)
	if !assert.NoError(err) {
		t.FailNow()
	}

	t.Log(output)

	// Close the file
	assert.NoError(AVFormat_close_writer(output))

}

func Test_avformat_mux_002(t *testing.T) {
	assert := assert.New(t)

	// Allocate a packet
	pkt := AVCodec_packet_alloc()
	if !assert.NotNil(pkt) {
		t.SkipNow()
	}
	defer AVCodec_packet_free(pkt)

	// Open input file
	input, err := AVFormat_open_url(TEST_MP4_FILE, nil, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_input(input)

	// Fine stream information
	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Dump the input format
	AVFormat_dump_format(input, 0, TEST_MP4_FILE)

	// Open the output file
	outfile := filepath.Join(os.TempDir(), "test.mp4")
	output, err := AVFormat_create_file(outfile, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_writer(output)

	// Stream mapping
	stream_map := make([]int, input.NumStreams())
	stream_index := 0
	for i := range stream_map {
		in_stream := input.Stream(i)
		in_codec_par := in_stream.CodecPar()

		// Ignore if not audio, video or subtitle
		if in_codec_par.CodecType() != AVMEDIA_TYPE_AUDIO && in_codec_par.CodecType() != AVMEDIA_TYPE_VIDEO && in_codec_par.CodecType() != AVMEDIA_TYPE_SUBTITLE {
			stream_map[i] = -1
			continue
		}

		// Create a new stream
		stream_map[i] = stream_index
		stream_index = stream_index + 1

		// Create a new output stream
		out_stream := AVFormat_new_stream(output, nil)
		if !assert.NotNil(out_stream) {
			t.FailNow()
		}

		// Copy the codec parameters
		if err := AVCodec_parameters_copy(out_stream.CodecPar(), in_codec_par); !assert.NoError(err) {
			t.FailNow()
		}

		out_stream.CodecPar().SetCodecTag(0)
	}

	// Dump the output format
	AVFormat_dump_format(output, 0, outfile)

	// Write the header
	if err := AVFormat_write_header(output, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Write the frames
	for {
		if err := AVFormat_read_frame(input, pkt); err != nil {
			if err == io.EOF {
				break
			}
			if !assert.NoError(err) {
				t.FailNow()
			}
		}
		in_stream := input.Stream(pkt.StreamIndex())
		if out_stream_index := stream_map[pkt.StreamIndex()]; out_stream_index < 0 {
			continue
		} else {
			out_stream := output.Stream(out_stream_index)

			/* copy packet */
			AVCodec_packet_rescale_ts(pkt, in_stream.TimeBase(), out_stream.TimeBase())
			pkt.SetPos(-1)

			if err := AVFormat_interleaved_write_frame(output, pkt); !assert.NoError(err) {
				t.FailNow()
			}
		}
	}

	// Write the trailer
	if err := AVFormat_write_trailer(output); !assert.NoError(err) {
		t.FailNow()
	}
}
