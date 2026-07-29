package encoder_test

import (
	"os"
	"path/filepath"
	"testing"

	// Packages
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
	encoder "github.com/mutablelogic/go-media/task/encoder"
	schema "github.com/mutablelogic/go-media/task/schema"
	test "github.com/mutablelogic/go-media/task/test"
	require "github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func sampleFilePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "etc", "test", name)
}

func mp4Output(t *testing.T) *profile.Output {
	t.Helper()
	output := profile.OutputWithName("mp4")
	require.NotNil(t, output)
	return output
}

func aacProfile(t *testing.T) *profile.AudioProfile {
	t.Helper()
	p, err := profile.NewAudioProfile("aac")
	require.NoError(t, err)
	require.NoError(t, p.Set(profile.OptionSampleRate, uint64(44100)))
	require.NoError(t, p.Set(profile.OptionSampleFormat, "fltp"))
	require.NoError(t, p.Set(profile.OptionChannelLayout, "stereo"))
	return p
}

// mpeg4Profile matches sample.mp4's own video parameters (1280x720,
// yuv420p, 25fps) - mpeg4 is a built-in FFmpeg encoder, so this test has no
// dependency on an external codec library being available.
func mpeg4Profile(t *testing.T) *profile.VideoProfile {
	t.Helper()
	p, err := profile.NewVideoProfile("mpeg4")
	require.NoError(t, err)
	require.NoError(t, p.Set(profile.OptionWidth, uint64(1280)))
	require.NoError(t, p.Set(profile.OptionHeight, uint64(720)))
	require.NoError(t, p.Set(profile.OptionPixelFormat, "yuv420p"))
	require.NoError(t, p.Set(profile.OptionFrameRate, float64(25)))
	return p
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

// TestEncodeTask runs an EncodeRequest through the real task manager - Add,
// Start, Wait - rather than calling Run directly, to exercise the same path
// a server would use.
func TestEncodeTask(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	require.NoError(err)
	defer f.Close()

	req := &encoder.EncodeRequest{
		Reader: f,
		Output: mp4Output(t),
		Audio:  aacProfile(t),
	}

	id, err := mgr.Add(ctx, "encoder", req)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	status, err := mgr.Wait(ctx, id)
	require.NoError(err)
	require.Equal(schema.StateDone, status.State())

	result, ok := status.Result.(*encoder.EncodeResponse)
	require.True(ok, "expected *encoder.EncodeResponse, got %T", status.Result)
	defer os.Remove(result.Path)

	// Progress should have counted the input's decoded frames.
	require.Positive(status.Progress.Current)

	// The encoded output must itself be a real, decodable file.
	out, err := reader.Open(result.Path)
	require.NoError(err)
	defer out.Close()
	require.True(out.Duration() > 0)
	require.Len(out.Streams(), 1)
}

// TestEncodeTask_Video is the video-only counterpart to TestEncodeTask -
// same shape, but transcoding sample.mp4's video stream (and dropping its
// audio stream, since Audio is left unset).
func TestEncodeTask_Video(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp4"))
	require.NoError(err)
	defer f.Close()

	req := &encoder.EncodeRequest{
		Reader: f,
		Output: mp4Output(t),
		Video:  mpeg4Profile(t),
	}

	id, err := mgr.Add(ctx, "encoder", req)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	status, err := mgr.Wait(ctx, id)
	require.NoError(err)
	require.Equal(schema.StateDone, status.State())

	result, ok := status.Result.(*encoder.EncodeResponse)
	require.True(ok, "expected *encoder.EncodeResponse, got %T", status.Result)
	defer os.Remove(result.Path)

	// Progress should have counted the input's decoded frames.
	require.Positive(status.Progress.Current)

	out, err := reader.Open(result.Path)
	require.NoError(err)
	defer out.Close()
	require.True(out.Duration() > 0)
	require.Len(out.Streams(), 1)
}

// TestEncodeTask_AudioOnlyFromMultiStreamInput is a regression test: sample.mp4's
// audio stream is at input index 1 (video is 0), and selecting it alone
// produces an output with exactly one stream - at physical index 0. A
// previous version of Run reused the input's own stream index as the
// writer's id, tagging every packet stream_index=1 against an output that
// only ever had a stream 0, which FFmpeg silently rejects ("Invalid packet
// stream index"), producing an empty audio track. Unlike TestEncodeTask
// (audio-only from sample.mp3, whose only stream is already index 0) and
// TestEncodeTask_AudioAndVideo (which selects both of sample.mp4's streams,
// so the input and output indices happen to already line up), this is the
// one shape that actually exercises the mismatch.
func TestEncodeTask_AudioOnlyFromMultiStreamInput(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp4"))
	require.NoError(err)
	defer f.Close()

	req := &encoder.EncodeRequest{
		Reader: f,
		Output: mp4Output(t),
		Audio:  aacProfile(t),
	}

	id, err := mgr.Add(ctx, "encoder", req)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	status, err := mgr.Wait(ctx, id)
	require.NoError(err)
	require.Equal(schema.StateDone, status.State())

	result, ok := status.Result.(*encoder.EncodeResponse)
	require.True(ok, "expected *encoder.EncodeResponse, got %T", status.Result)
	defer os.Remove(result.Path)

	out, err := reader.Open(result.Path)
	require.NoError(err)
	defer out.Close()
	require.True(out.Duration() > 0)
	require.Len(out.Streams(), 1)
}

// TestEncodeTask_AudioAndVideo exercises both streams at once, checking
// that each is matched to its own distinct input stream and encoded
// concurrently into the same output.
func TestEncodeTask_AudioAndVideo(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp4"))
	require.NoError(err)
	defer f.Close()

	req := &encoder.EncodeRequest{
		Reader: f,
		Output: mp4Output(t),
		Audio:  aacProfile(t),
		Video:  mpeg4Profile(t),
	}

	id, err := mgr.Add(ctx, "encoder", req)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	status, err := mgr.Wait(ctx, id)
	require.NoError(err)
	require.Equal(schema.StateDone, status.State())

	result, ok := status.Result.(*encoder.EncodeResponse)
	require.True(ok, "expected *encoder.EncodeResponse, got %T", status.Result)
	defer os.Remove(result.Path)

	// Progress should have counted decoded frames from both streams.
	require.Positive(status.Progress.Current)

	out, err := reader.Open(result.Path)
	require.NoError(err)
	defer out.Close()
	require.True(out.Duration() > 0)
	require.Len(out.Streams(), 2)
}

// TestEncodeTask_NilReader, TestEncodeTask_MissingOutput and
// TestEncodeTask_NoProfile check that a malformed request is rejected by
// Add itself (via EncodeRequest.Validate), before a task is even
// registered - not just eventually, once waited on.

func TestEncodeTask_NilReader(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	req := &encoder.EncodeRequest{Output: mp4Output(t), Audio: aacProfile(t)}
	_, err := mgr.Add(ctx, "encoder", req)
	require.Error(err)
}

func TestEncodeTask_MissingOutput(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	require.NoError(err)
	defer f.Close()

	req := &encoder.EncodeRequest{Reader: f, Audio: aacProfile(t)}
	_, err = mgr.Add(ctx, "encoder", req)
	require.Error(err)
}

func TestEncodeTask_NoProfile(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	require.NoError(err)
	defer f.Close()

	req := &encoder.EncodeRequest{Reader: f}
	_, err = mgr.Add(ctx, "encoder", req)
	require.Error(err)
}
