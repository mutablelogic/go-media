package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"

	// Packages
	media "github.com/mutablelogic/go-media"
	assert "github.com/stretchr/testify/assert"
)

const (
	testDir = "../../etc/test"
)

////////////////////////////////////////////////////////////////////////////////
// TEST OPEN FROM FILE PATH

func Test_reader_open_mp4(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}

func Test_reader_open_mp3(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp3")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}

func Test_reader_open_wav(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "jfk.wav")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}

func Test_reader_open_jpg(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.jpg")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}

func Test_reader_open_png(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.png")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}

func Test_reader_open_nonexistent(t *testing.T) {
	assert := assert.New(t)

	_, err := Open("/nonexistent/file.mp4")
	assert.Error(err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST OPEN FROM io.Reader

func Test_reader_from_io_reader_mp4(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	f, err := os.Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer f.Close()

	r, err := NewReader(f)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}

func Test_reader_from_io_reader_mp3(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp3")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	f, err := os.Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer f.Close()

	r, err := NewReader(f)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}

func Test_reader_from_io_reader_wav(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "jfk.wav")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	f, err := os.Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer f.Close()

	r, err := NewReader(f)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}

////////////////////////////////////////////////////////////////////////////////
// TEST PROPERTIES

func Test_reader_type(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	mediaType := r.Type()
	assert.True(mediaType.Is(media.INPUT))
	t.Log("Type:", mediaType)
}

func Test_reader_duration(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	duration := r.Duration()
	assert.NotZero(duration)
	t.Log("Duration:", duration)
}

func Test_reader_duration_wav(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "jfk.wav")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	duration := r.Duration()
	assert.NotZero(duration)
	t.Log("Duration:", duration)
}

func Test_reader_duration_image(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.jpg")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Images may have a small duration (default framerate)
	duration := r.Duration()
	t.Log("Duration:", duration)
}

////////////////////////////////////////////////////////////////////////////////
// TEST STREAMS

func Test_reader_streams_any(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	streams := r.Streams(media.ANY)
	assert.NotEmpty(streams)
	assert.GreaterOrEqual(len(streams), 1)

	for _, stream := range streams {
		assert.NotNil(stream)
		t.Logf("Stream %d: Type=%v", stream.Index(), stream.Type())
	}
}

func Test_reader_streams_video(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	streams := r.Streams(media.VIDEO)
	assert.NotEmpty(streams)

	for _, stream := range streams {
		assert.NotNil(stream)
		assert.True(stream.Type().Is(media.VIDEO))
		t.Logf("Video stream %d", stream.Index())
	}
}

func Test_reader_streams_audio(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	streams := r.Streams(media.AUDIO)
	assert.NotEmpty(streams)

	for _, stream := range streams {
		assert.NotNil(stream)
		assert.True(stream.Type().Is(media.AUDIO))
		t.Logf("Audio stream %d", stream.Index())
	}
}

func Test_reader_streams_audio_only(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "jfk.wav")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Should have audio streams
	audioStreams := r.Streams(media.AUDIO)
	assert.NotEmpty(audioStreams)
	assert.Equal(1, len(audioStreams))

	// Should not have video streams
	videoStreams := r.Streams(media.VIDEO)
	assert.Empty(videoStreams)

	// All streams should return audio only
	allStreams := r.Streams(media.ANY)
	assert.Equal(len(audioStreams), len(allStreams))
}

func Test_reader_streams_image(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.jpg")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Images are treated as video streams
	streams := r.Streams(media.VIDEO)
	assert.NotEmpty(streams)
	assert.Equal(1, len(streams))

	allStreams := r.Streams(media.ANY)
	assert.Equal(len(streams), len(allStreams))
}

func Test_reader_streams_subtitle(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Most test files don't have subtitles
	streams := r.Streams(media.SUBTITLE)
	// Just ensure it doesn't crash and returns valid (possibly empty) slice
	assert.NotNil(streams)
	t.Logf("Found %d subtitle streams", len(streams))
}

func Test_reader_streams_multiple_types(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Get all streams
	allStreams := r.Streams(media.ANY)

	// Get video and audio separately
	videoStreams := r.Streams(media.VIDEO)
	audioStreams := r.Streams(media.AUDIO)

	// The sum should be less than or equal to all (some files may have data/subtitle)
	assert.GreaterOrEqual(len(allStreams), len(videoStreams)+len(audioStreams))

	t.Logf("Total: %d, Video: %d, Audio: %d",
		len(allStreams), len(videoStreams), len(audioStreams))
}

////////////////////////////////////////////////////////////////////////////////
// TEST BEST STREAM

func Test_reader_best_stream_video(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	stream := r.BestStream(media.VIDEO)
	assert.NotEqual(-1, stream)
	t.Log("Best video stream:", stream)
}

func Test_reader_best_stream_audio(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	stream := r.BestStream(media.AUDIO)
	assert.NotEqual(-1, stream)
	t.Log("Best audio stream:", stream)
}

func Test_reader_best_stream_audio_only(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "jfk.wav")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Should have audio
	stream := r.BestStream(media.AUDIO)
	assert.NotEqual(-1, stream)
	t.Log("Best audio stream:", stream)

	// Should not have video
	stream = r.BestStream(media.VIDEO)
	assert.Equal(-1, stream)
}

func Test_reader_best_stream_image(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.jpg")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	stream := r.BestStream(media.VIDEO)
	assert.NotEqual(-1, stream)
	t.Log("Best video stream:", stream)
}

func Test_reader_best_stream_subtitle(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// MP4 sample likely doesn't have subtitles
	stream := r.BestStream(media.SUBTITLE)
	t.Log("Best subtitle stream:", stream)
}

////////////////////////////////////////////////////////////////////////////////
// TEST METADATA

func Test_reader_metadata(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	meta := r.Metadata()
	assert.NotNil(meta)
	for _, m := range meta {
		t.Logf("  %s: %v", m.Key(), m.Value())
	}
}

func Test_reader_metadata_filtered(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Request specific keys
	meta := r.Metadata("encoder", "title")
	for _, m := range meta {
		t.Logf("  %s: %v", m.Key(), m.Value())
	}
}

func Test_reader_metadata_artwork(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp3")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Request artwork
	meta := r.Metadata(MetaArtwork)
	for _, m := range meta {
		if m.Key() == MetaArtwork {
			bytes := m.Bytes()
			t.Logf("Artwork size: %d bytes", len(bytes))

			// Try to decode as image
			img := m.Image()
			if img != nil {
				t.Logf("Artwork dimensions: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST SEEK

func Test_reader_seek(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Get best video stream
	stream := r.BestStream(media.VIDEO)
	if !assert.NotEqual(-1, stream) {
		t.FailNow()
	}

	// Seek to 1 second
	err = r.Seek(stream, 1.0)
	assert.NoError(err)
}

func Test_reader_seek_invalid_stream(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	// Seek with invalid stream
	err = r.Seek(999, 1.0)
	assert.Error(err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_reader_json(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	data, err := r.MarshalJSON()
	assert.NoError(err)
	assert.NotEmpty(data)
	t.Log("JSON:", string(data))
}

func Test_reader_string(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	str := r.String()
	assert.NotEmpty(str)
	t.Log(str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST CLOSE

func Test_reader_close(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}

	err = r.Close()
	assert.NoError(err)
}

func Test_reader_close_twice(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// First close
	err = r.Close()
	assert.NoError(err)

	// Second close - should not panic
	err = r.Close()
	assert.NoError(err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST WITH OPTIONS

func Test_reader_with_input_format(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile, WithInput("mp4"))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}

func Test_reader_with_invalid_input_format(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	_, err := Open(testFile, WithInput("invalid_format"))
	assert.Error(err)
}

func Test_reader_with_input_opt(t *testing.T) {
	assert := assert.New(t)

	testFile := filepath.Join(testDir, "sample.mp4")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test file not available:", testFile)
	}

	r, err := Open(testFile, WithInput("", "analyzeduration=1000000"))
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	assert.NotNil(r)
	t.Log(r)
}
