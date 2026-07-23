package testcard_test

import (
	"context"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Packages
	testcard "github.com/mutablelogic/go-media/gomedia/testcard"
	profile "github.com/mutablelogic/go-media/profile/schema"
	writer "github.com/mutablelogic/go-media/writer"
)

// monoFloatProfile returns a raw audio profile shaped like what TestCard
// requires: mono, float32, at the given sample rate. pcm_f32le (like all raw
// PCM codecs) only supports interleaved "flt", not planar "fltp" — planar is
// an in-memory layout internal to codecs like AAC/Opus, not a raw byte
// format. For mono audio the two are identical anyway (a single channel has
// no interleaving to do), so Frame.Float32(0) works the same either way.
func monoFloatProfile(t *testing.T, sampleRate uint64) *profile.AudioProfile {
	t.Helper()
	p, err := profile.NewAudioProfile("pcm_f32le")
	if err != nil {
		t.Fatalf("NewAudioProfile(pcm_f32le): %v", err)
	}
	if err := p.Set(profile.OptionSampleRate, sampleRate); err != nil {
		t.Fatalf("Set(sample_rate): %v", err)
	}
	if err := p.Set(profile.OptionSampleFormat, "flt"); err != nil {
		t.Fatalf("Set(sample_format): %v", err)
	}
	if err := p.Set(profile.OptionChannelLayout, "mono"); err != nil {
		t.Fatalf("Set(channel_layout): %v", err)
	}
	return p
}

func TestNew(t *testing.T) {
	tc, err := testcard.New(monoFloatProfile(t, 44100))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer tc.Close()
}

func TestNew_NilProfile(t *testing.T) {
	if _, err := testcard.New(nil); err == nil {
		t.Fatal("New: expected error for nil profile")
	}
}

func TestNew_Stereo(t *testing.T) {
	p, err := profile.NewAudioProfile("pcm_f32le")
	if err != nil {
		t.Fatalf("NewAudioProfile(pcm_f32le): %v", err)
	}
	if err := p.Set(profile.OptionSampleRate, uint64(44100)); err != nil {
		t.Fatalf("Set(sample_rate): %v", err)
	}
	if err := p.Set(profile.OptionSampleFormat, "flt"); err != nil {
		t.Fatalf("Set(sample_format): %v", err)
	}
	if err := p.Set(profile.OptionChannelLayout, "stereo"); err != nil {
		t.Fatalf("Set(channel_layout): %v", err)
	}

	if _, err := testcard.New(p); err == nil {
		t.Fatal("New: expected error for stereo profile")
	}
}

func TestNew_NonFloatFormat(t *testing.T) {
	p, err := profile.NewAudioProfile("pcm_s16le")
	if err != nil {
		t.Fatalf("NewAudioProfile(pcm_s16le): %v", err)
	}
	if err := p.Set(profile.OptionSampleRate, uint64(44100)); err != nil {
		t.Fatalf("Set(sample_rate): %v", err)
	}
	if err := p.Set(profile.OptionSampleFormat, "s16"); err != nil {
		t.Fatalf("Set(sample_format): %v", err)
	}
	if err := p.Set(profile.OptionChannelLayout, "mono"); err != nil {
		t.Fatalf("Set(channel_layout): %v", err)
	}

	if _, err := testcard.New(p); err == nil {
		t.Fatal("New: expected error for non-float sample format")
	}
}

func TestStreams(t *testing.T) {
	audio := monoFloatProfile(t, 44100)
	tc, err := testcard.New(audio)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer tc.Close()

	streams := tc.Streams()
	if len(streams) != 1 {
		t.Fatalf("Streams: expected 1 stream, got %d", len(streams))
	}
	if streams[0] != profile.Profile(audio) {
		t.Fatal("Streams: expected the audio profile passed to New")
	}
}

func TestNextFrame(t *testing.T) {
	tc, err := testcard.New(monoFloatProfile(t, 44100), testcard.WithFrequency(1000), testcard.WithVolume(0))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer tc.Close()

	ctx := context.Background()

	var lastPts int64
	var sawNonZero bool
	var peak float32
	for i := 0; i < 3; i++ {
		f, err := tc.NextFrame(ctx)
		if err != nil {
			t.Fatalf("NextFrame: %v", err)
		}
		if f.StreamID != 0 {
			t.Fatalf("NextFrame: expected StreamID 0, got %d", f.StreamID)
		}
		if i > 0 && f.Pts() <= lastPts {
			t.Fatalf("NextFrame: expected increasing Pts, got %d after %d", f.Pts(), lastPts)
		}
		lastPts = f.Pts()

		for _, s := range f.Float32(0) {
			if s != 0 {
				sawNonZero = true
			}
			if abs := float32(math.Abs(float64(s))); abs > peak {
				peak = abs
			}
		}
	}

	if !sawNonZero {
		t.Fatal("expected non-zero samples")
	}
	// At 0dB (amplitude 1.0), the sine wave's peak should approach 1.0.
	if peak < 0.9 {
		t.Fatalf("expected peak amplitude near 1.0 at 0dB, got %v", peak)
	}
}

func TestNextFrame_NotPaced(t *testing.T) {
	tc, err := testcard.New(monoFloatProfile(t, 44100))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer tc.Close()

	// The source does not pace itself to real time — the caller decides
	// pacing, so many frames back-to-back should return quickly.
	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 100; i++ {
		if _, err := tc.NextFrame(ctx); err != nil {
			t.Fatalf("NextFrame: %v", err)
		}
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Fatalf("expected NextFrame to return immediately (no pacing), 100 calls took %v", elapsed)
	}
}

func TestNextFrame_ContextCancelled(t *testing.T) {
	tc, err := testcard.New(monoFloatProfile(t, 44100))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer tc.Close()

	ctx := context.Background()
	if _, err := tc.NextFrame(ctx); err != nil {
		t.Fatalf("NextFrame (first): %v", err)
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := tc.NextFrame(cancelCtx); err == nil {
		t.Fatal("NextFrame: expected error for cancelled context")
	}
}

func TestNextFrame_AfterClose(t *testing.T) {
	tc, err := testcard.New(monoFloatProfile(t, 44100))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := tc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := tc.NextFrame(context.Background()); err == nil {
		t.Fatal("NextFrame: expected error after Close")
	}
}

// TestPipeline drives the test card straight into a real Writer/Encoder,
// confirming the whole chain produces a valid, non-trivial encoded file.
// Uses libopus rather than aac: aac only accepts planar "fltp" samples, and
// bridging the test card's interleaved "flt" output to that would need a
// resampler — a real need (see the format-continuity discussion), but out
// of scope here. libopus accepts interleaved "flt" directly, so this stays
// a genuine no-resampling pipeline test.
func TestPipeline(t *testing.T) {
	const sampleRate = 48000 // libopus only accepts a fixed set of rates

	tc, err := testcard.New(monoFloatProfile(t, sampleRate), testcard.WithFrequency(1000))
	if err != nil {
		t.Fatalf("testcard.New: %v", err)
	}
	defer tc.Close()

	encodeProfile, err := profile.NewAudioProfile("libopus")
	if err != nil {
		t.Skipf("libopus encoder not available: %v", err)
	}
	if err := encodeProfile.Set(profile.OptionSampleRate, uint64(sampleRate)); err != nil {
		t.Fatalf("Set(sample_rate): %v", err)
	}
	if err := encodeProfile.Set(profile.OptionSampleFormat, "flt"); err != nil {
		t.Fatalf("Set(sample_format): %v", err)
	}
	if err := encodeProfile.Set(profile.OptionChannelLayout, "mono"); err != nil {
		t.Fatalf("Set(channel_layout): %v", err)
	}

	output := profile.OutputWithName("ogg")
	if output == nil {
		t.Fatal("OutputWithName(ogg): nil output")
	}

	path := filepath.Join(t.TempDir(), "testcard.ogg")
	w, err := writer.Create(&url.URL{Path: path}, output, writer.WithProfile(0, encodeProfile))
	if err != nil {
		t.Fatalf("writer.Create: %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		f, err := tc.NextFrame(ctx)
		if err != nil {
			t.Fatalf("NextFrame: %v", err)
		}
		if err := w.Encode(f); err != nil {
			t.Fatalf("Encode: %v", err)
		}
	}
	if err := w.Flush(0); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty output file")
	}
}
