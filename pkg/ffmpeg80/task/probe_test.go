package task_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mutablelogic/go-media/pkg/ffmpeg80/schema"
	"github.com/mutablelogic/go-media/pkg/ffmpeg80/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProbe_MP4(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "sample.mp4")

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request: schema.Request{Path: testPath},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Check format info
	assert.NotEmpty(t, resp.Format)
	assert.Contains(t, resp.Format, "mp4")
	t.Logf("Format: %s (%s)", resp.Format, resp.Description)

	// Check duration
	assert.Greater(t, resp.Duration, 0.0)
	t.Logf("Duration: %.3f seconds", resp.Duration)

	// Check streams
	assert.GreaterOrEqual(t, len(resp.Streams), 2)
	t.Logf("Streams: %d", len(resp.Streams))

	// Check video stream
	var hasVideo, hasAudio bool
	for _, s := range resp.Streams {
		t.Logf("  Stream %d: %s codec=%s", s.Index, s.Type, s.Codec)
		if s.Type == "video" {
			hasVideo = true
			assert.Greater(t, s.Width, 0)
			assert.Greater(t, s.Height, 0)
			assert.NotEmpty(t, s.PixelFormat)
			t.Logf("    Video: %dx%d %s", s.Width, s.Height, s.PixelFormat)
		}
		if s.Type == "audio" {
			hasAudio = true
			assert.Greater(t, s.SampleRate, 0)
			assert.Greater(t, s.Channels, 0)
			t.Logf("    Audio: %dHz %dch %s", s.SampleRate, s.Channels, s.ChannelLayout)
		}
	}
	assert.True(t, hasVideo, "expected video stream")
	assert.True(t, hasAudio, "expected audio stream")
}

func TestProbe_MP3(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "sample.mp3")

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request: schema.Request{Path: testPath},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Check format
	assert.NotEmpty(t, resp.Format)
	t.Logf("Format: %s (%s)", resp.Format, resp.Description)
	t.Logf("Duration: %.3f seconds", resp.Duration)

	// Check audio stream
	require.NotEmpty(t, resp.Streams)
	audioStream := resp.Streams[0]
	assert.Equal(t, "audio", audioStream.Type)
	assert.NotEmpty(t, audioStream.Codec)
	assert.Greater(t, audioStream.SampleRate, 0)
	t.Logf("Audio: %s %dHz %dch", audioStream.Codec, audioStream.SampleRate, audioStream.Channels)
}

func TestProbe_WAV(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "jfk.wav")

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request: schema.Request{Path: testPath},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Check format
	assert.NotEmpty(t, resp.Format)
	t.Logf("Format: %s (%s)", resp.Format, resp.Description)
	t.Logf("Duration: %.3f seconds", resp.Duration)

	// Check audio stream
	require.NotEmpty(t, resp.Streams)
	audioStream := resp.Streams[0]
	assert.Equal(t, "audio", audioStream.Type)
	assert.NotEmpty(t, audioStream.Codec)
	assert.Greater(t, audioStream.SampleRate, 0)
	assert.NotEmpty(t, audioStream.SampleFormat)
	t.Logf("Audio: %s %dHz %s", audioStream.Codec, audioStream.SampleRate, audioStream.SampleFormat)
}

func TestProbe_JPEG(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "sample.jpg")

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request: schema.Request{Path: testPath},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Check format
	assert.NotEmpty(t, resp.Format)
	t.Logf("Format: %s (%s)", resp.Format, resp.Description)

	// Check video stream (images appear as video)
	require.NotEmpty(t, resp.Streams)
	videoStream := resp.Streams[0]
	assert.Equal(t, "video", videoStream.Type)
	assert.Greater(t, videoStream.Width, 0)
	assert.Greater(t, videoStream.Height, 0)
	t.Logf("Image: %dx%d %s", videoStream.Width, videoStream.Height, videoStream.PixelFormat)
}

func TestProbe_PNG(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "sample.png")

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request: schema.Request{Path: testPath},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Check format
	assert.NotEmpty(t, resp.Format)
	t.Logf("Format: %s (%s)", resp.Format, resp.Description)

	// Check video stream (images appear as video)
	require.NotEmpty(t, resp.Streams)
	videoStream := resp.Streams[0]
	assert.Equal(t, "video", videoStream.Type)
	assert.Greater(t, videoStream.Width, 0)
	assert.Greater(t, videoStream.Height, 0)
	t.Logf("Image: %dx%d %s", videoStream.Width, videoStream.Height, videoStream.PixelFormat)
}

func TestProbe_AllFiles(t *testing.T) {
	m := task.NewManager()
	testDir := testDataPath(t)

	// List of test files to probe
	testFiles := []string{
		"sample.mp4",
		"sample.mp3",
		"jfk.wav",
		"sample.jpg",
		"sample.png",
	}

	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			testPath := filepath.Join(testDir, file)

			resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
				Request: schema.Request{Path: testPath},
			})

			require.NoError(t, err)
			assert.NotNil(t, resp)
			t.Logf("%s duration: %v", file, resp.Duration)
		})
	}
}

func TestProbe_FileNotFound(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "nonexistent.mp4")

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request: schema.Request{Path: testPath},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestProbeStream_MP4(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "sample.mp4")

	f, err := os.Open(testPath)
	require.NoError(t, err)
	defer f.Close()

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request: schema.Request{Reader: f, Path: "mp4"},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	t.Logf("MP4 stream duration: %v", resp.Duration)
}

func TestProbeStream_MP3(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "sample.mp3")

	f, err := os.Open(testPath)
	require.NoError(t, err)
	defer f.Close()

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request: schema.Request{Reader: f, Path: "mp3"},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	t.Logf("MP3 stream duration: %v", resp.Duration)
}

func TestProbeStream_WAV(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "jfk.wav")

	f, err := os.Open(testPath)
	require.NoError(t, err)
	defer f.Close()

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request: schema.Request{Reader: f, Path: "wav"},
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)
	t.Logf("WAV stream duration: %v", resp.Duration)
}

func TestProbeStream_AllFiles(t *testing.T) {
	m := task.NewManager()
	testDir := testDataPath(t)

	// List of test files to probe with their format hints
	testFiles := []struct {
		file   string
		format string
	}{
		{"sample.mp4", "mp4"},
		{"sample.mp3", "mp3"},
		{"jfk.wav", "wav"},
	}

	for _, tc := range testFiles {
		t.Run(tc.file+"_stream", func(t *testing.T) {
			testPath := filepath.Join(testDir, tc.file)

			f, err := os.Open(testPath)
			require.NoError(t, err)
			defer f.Close()

			resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
				Request: schema.Request{Reader: f, Path: tc.format},
			})

			require.NoError(t, err)
			assert.NotNil(t, resp)
			t.Logf("%s stream duration: %v", tc.file, resp.Duration)
		})
	}
}

func TestProbe_WithMetadata(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "sample.mp4")

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request:  schema.Request{Path: testPath},
		Metadata: true,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Log metadata
	t.Logf("Format: %s", resp.Format)
	t.Logf("Metadata entries: %d", len(resp.Metadata))
	for k, v := range resp.Metadata {
		t.Logf("  %s: %s", k, v)
	}

	// Artwork should be nil when not requested
	assert.Nil(t, resp.Artwork)
}

func TestProbe_WithoutMetadata(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "sample.mp4")

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request:  schema.Request{Path: testPath},
		Metadata: false,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Metadata should be nil when not requested
	assert.Nil(t, resp.Metadata)
	assert.Nil(t, resp.Artwork)
}

func TestProbe_MP3_MimeType(t *testing.T) {
	m := task.NewManager()
	testPath := filepath.Join(testDataPath(t), "sample.mp3")

	resp, err := m.Probe(context.Background(), &schema.ProbeRequest{
		Request:  schema.Request{Path: testPath},
		Metadata: true,
	})

	require.NoError(t, err)
	assert.NotNil(t, resp)

	// MP3 should have mime type
	assert.NotEmpty(t, resp.MimeTypes)
	assert.Contains(t, resp.MimeTypes, "audio/mpeg")
	t.Logf("MIME types: %v", resp.MimeTypes)
}
