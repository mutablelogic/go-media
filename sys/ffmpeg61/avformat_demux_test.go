package ffmpeg_test

import (
	"io"
	"os"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

const (
	TEST_MP4_FILE = "../../etc/test/sample.mp4"
)

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
		// TODO: Return errno errors
		return AVERROR_UNKNOWN
	} else {
		return n
	}
}

func (r *filereader) Writer([]byte) int {
	return AVERROR_EOF
}

func (r *filereader) Seeker(offset int64, whence int) int64 {
	n, err := r.r.Seek(offset, whence)
	if err != nil {
		return -1
	}
	return n
}

func (r *filereader) Close() error {
	return r.r.Close()
}

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

	t.Log(input)
}
