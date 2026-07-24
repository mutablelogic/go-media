package writer_test

import (
	"bytes"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	// Packages
	profile "github.com/mutablelogic/go-media/profile/schema"
	writer "github.com/mutablelogic/go-media/writer"
)

func tempFileURL(t *testing.T, name string) *url.URL {
	t.Helper()
	return &url.URL{Path: filepath.Join(t.TempDir(), name)}
}

func audioStream(t *testing.T) *profile.AudioProfile {
	t.Helper()
	p, err := profile.NewAudioProfile("aac")
	if err != nil {
		t.Fatalf("NewAudioProfile(aac): %v", err)
	}
	if err := p.Set(profile.OptionSampleRate, uint64(44100)); err != nil {
		t.Fatalf("Set(sample_rate): %v", err)
	}
	if err := p.Set(profile.OptionSampleFormat, "fltp"); err != nil {
		t.Fatalf("Set(sample_format): %v", err)
	}
	if err := p.Set(profile.OptionChannelLayout, "stereo"); err != nil {
		t.Fatalf("Set(channel_layout): %v", err)
	}
	return p
}

// subtitleStream returns an "ass" subtitle profile. Of the sampled text
// subtitle codecs (ass, srt, mov_text, webvtt), only "ass" opens cleanly
// against this build of libavcodec - see sys/ffmpeg80's subtitle encode
// tests for details - so it's the one used for subtitle coverage here.
func subtitleStream(t *testing.T) *profile.SubtitleProfile {
	t.Helper()
	p, err := profile.NewSubtitleProfile("ass")
	if err != nil {
		t.Skipf("ass encoder not available: %v", err)
	}
	return p
}

func TestCreate(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	w, err := writer.Create(tempFileURL(t, "out.mp4"), output, writer.WithProfile(0, audioStream(t)))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if w == nil {
		t.Fatal("Create: nil writer")
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestCreate_NilURL(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	if _, err := writer.Create(nil, output, writer.WithProfile(0, audioStream(t))); err == nil {
		t.Fatal("Create: expected error for nil URL")
	}
}

func TestCreate_NilOutput(t *testing.T) {
	if _, err := writer.Create(tempFileURL(t, "out.mp4"), nil, writer.WithProfile(0, audioStream(t))); err == nil {
		t.Fatal("Create: expected error for nil output")
	}
}

func TestCreate_NilContext(t *testing.T) {
	// An Output constructed directly (rather than via OutputWithName/URL/Type)
	// has no resolved *ff.AVOutputFormat, and Create should reject it rather
	// than dereferencing a nil context.
	output := new(profile.Output)

	if _, err := writer.Create(tempFileURL(t, "out.mp4"), output, writer.WithProfile(0, audioStream(t))); err == nil {
		t.Fatal("Create: expected error for output with nil context")
	}
}

func TestCreate_NoStreams(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	if _, err := writer.Create(tempFileURL(t, "out.mp4"), output); err == nil {
		t.Fatal("Create: expected error for no streams")
	}
}

func TestCreate_OutputOpts(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}
	output.Opts = json.RawMessage(`{"movflags":"faststart"}`)

	path := tempFileURL(t, "out.mp4")
	w, err := writer.Create(path, output, writer.WithProfile(0, audioStream(t)))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	info, err := os.Stat(path.Path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty output file")
	}
}

func TestCreate_OutputOpts_Invalid(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}
	output.Opts = json.RawMessage(`{"not_a_real_mp4_option":"x"}`)

	if _, err := writer.Create(tempFileURL(t, "out.mp4"), output, writer.WithProfile(0, audioStream(t))); err == nil {
		t.Fatal("Create: expected error for unrecognized output option")
	}
}

func TestNewWriter(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	path := filepath.Join(t.TempDir(), "out.mp4")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("os.Create: %v", err)
	}
	defer f.Close()

	w, err := writer.NewWriter(f, output, writer.WithProfile(0, audioStream(t)))
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	if w == nil {
		t.Fatal("NewWriter: nil writer")
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

func TestNewWriter_BytesBuffer(t *testing.T) {
	// bytes.Buffer implements io.Writer but not gomedia.NamedWriter, so
	// NewWriter must fall back to an empty filename hint rather than panic.
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	var buf bytes.Buffer
	w, err := writer.NewWriter(&buf, output, writer.WithProfile(0, audioStream(t)))
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("expected non-empty output buffer")
	}
}

func TestNewWriter_NilWriter(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	if _, err := writer.NewWriter(nil, output, writer.WithProfile(0, audioStream(t))); err == nil {
		t.Fatal("NewWriter: expected error for nil writer")
	}
}

func TestNewWriter_NilOutput(t *testing.T) {
	var buf bytes.Buffer
	if _, err := writer.NewWriter(&buf, nil, writer.WithProfile(0, audioStream(t))); err == nil {
		t.Fatal("NewWriter: expected error for nil output")
	}
}

func TestNewWriter_NoStreams(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	var buf bytes.Buffer
	if _, err := writer.NewWriter(&buf, output); err == nil {
		t.Fatal("NewWriter: expected error for no streams")
	}
}

func TestNewWriter_EncodeFrame(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	var buf bytes.Buffer
	w, err := writer.NewWriter(&buf, output, writer.WithProfile(0, audioStream(t)))
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}

	numSamples := w.FrameSize(0)
	if numSamples == 0 {
		numSamples = 1024
	}

	for i := 0; i < 10; i++ {
		frame := silentFrame(t, 0, numSamples)
		frame.SetPts(int64(i * numSamples))

		if err := w.Encode(frame); err != nil {
			frame.Close()
			t.Fatalf("Encode: %v", err)
		}
		frame.Close()
	}

	if err := w.Flush(0); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("expected non-empty output buffer")
	}
}

func TestWriter_EncodeFrame(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	// Encode() with no packets muxed would still leave a minimal valid
	// header+trailer, so compare against that baseline to confirm frames
	// actually contributed real packet data, not just a bigger header.
	baselinePath := tempFileURL(t, "baseline.mp4")
	baseline, err := writer.Create(baselinePath, output, writer.WithProfile(0, audioStream(t)))
	if err != nil {
		t.Fatalf("Create (baseline): %v", err)
	}
	if err := baseline.Close(); err != nil {
		t.Fatalf("Close (baseline): %v", err)
	}
	baselineInfo, err := os.Stat(baselinePath.Path)
	if err != nil {
		t.Fatalf("Stat (baseline): %v", err)
	}

	path := tempFileURL(t, "out.mp4")
	w, err := writer.Create(path, output, writer.WithProfile(0, audioStream(t)))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	numSamples := w.FrameSize(0)
	if numSamples == 0 {
		numSamples = 1024
	}

	for i := 0; i < 10; i++ {
		frame := silentFrame(t, 0, numSamples)
		frame.SetPts(int64(i * numSamples))

		if err := w.Encode(frame); err != nil {
			frame.Close()
			t.Fatalf("Encode: %v", err)
		}
		frame.Close()
	}

	if err := w.Flush(0); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	info, err := os.Stat(path.Path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() <= baselineInfo.Size() {
		t.Fatalf("expected encoded output (%d bytes) to be larger than an empty baseline (%d bytes)", info.Size(), baselineInfo.Size())
	}
}

func TestClose_Idempotent(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	w, err := writer.Create(tempFileURL(t, "out.mp4"), output, writer.WithProfile(0, audioStream(t)))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close (first): %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close (second): %v", err)
	}
}
