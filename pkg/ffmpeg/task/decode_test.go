package task_test

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	"github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	"github.com/mutablelogic/go-media/pkg/ffmpeg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func TestDecode_NilRequest(t *testing.T) {
	m, err := task.NewManager()
	require.NoError(t, err)

	var buf bytes.Buffer
	err = m.Decode(context.Background(), &buf, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil decode request")
}

func TestDecode_NilWriter(t *testing.T) {
	m, err := task.NewManager()
	require.NoError(t, err)

	err = m.Decode(context.Background(), nil, &schema.DecodeRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil writer")
}

func TestDecode_NoInput(t *testing.T) {
	m, err := task.NewManager()
	require.NoError(t, err)

	var buf bytes.Buffer
	err = m.Decode(context.Background(), &buf, &schema.DecodeRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Input or Reader must be provided")
}

func TestDecode_MP4_JSON(t *testing.T) {
	m, err := task.NewManager()
	require.NoError(t, err)

	testPath := filepath.Join(testDataPath(t), "sample.mp4")
	var buf bytes.Buffer

	err = m.Decode(context.Background(), &buf, &schema.DecodeRequest{
		Request: schema.Request{Input: testPath},
	})
	require.NoError(t, err)

	// Parse JSON output
	output := buf.String()
	assert.NotEmpty(t, output)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Greater(t, len(lines), 0, "should have decoded at least one frame")

	// Parse each line to verify it's valid JSON
	for i, line := range lines {
		var frame map[string]interface{}
		err := json.Unmarshal([]byte(line), &frame)
		assert.NoError(t, err, "line %d should be valid JSON", i)
	}

	t.Logf("Decoded %d frames", len(lines))
}

func TestDecode_MP3_JSON(t *testing.T) {
	m, err := task.NewManager()
	require.NoError(t, err)

	testPath := filepath.Join(testDataPath(t), "sample.mp3")
	var buf bytes.Buffer

	err = m.Decode(context.Background(), &buf, &schema.DecodeRequest{
		Request: schema.Request{Input: testPath},
	})
	require.NoError(t, err)

	// Parse JSON output
	output := buf.String()
	assert.NotEmpty(t, output)

	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Greater(t, len(lines), 0, "should have decoded at least one frame")

	// Parse each line to verify it's valid JSON
	for i, line := range lines {
		var frame map[string]interface{}
		err := json.Unmarshal([]byte(line), &frame)
		assert.NoError(t, err, "line %d should be valid JSON", i)
	}

	t.Logf("Decoded %d audio frames", len(lines))
}

func TestDecode_WithCancellation(t *testing.T) {
	m, err := task.NewManager()
	require.NoError(t, err)

	testPath := filepath.Join(testDataPath(t), "sample.mp4")
	var buf bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = m.Decode(ctx, &buf, &schema.DecodeRequest{
		Request: schema.Request{Input: testPath},
	})
	assert.Error(t, err)
}

////////////////////////////////////////////////////////////////////////////////
// FRAME WRITER TESTS

type testFrameWriter struct {
	bytes.Buffer
	frames []testFrame
}

type testFrame struct {
	streamIndex int
	mediaType   string
	width       int
	height      int
	sampleRate  int
	channels    int
}

func (w *testFrameWriter) WriteFrame(streamIndex int, frame interface{}) error {
	f, ok := frame.(*ffmpeg.Frame)
	if !ok {
		return nil
	}

	tf := testFrame{
		streamIndex: streamIndex,
	}

	// Debug: log frame type
	frameType := f.Type()
	sampleRate := f.SampleRate()
	width := f.Width()

	// Determine media type from frame
	switch frameType {
	case 2: // media.VIDEO
		tf.mediaType = "video"
		tf.width = f.Width()
		tf.height = f.Height()
	case 1: // media.AUDIO
		tf.mediaType = "audio"
		tf.sampleRate = f.SampleRate()
		tf.channels = f.ChannelLayout().NumChannels()
	default:
		// Frame.Type() returns UNKNOWN - try to infer from properties
		if sampleRate > 0 {
			tf.mediaType = "audio"
			tf.sampleRate = sampleRate
			tf.channels = f.ChannelLayout().NumChannels()
		} else if width > 0 {
			tf.mediaType = "video"
			tf.width = width
			tf.height = f.Height()
		} else {
			tf.mediaType = "unknown"
		}
	}

	w.frames = append(w.frames, tf)
	return nil
}

func TestDecode_MP4_FrameWriter(t *testing.T) {
	m, err := task.NewManager()
	require.NoError(t, err)

	testPath := filepath.Join(testDataPath(t), "sample.mp4")
	writer := &testFrameWriter{}

	err = m.Decode(context.Background(), writer, &schema.DecodeRequest{
		Request: schema.Request{Input: testPath},
	})
	require.NoError(t, err)

	// Check that frames were captured
	assert.Greater(t, len(writer.frames), 0, "should have decoded frames")

	// Find video and audio frames
	var hasVideo, hasAudio bool
	for _, f := range writer.frames {
		if f.mediaType == "video" {
			hasVideo = true
			assert.Greater(t, f.width, 0)
			assert.Greater(t, f.height, 0)
			t.Logf("Video frame: %dx%d", f.width, f.height)
		} else if f.mediaType == "audio" {
			hasAudio = true
			assert.Greater(t, f.sampleRate, 0)
			assert.Greater(t, f.channels, 0)
			t.Logf("Audio frame: rate=%d, channels=%d", f.sampleRate, f.channels)
		}
	}

	assert.True(t, hasVideo, "should have decoded video frames")
	assert.True(t, hasAudio, "should have decoded audio frames")

	// Check that JSON output was also written to the buffer
	assert.NotEmpty(t, writer.Buffer.String(), "Should also write JSON even with FrameWriter")
}

func TestDecode_MP3_FrameWriter(t *testing.T) {
	m, err := task.NewManager()
	require.NoError(t, err)

	testPath := filepath.Join(testDataPath(t), "sample.mp3")
	writer := &testFrameWriter{}

	err = m.Decode(context.Background(), writer, &schema.DecodeRequest{
		Request: schema.Request{Input: testPath},
	})
	require.NoError(t, err)

	// Check that frames were captured
	assert.Greater(t, len(writer.frames), 0, "should have decoded frames")

	// All frames should be audio
	for _, f := range writer.frames {
		assert.Equal(t, "audio", f.mediaType)
		assert.Greater(t, f.sampleRate, 0)
		assert.Greater(t, f.channels, 0)
	}
}
