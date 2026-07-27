package task_test

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	task "github.com/mutablelogic/go-media/gomedia/task"
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	writer "github.com/mutablelogic/go-media/writer"
)

///////////////////////////////////////////////////////////////////////////////
// HELPERS

// sourceFilePath synthesizes ~2.3s of silent AAC audio into a real temp
// mp4 file (rather than an in-memory buffer - the mp4 muxer needs to seek
// back to finalize the moov atom, which a bytes.Buffer can't offer) and
// returns its path, for use as a AudioEncodeMediaRequest.Reader source.
func sourceFilePath(t *testing.T) string {
	t.Helper()

	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	p := audioEncodeProfile(t)
	path := filepath.Join(t.TempDir(), "src.mp4")
	w, err := writer.Create(&url.URL{Path: path}, output, writer.WithProfile(0, &p))
	if err != nil {
		t.Fatalf("writer.Create: %v", err)
	}

	numSamples := w.FrameSize(0)
	if numSamples == 0 {
		numSamples = 1024
	}
	for i := 0; i < 100; i++ {
		f, err := frame.NewAudioFrame(0)
		if err != nil {
			t.Fatalf("NewAudioFrame: %v", err)
		}
		f.SetSampleFormat(ff.AVUtil_get_sample_fmt("fltp"))
		f.SetSampleRate(44100)
		var ch ff.AVChannelLayout
		if err := ff.AVUtil_channel_layout_from_string(&ch, "stereo"); err != nil {
			t.Fatalf("AVUtil_channel_layout_from_string: %v", err)
		}
		if err := f.SetChannelLayout(ch); err != nil {
			t.Fatalf("SetChannelLayout: %v", err)
		}
		f.SetNumSamples(numSamples)
		if err := f.AllocateBuffers(); err != nil {
			t.Fatalf("AllocateBuffers: %v", err)
		}
		f.SetPts(int64(i * numSamples))

		if err := w.Encode(f); err != nil {
			f.Close()
			t.Fatalf("Encode: %v", err)
		}
		f.Close()
	}

	if err := w.Flush(0); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	return path
}

func audioEncodeProfile(t *testing.T) profile.AudioProfile {
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
	return *p
}

///////////////////////////////////////////////////////////////////////////////
// TESTS

func TestAudioEncodeMediaRequest_Run(t *testing.T) {
	src, err := os.Open(sourceFilePath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer src.Close()

	req := &task.AudioEncodeMediaRequest{
		Reader:  src,
		Profile: audioEncodeProfile(t),
	}

	var result any
	tctx := task.Context{
		Context: context.Background(),
		Result:  func(r any) { result = r },
	}
	if err := req.Run(tctx); err != nil {
		t.Fatalf("Run: %v", err)
	}

	resp, ok := result.(*task.AudioEncodeMediaResponse)
	if !ok || resp == nil {
		t.Fatalf("Result = %T, want *task.AudioEncodeMediaResponse", result)
	}
	defer os.Remove(resp.Path)

	// The encoded output must itself be a real, decodable file.
	out, err := reader.Open(resp.Path)
	if err != nil {
		t.Fatalf("reader.Open(%q): %v", resp.Path, err)
	}
	defer out.Close()

	if out.Duration() <= 0 {
		t.Fatalf("Duration() = %v, want > 0", out.Duration())
	}
	if len(out.Streams()) != 1 {
		t.Fatalf("Streams() = %d entries, want 1", len(out.Streams()))
	}
}

func TestAudioEncodeMediaRequest_Run_NilReader(t *testing.T) {
	req := &task.AudioEncodeMediaRequest{Profile: audioEncodeProfile(t)}
	if err := req.Run(task.Context{Context: context.Background()}); err == nil {
		t.Fatal("Run: expected an error for a nil reader")
	}
}
