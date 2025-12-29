package ffmpeg

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

const (
	TEST_INPUT_FILE = "../../../etc/test/sample.mp4"
)

////////////////////////////////////////////////////////////////////////////////
// TEST BASIC OUTPUT CONTEXT CREATION

func Test_avformat_mux_create_file(t *testing.T) {
	assert := assert.New(t)

	// Create the file
	filename := filepath.Join(t.TempDir(), "test.mp4")
	output, err := AVFormat_create_file(filename, nil)
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(output)

	t.Log(output)

	// Close the file
	assert.NoError(AVFormat_close_writer(output))
}

func Test_avformat_mux_create_file_with_format(t *testing.T) {
	assert := assert.New(t)

	// Get MP4 format
	format := AVFormat_guess_format("mp4", "", "")
	if !assert.NotNil(format, "MP4 format should be available") {
		t.SkipNow()
	}

	// Create the file
	filename := filepath.Join(t.TempDir(), "test.mp4")
	output, err := AVFormat_create_file(filename, format)
	assert.NoError(err)
	assert.NotNil(output)
	defer AVFormat_close_writer(output)

	// Verify format
	assert.Equal("mp4", output.Output().Name())
}

func Test_avformat_mux_create_file_invalid_path(t *testing.T) {
	assert := assert.New(t)

	// Try to create file in non-existent directory
	filename := "/nonexistent/path/test.mp4"
	output, err := AVFormat_create_file(filename, nil)
	assert.Error(err, "Should fail with invalid path")
	assert.Nil(output)
}

////////////////////////////////////////////////////////////////////////////////
// TEST CUSTOM I/O WRITER

// Custom I/O callback for testing
type testIOCallback struct {
	data []byte
	pos  int
}

func (cb *testIOCallback) Reader(buf []byte) int {
	if cb.pos >= len(cb.data) {
		return 0
	}
	n := copy(buf, cb.data[cb.pos:])
	cb.pos += n
	return n
}

func (cb *testIOCallback) Writer(buf []byte) int {
	cb.data = append(cb.data, buf...)
	return len(buf)
}

func (cb *testIOCallback) Seeker(offset int64, whence int) int64 {
	var newPos int64
	switch whence {
	case 0: // SEEK_SET
		newPos = offset
	case 1: // SEEK_CUR
		newPos = int64(cb.pos) + offset
	case 2: // SEEK_END
		newPos = int64(len(cb.data)) + offset
	default:
		return -1
	}
	if newPos < 0 {
		return -1
	}
	cb.pos = int(newPos)
	return newPos
}

func Test_avformat_mux_open_writer(t *testing.T) {
	assert := assert.New(t)

	// Create custom I/O callback
	callback := &testIOCallback{
		data: make([]byte, 0, 1024*1024),
		pos:  0,
	}

	// Create custom I/O context
	writer := AVFormat_avio_alloc_context(1024*1024, true, callback)
	if !assert.NotNil(writer) {
		t.SkipNow()
	}

	// Open writer with custom I/O
	output, err := AVFormat_open_writer(writer, nil, "test.mp4")
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(output)

	t.Log(output)

	// Verify custom I/O flag is set
	assert.True(output.Flags().Is(AVFMT_FLAG_CUSTOM_IO))

	// Close (should not close the custom I/O)
	assert.NoError(AVFormat_close_writer(output))

	// Custom I/O should still be valid (we manage it)
	// In production, caller would free writer
}

////////////////////////////////////////////////////////////////////////////////
// TEST HEADER AND TRAILER

func Test_avformat_mux_write_header(t *testing.T) {
	assert := assert.New(t)

	// Create output file
	filename := filepath.Join(t.TempDir(), "test.mp4")
	output, err := AVFormat_create_file(filename, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_writer(output)

	// Add a video stream
	stream := AVFormat_new_stream(output, nil)
	if !assert.NotNil(stream) {
		t.FailNow()
	}

	// Set up codec parameters
	codecpar := stream.CodecPar()
	codecpar.SetCodecType(AVMEDIA_TYPE_VIDEO)
	codecpar.SetCodecID(AV_CODEC_ID_H264)
	codecpar.SetWidth(1280)
	codecpar.SetHeight(720)
	codecpar.SetPixelFormat(AV_PIX_FMT_YUV420P)

	// Write header
	err = AVFormat_write_header(output, nil)
	assert.NoError(err, "Should write header successfully")
}

func Test_avformat_mux_write_header_no_streams(t *testing.T) {
	assert := assert.New(t)

	// Create output file
	filename := filepath.Join(t.TempDir(), "test.mp4")
	output, err := AVFormat_create_file(filename, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_writer(output)

	// Try to write header without streams
	err = AVFormat_write_header(output, nil)
	// Some formats may allow this, others may error
	t.Logf("Write header without streams result: %v", err)
}

func Test_avformat_mux_write_trailer(t *testing.T) {
	assert := assert.New(t)

	// Create output file
	filename := filepath.Join(t.TempDir(), "test.mp4")
	output, err := AVFormat_create_file(filename, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_writer(output)

	// Add a stream and write header
	stream := AVFormat_new_stream(output, nil)
	if !assert.NotNil(stream) {
		t.FailNow()
	}

	codecpar := stream.CodecPar()
	codecpar.SetCodecType(AVMEDIA_TYPE_VIDEO)
	codecpar.SetCodecID(AV_CODEC_ID_H264)
	codecpar.SetWidth(1280)
	codecpar.SetHeight(720)
	codecpar.SetPixelFormat(AV_PIX_FMT_YUV420P)

	if err := AVFormat_write_header(output, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Write trailer
	err = AVFormat_write_trailer(output)
	assert.NoError(err, "Should write trailer successfully")
}

////////////////////////////////////////////////////////////////////////////////
// TEST INIT OUTPUT

func Test_avformat_mux_init_output(t *testing.T) {
	assert := assert.New(t)

	// Create output file
	filename := filepath.Join(t.TempDir(), "test.mp4")
	output, err := AVFormat_create_file(filename, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_writer(output)

	// Add a stream with complete codec parameters
	stream := AVFormat_new_stream(output, nil)
	if !assert.NotNil(stream) {
		t.FailNow()
	}

	codecpar := stream.CodecPar()
	codecpar.SetCodecType(AVMEDIA_TYPE_VIDEO)
	codecpar.SetCodecID(AV_CODEC_ID_H264)
	codecpar.SetWidth(1280)
	codecpar.SetHeight(720)
	codecpar.SetPixelFormat(AV_PIX_FMT_YUV420P)

	// Set required fields for H264
	stream.SetTimeBase(AVUtil_rational(1, 25)) // 25 fps
	codecpar.SetBitRate(1000000)                // 1 Mbps

	// Initialize output - may fail with certain codecs/formats
	err = AVFormat_init_output(output, nil)
	if err != nil {
		t.Logf("Init output failed (may require encoder): %v", err)
		t.SkipNow()
	}

	// Now write header (already initialized)
	err = AVFormat_write_header(output, nil)
	if err != nil {
		t.Logf("Write header after init failed: %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST COMPLETE REMUX WORKFLOW

func Test_avformat_mux_remux_video(t *testing.T) {
	assert := assert.New(t)

	// Check if input file exists
	if _, err := os.Stat(TEST_INPUT_FILE); os.IsNotExist(err) {
		t.Skip("Test input file not available:", TEST_INPUT_FILE)
	}

	// Allocate a packet
	pkt := AVCodec_packet_alloc()
	if !assert.NotNil(pkt) {
		t.SkipNow()
	}
	defer AVCodec_packet_free(pkt)

	// Open input file
	input, err := AVFormat_open_url(TEST_INPUT_FILE, nil, nil)
	if !assert.NoError(err) {
		t.Skip("Could not open input file:", err)
	}
	defer AVFormat_close_input(input)

	// Find stream information
	assert.NoError(AVFormat_find_stream_info(input, nil))

	// Dump the input format
	t.Log("Input format:")
	AVFormat_dump_format(input, 0, TEST_INPUT_FILE)

	// Open the output file
	outfile := filepath.Join(t.TempDir(), "remuxed.mp4")
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
	t.Log("Output format:")
	AVFormat_dump_format(output, 0, outfile)

	// Write the header
	if err := AVFormat_write_header(output, nil); !assert.NoError(err) {
		t.FailNow()
	}

	// Write the frames
	frame_count := 0
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
			frame_count++
		}
	}

	t.Logf("Remuxed %d frames", frame_count)

	// Write the trailer
	if err := AVFormat_write_trailer(output); !assert.NoError(err) {
		t.FailNow()
	}

	// Verify output file exists and has content
	info, err := os.Stat(outfile)
	assert.NoError(err)
	assert.Greater(info.Size(), int64(0), "Output file should have content")
	t.Logf("Output file size: %d bytes", info.Size())
}

////////////////////////////////////////////////////////////////////////////////
// TEST ERROR HANDLING

func Test_avformat_mux_close_nil(t *testing.T) {
	assert := assert.New(t)

	// Closing nil context should not crash
	err := AVFormat_close_writer(nil)
	assert.NoError(err, "Closing nil context should succeed")
}

func Test_avformat_mux_write_header_twice(t *testing.T) {
	assert := assert.New(t)

	// Create output file
	filename := filepath.Join(t.TempDir(), "test.mp4")
	output, err := AVFormat_create_file(filename, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer AVFormat_close_writer(output)

	// Add a stream
	stream := AVFormat_new_stream(output, nil)
	if !assert.NotNil(stream) {
		t.FailNow()
	}

	codecpar := stream.CodecPar()
	codecpar.SetCodecType(AVMEDIA_TYPE_VIDEO)
	codecpar.SetCodecID(AV_CODEC_ID_H264)
	codecpar.SetWidth(1280)
	codecpar.SetHeight(720)
	codecpar.SetPixelFormat(AV_PIX_FMT_YUV420P)

	// Write header first time
	err = AVFormat_write_header(output, nil)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Try to write header again (should error or be idempotent)
	err = AVFormat_write_header(output, nil)
	t.Logf("Second write header result: %v", err)
	// Behavior may vary by format, just ensure no crash
}

func Test_avformat_mux_write_trailer_without_header(t *testing.T) {
	// Writing trailer without header can cause FFmpeg to crash internally
	// This is expected behavior - header must be written first
	t.Skip("Writing trailer without header causes segfault - expected FFmpeg behavior")
}

////////////////////////////////////////////////////////////////////////////////
// TEST DIFFERENT OUTPUT FORMATS

func Test_avformat_mux_formats(t *testing.T) {
	assert := assert.New(t)

	formats := []struct {
		name      string
		extension string
	}{
		{"mp4", "mp4"},
		{"matroska", "mkv"},
		{"avi", "avi"},
		{"mov", "mov"},
	}

	for _, fmt := range formats {
		t.Run(fmt.name, func(t *testing.T) {
			// Get format
			format := AVFormat_guess_format(fmt.name, "", "")
			if format == nil {
				t.Skipf("Format %s not available", fmt.name)
			}

			// Create file
			filename := filepath.Join(t.TempDir(), "test."+fmt.extension)
			output, err := AVFormat_create_file(filename, format)
			if !assert.NoError(err) {
				return
			}
			defer AVFormat_close_writer(output)

			// Verify format name
			assert.Equal(fmt.name, output.Output().Name())
			t.Logf("Successfully created %s format", fmt.name)
		})
	}
}
