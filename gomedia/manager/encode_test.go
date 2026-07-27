package manager_test

import (
	"os"
	"testing"

	// Packages
	task "github.com/mutablelogic/go-media/gomedia/task"
	test "github.com/mutablelogic/go-media/gomedia/test"
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
)

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func aacProfile(t *testing.T) profile.AudioProfile {
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

////////////////////////////////////////////////////////////////////////////////
// TESTS

func TestEncodeAudio(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	resp, err := m.EncodeAudio(ctx, task.AudioEncodeMediaRequest{
		Reader:  f,
		Profile: aacProfile(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response")
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
}

func TestEncodeAudio_NilReader(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	req := task.AudioEncodeMediaRequest{Profile: aacProfile(t)}
	if _, err := m.EncodeAudio(ctx, req); err == nil {
		t.Fatal("expected an error for a nil reader")
	}
}
