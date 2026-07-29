package manager_test

import (
	"os"
	"testing"

	// Packages
	test "github.com/mutablelogic/go-media/gomedia/test"
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
	taskencoder "github.com/mutablelogic/go-media/task/encoder"
)

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func mp4Output(t *testing.T) *profile.Output {
	t.Helper()
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}
	return output
}

func aacProfile(t *testing.T) *profile.AudioProfile {
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

////////////////////////////////////////////////////////////////////////////////
// TESTS

// TestEncode checks that Encode returns promptly with the task's status,
// without waiting for the (long-running) encode to finish itself - the
// caller is expected to wait on or poll the returned task's UUID via the
// task manager instead.
func TestEncode(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	status, err := m.Encode(ctx, taskencoder.EncodeRequest{
		Reader: f,
		Output: mp4Output(t),
		Audio:  aacProfile(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	if status == nil {
		t.Fatal("expected a non-nil status")
	}

	// Encode doesn't remove the task either, so it stays tracked by the task
	// manager - wait on it directly to get the final result.
	final, err := test.TaskManager(t).Wait(ctx, status.UUID)
	if err != nil {
		t.Fatal(err)
	}

	resp, ok := final.Result.(*taskencoder.EncodeResponse)
	if !ok || resp == nil {
		t.Fatalf("Result = %T, want *taskencoder.EncodeResponse", final.Result)
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

// TestEncode_NilReader checks that a malformed request is rejected by
// Encode itself, before any task is even created - the caller shouldn't
// have to poll a "started" task just to learn it was doomed from the start.
func TestEncode_NilReader(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	req := taskencoder.EncodeRequest{Output: mp4Output(t), Audio: aacProfile(t)}
	if _, err := m.Encode(ctx, req); err == nil {
		t.Fatal("expected an error for a nil reader")
	}
}
