package ffmpeg_test

import (
	"os"
	"strings"
	"testing"

	// Namespace imports
	. "github.com/djthorpe/go-media/sys/ffmpeg"
)

const (
	bufferSize = 4096 * 4

//	SAMPLE_MP4 = "../../etc/sample.mp4"
)

func Test_AVIO_001(t *testing.T) {
	AVFormatInit()
	defer AVFormatDeinit()

	// Set log callback
	AVLogSetCallback(AV_LOG_DEBUG, func(level AVLogLevel, message string, _ uintptr) {
		t.Log("level=", level, "message=", strings.TrimSpace(message))
	})

	// Open file for reading
	r, err := os.Open(SAMPLE_MP4)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	// Create IO context
	io := NewAVIOContext(bufferSize, false, r.Read, nil, nil)
	if io == nil {
		t.Fatal("Failed to create AVIOContext")
	}
	defer io.Free()

	// Open input file
	ctx := NewAVFormatContext()
	if err := ctx.OpenInputIO(io.AVIOContext, nil); err != nil {
		t.Fatal(err)
	} else {
		defer ctx.CloseInput()
	}

	// Find stream information
	if err := ctx.FindStreamInfo(); err != nil {
		t.Error(err)
	} else {
		ctx.Dump(0)
	}
}
