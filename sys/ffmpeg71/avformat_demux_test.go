package ffmpeg_test

import (
	"io"
	"os"
	"syscall"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

const (
	TEST_MP4_FILE = "../../etc/test/sample.mp4"
)

func Test_avformat_demux_001(t *testing.T) {
	assert := assert.New(t)

	// Open the file
	filereader, err := NewFileReader(TEST_MP4_FILE)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer filereader.Close()

	// Create the context
	ctx := AVFormat_avio_alloc_context(20, false, filereader)
	assert.NotNil(ctx)
	defer AVFormat_avio_context_free(ctx)

	// Open for demuxing
	input, err := AVFormat_open_reader(ctx, nil, nil)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer AVFormat_free_context(input)

	t.Log(input)
}

func Test_avformat_demux_002(t *testing.T) {
	assert := assert.New(t)

	// Open for demuxing
	input, err := AVFormat_open_url(TEST_MP4_FILE, nil, nil)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer AVFormat_free_context(input)

	packet := AVCodec_packet_alloc()
	if !assert.NotNil(packet) {
		t.SkipNow()
	}
	defer AVCodec_packet_free(packet)

	for {
		if err := AVFormat_read_frame(input, packet); err != nil {
			if err == io.EOF {
				break
			}
			if !assert.NoError(err) {
				t.FailNow()
			}
		}

		// Output the packet
		t.Logf("Packet: %v", packet)

		// Mark the packet as consumed
		AVCodec_packet_unref(packet)
	}

}

////////////////////////////////////////////////////////////////////////////////
// filereader implements the AVIOContext interface for reading from a file

type filereader struct {
	r io.ReadSeekCloser
}

func NewFileReader(filename string) (*filereader, error) {
	r, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &filereader{r}, nil
}

func (r *filereader) Reader(buf []byte) int {
	if n, err := r.r.Read(buf); err == io.EOF {
		return AVERROR_EOF
	} else if err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			return int(errno)
		} else {
			return AVERROR_UNKNOWN
		}
	} else {
		return n
	}
}

func (r *filereader) Writer([]byte) int {
	// Reader does not implement the writer
	return AVERROR_EOF
}

func (r *filereader) Seeker(offset int64, whence int) int64 {
	whence = whence & ^AVSEEK_FORCE
	switch whence {
	case AVSEEK_SIZE:
		// TODO: Not sure what to put here yet
		return -1
	case io.SeekStart, io.SeekCurrent, io.SeekEnd:
		n, err := r.r.Seek(offset, whence)
		if err != nil {
			return -1
		}
		return n
	default:
		return -1
	}
}

func (r *filereader) Close() error {
	return r.r.Close()
}
