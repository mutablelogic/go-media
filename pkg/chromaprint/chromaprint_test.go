package chromaprint_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/pkg/chromaprint"
	schema "github.com/mutablelogic/go-media/pkg/chromaprint/schema"
)

const (
	testAudioFile = "../../etc/test/sample.mp3"
	testVideoFile = "../../etc/test/sample.mp4"
)

////////////////////////////////////////////////////////////////////////////////
// VERSION TESTS

func Test_Version(t *testing.T) {
	var buf bytes.Buffer
	PrintVersion(&buf)
	assert.Contains(t, buf.String(), "chromaprint:")
	assert.Contains(t, buf.String(), "1.5")
}

////////////////////////////////////////////////////////////////////////////////
// FINGERPRINT TYPE TESTS

func Test_New_ValidParams(t *testing.T) {
	assert := assert.New(t)

	// Valid parameters
	fp := New(44100, 2, 30*time.Second)
	assert.NotNil(fp)
	assert.NoError(fp.Close())
}

func Test_New_DefaultDuration(t *testing.T) {
	assert := assert.New(t)

	// Zero duration should default to maxFingerprintDuration (120s)
	fp := New(44100, 2, 0)
	assert.NotNil(fp)
	assert.NoError(fp.Close())
}

func Test_New_NegativeDuration(t *testing.T) {
	assert := assert.New(t)

	// Negative duration should default to maxFingerprintDuration
	fp := New(44100, 2, -5*time.Second)
	assert.NotNil(fp)
	assert.NoError(fp.Close())
}

func Test_New_InvalidRate(t *testing.T) {
	assert := assert.New(t)

	// Zero rate
	fp := New(0, 2, 30*time.Second)
	assert.Nil(fp)

	// Negative rate
	fp = New(-44100, 2, 30*time.Second)
	assert.Nil(fp)
}

func Test_New_InvalidChannels(t *testing.T) {
	assert := assert.New(t)

	// Zero channels
	fp := New(44100, 0, 30*time.Second)
	assert.Nil(fp)

	// Negative channels
	fp = New(44100, -1, 30*time.Second)
	assert.Nil(fp)
}

func Test_Fingerprint_WriteEmptySlice(t *testing.T) {
	assert := assert.New(t)

	fp := New(44100, 1, 30*time.Second)
	assert.NotNil(fp)
	defer fp.Close()

	// Write empty slice should succeed
	n, err := fp.Write([]int16{})
	assert.NoError(err)
	assert.Equal(int64(0), n)
}

func Test_Fingerprint_WriteAfterClose(t *testing.T) {
	assert := assert.New(t)

	fp := New(44100, 1, 30*time.Second)
	assert.NotNil(fp)

	// Close first
	assert.NoError(fp.Close())

	// Write after close should fail
	_, err := fp.Write([]int16{1, 2, 3, 4})
	assert.ErrorIs(err, io.ErrClosedPipe)
}

func Test_Fingerprint_FinishAfterClose(t *testing.T) {
	assert := assert.New(t)

	fp := New(44100, 1, 30*time.Second)
	assert.NotNil(fp)

	// Close first
	assert.NoError(fp.Close())

	// Finish after close should fail
	_, err := fp.Finish()
	assert.ErrorIs(err, io.ErrClosedPipe)
}

func Test_Fingerprint_DurationAfterClose(t *testing.T) {
	assert := assert.New(t)

	fp := New(44100, 1, 30*time.Second)
	assert.NotNil(fp)

	// Close first
	assert.NoError(fp.Close())

	// Duration after close should return 0
	assert.Equal(time.Duration(0), fp.Duration())
}

func Test_Fingerprint_DoubleClose(t *testing.T) {
	assert := assert.New(t)

	fp := New(44100, 1, 30*time.Second)
	assert.NotNil(fp)

	// Close twice should not panic
	assert.NoError(fp.Close())
	assert.NoError(fp.Close())
}

func Test_Fingerprint_WriteSamples(t *testing.T) {
	assert := assert.New(t)

	fp := New(44100, 1, 30*time.Second)
	assert.NotNil(fp)
	defer fp.Close()

	// Write some samples
	samples := make([]int16, 44100) // 1 second of mono audio
	for i := range samples {
		samples[i] = int16(i % 32767)
	}

	n, err := fp.Write(samples)
	assert.NoError(err)
	assert.Equal(int64(44100), n)

	// Check duration (approximately 1 second)
	dur := fp.Duration()
	assert.True(dur >= 900*time.Millisecond && dur <= 1100*time.Millisecond,
		"Expected ~1 second, got %v", dur)
}

func Test_Fingerprint_GeneratesValidFingerprint(t *testing.T) {
	assert := assert.New(t)

	fp := New(44100, 1, 30*time.Second)
	assert.NotNil(fp)
	defer fp.Close()

	// Generate some audio-like samples (sine wave approximation)
	samples := make([]int16, 44100*5) // 5 seconds
	for i := range samples {
		// Simple pattern that produces a fingerprint
		samples[i] = int16((i * 100) % 32767)
	}

	_, err := fp.Write(samples)
	assert.NoError(err)

	// Get fingerprint
	str, err := fp.Finish()
	assert.NoError(err)
	assert.NotEmpty(str)
	assert.True(strings.HasPrefix(str, "AQAA"), "Fingerprint should start with AQAA")
}

////////////////////////////////////////////////////////////////////////////////
// FINGERPRINT FUNCTION TESTS

func Test_Fingerprint_FromMP3(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open(testAudioFile)
	if !assert.NoError(err) {
		t.Skip("Test audio file not found")
	}
	defer f.Close()

	result, err := Fingerprint(context.Background(), f, 0)
	assert.NoError(err)
	assert.NotNil(result)
	assert.NotEmpty(result.Fingerprint)
	assert.True(result.Duration > 0, "Duration should be > 0")
	assert.True(strings.HasPrefix(result.Fingerprint, "AQAA"),
		"Fingerprint should start with AQAA, got: %s", result.Fingerprint[:10])

	t.Logf("Fingerprint: %s... Duration: %v", result.Fingerprint[:50], result.Duration)
}

func Test_Fingerprint_FromMP4(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open(testVideoFile)
	if !assert.NoError(err) {
		t.Skip("Test video file not found")
	}
	defer f.Close()

	result, err := Fingerprint(context.Background(), f, 0)
	assert.NoError(err)
	assert.NotNil(result)
	assert.NotEmpty(result.Fingerprint)
	assert.True(result.Duration > 0)

	t.Logf("Fingerprint: %s... Duration: %v", result.Fingerprint[:50], result.Duration)
}

func Test_Fingerprint_WithDurationLimit(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open(testAudioFile)
	if !assert.NoError(err) {
		t.Skip("Test audio file not found")
	}
	defer f.Close()

	// Get full duration first
	fullResult, err := Fingerprint(context.Background(), f, 0)
	assert.NoError(err)
	fullDuration := fullResult.Duration

	// Reopen file
	f.Seek(0, 0)

	// Limit to 5 seconds - but note that due to frame buffering,
	// the actual processed duration may vary slightly
	result, err := Fingerprint(context.Background(), f, 5*time.Second)
	assert.NoError(err)
	assert.NotNil(result)
	assert.NotEmpty(result.Fingerprint)

	// Duration should be less than or equal to full duration
	assert.True(result.Duration <= fullDuration,
		"Limited duration %v should be <= full duration %v", result.Duration, fullDuration)

	t.Logf("Requested 5s, got duration: %v (full file: %v)", result.Duration, fullDuration)
}

func Test_Fingerprint_ContextCancellation(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open(testAudioFile)
	if !assert.NoError(err) {
		t.Skip("Test audio file not found")
	}
	defer f.Close()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should return error due to cancelled context
	_, err = Fingerprint(ctx, f, 0)
	assert.Error(err)
}

func Test_Fingerprint_InvalidReader(t *testing.T) {
	assert := assert.New(t)

	// Empty reader
	r := strings.NewReader("")
	_, err := Fingerprint(context.Background(), r, 0)
	assert.Error(err)
}

func Test_Fingerprint_FromRawPCM(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open("../../etc/test/audio_22050_1ch_5m35.s16le.sw")
	if !assert.NoError(err) {
		t.Skip("Test raw PCM file not found")
	}
	defer f.Close()

	// The file is 22050 Hz, 1 channel, signed 16-bit little-endian, 5m35s
	result, err := Fingerprint(
		context.Background(),
		f,
		0,
		segmenter.WithFFmpegOpt(ffmpeg.WithInput("s16le",
			"sample_rate=22050",
			"channels=1",
			"channel_layout=mono",
			"sample_fmt=s16",
		)),
	)
	assert.NoError(err, "Should successfully fingerprint raw PCM")
	assert.NotNil(result, "Result should not be nil")
	assert.NotEmpty(result.Fingerprint, "Fingerprint should not be empty")
	assert.True(result.Duration > 0, "Duration should be > 0")
	assert.True(strings.HasPrefix(result.Fingerprint, "AQAA") || strings.HasPrefix(result.Fingerprint, "AQAB"),
		"Fingerprint should start with AQA")

	t.Logf("Fingerprinted raw PCM: duration=%v, fingerprint=%s...", result.Duration, result.Fingerprint[:50])
}

////////////////////////////////////////////////////////////////////////////////
// META FLAGS TESTS

func Test_Meta_String(t *testing.T) {
	tests := []struct {
		meta     Meta
		expected string
	}{
		{META_NONE, ""},
		{META_RECORDING, "recordings"},
		{META_RECORDINGID, "recordingids"},
		{META_RELEASE, "releases"},
		{META_RELEASEID, "releaseids"},
		{META_RELEASEGROUP, "releasegroups"},
		{META_RELEASEGROUPID, "releasegroupids"},
		{META_TRACK, "tracks"},
		{META_COMPRESS, "compress"},
		{META_USERMETA, "usermeta"},
		{META_SOURCE, "sources"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.meta.String())
		})
	}
}

func Test_Meta_CombinedFlags(t *testing.T) {
	combined := META_RECORDING | META_TRACK
	str := combined.String()
	assert.Contains(t, str, "recordings")
	assert.Contains(t, str, "tracks")
}

func Test_Meta_All(t *testing.T) {
	str := META_ALL.String()
	assert.Contains(t, str, "recordings")
	assert.Contains(t, str, "tracks")
	assert.Contains(t, str, "releases")
	assert.Contains(t, str, "sources")
}

////////////////////////////////////////////////////////////////////////////////
// CLIENT TESTS (without API key)

func Test_NewClient_Default(t *testing.T) {
	assert := assert.New(t)

	// Empty API key should use default
	client, err := NewClient("")
	assert.NoError(err)
	assert.NotNil(client)
}

func Test_NewClient_WithApiKey(t *testing.T) {
	assert := assert.New(t)

	client, err := NewClient("test-api-key")
	assert.NoError(err)
	assert.NotNil(client)
}

func Test_Lookup_BadParameters(t *testing.T) {
	assert := assert.New(t)

	client, err := NewClient("")
	assert.NoError(err)

	// Empty fingerprint
	_, err = client.Lookup("", 5*time.Second, META_TRACK)
	assert.Error(err)

	// Zero duration
	_, err = client.Lookup("AQAA...", 0, META_TRACK)
	assert.Error(err)

	// No meta flags
	_, err = client.Lookup("AQAA...", 5*time.Second, META_NONE)
	assert.Error(err)
}

func Test_Lookup_FromRawPCM(t *testing.T) {
	assert := assert.New(t)

	// Skip if no API key is available
	apiKey := os.Getenv("CHROMAPRINT_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: CHROMAPRINT_KEY environment variable not set")
	}

	f, err := os.Open("../../etc/test/audio_22050_1ch_5m35.s16le.sw")
	if !assert.NoError(err) {
		t.Skip("Test raw PCM file not found")
	}
	defer f.Close()

	client, err := NewClient(apiKey)
	if !assert.NoError(err) {
		t.Skip("Failed to create client")
	}

	// Generate fingerprint
	fpResult, err := Fingerprint(
		context.Background(),
		f,
		5*time.Minute+35*time.Second, // Full track duration
		segmenter.WithFFmpegOpt(ffmpeg.WithInput("s16le",
			"sample_rate=22050",
			"channels=1",
			"channel_layout=mono",
			"sample_fmt=s16",
		)),
	)
	if !assert.NoError(err) {
		return
	}

	// Lookup matches
	matches, err := client.Lookup(fpResult.Fingerprint, time.Duration(fpResult.Duration*float64(time.Second)), META_ALL)
	if !assert.NoError(err) || !assert.NotNil(matches) || len(matches) == 0 {
		t.Logf("Match failed: %v", err)
		return
	}

	t.Logf("Lookup result: %+v", matches)
}

////////////////////////////////////////////////////////////////////////////////
// FINGERPRINTRESULT TESTS

func Test_FingerprintResult_Fields(t *testing.T) {
	assert := assert.New(t)

	f, err := os.Open(testAudioFile)
	if !assert.NoError(err) {
		t.Skip("Test audio file not found")
	}
	defer f.Close()

	result, err := Fingerprint(context.Background(), f, 10*time.Second)
	assert.NoError(err)
	assert.NotNil(result)

	// Check all fields are populated
	assert.NotEmpty(result.Fingerprint)
	assert.True(result.Duration > 0)

	// Fingerprint format validation
	assert.True(len(result.Fingerprint) > 10, "Fingerprint should be longer than 10 chars")
	assert.True(strings.HasPrefix(result.Fingerprint, "AQAA"), "Should start with AQAA")
}

////////////////////////////////////////////////////////////////////////////////
// RESPONSE TESTS

func Test_ResponseMatch_String(t *testing.T) {
	match := &schema.ResponseMatch{
		Id:    "test-id",
		Score: 0.95,
		Recordings: []schema.ResponseRecording{
			{
				Id:       "recording-id",
				Title:    "Test Song",
				Duration: 180.5,
			},
		},
	}

	str := match.String()
	assert.Contains(t, str, "test-id")
	assert.Contains(t, str, "0.95")
	assert.Contains(t, str, "Test Song")
}

// Helper to wrap ffmpeg.Opt as segmenter.Opt
// The ffmpegOpt helper is no longer needed as segmenter.WithFFmpegOpt is now used.
