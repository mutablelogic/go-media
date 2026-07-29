package httpclient_test

import (
	"context"
	"os"
	"testing"

	// Packages
	test "github.com/mutablelogic/go-media/gomedia/test"
	profile "github.com/mutablelogic/go-media/profile/schema"
	taskencoder "github.com/mutablelogic/go-media/task/encoder"
	taskschema "github.com/mutablelogic/go-media/task/schema"
	require "github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func aacProfile(t *testing.T) *profile.AudioProfile {
	t.Helper()
	p, err := profile.NewAudioProfile("aac")
	require.NoError(t, err)
	require.NoError(t, p.Set(profile.OptionSampleRate, uint64(44100)))
	require.NoError(t, p.Set(profile.OptionSampleFormat, "fltp"))
	require.NoError(t, p.Set(profile.OptionChannelLayout, "stereo"))
	return p
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

// TestEncode drives gomedia/httpclient's typed Encode method against the
// real test server, checking it returns the started task's status promptly.
func TestEncode(t *testing.T) {
	require := require.New(t)
	client := test.Client(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	require.NoError(err)
	defer f.Close()

	output := profile.OutputWithName("mp4")
	require.NotNil(output)

	status, err := client.Encode(context.Background(), taskencoder.EncodeRequest{
		Reader: f,
		Output: output,
		Audio:  aacProfile(t),
	}, nil, nil)
	require.NoError(err)
	require.NotZero(status.UUID)
	require.Equal("encoder", status.Task)
}

// TestEncode_Stream drives Encode with a non-nil fn against the real test
// server, checking it streams this task's own events through to a finished
// result and returns that same final status.
func TestEncode_Stream(t *testing.T) {
	require := require.New(t)
	client := test.Client(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	require.NoError(err)
	defer f.Close()

	output := profile.OutputWithName("mp4")
	require.NotNil(output)

	var seenFinished bool
	status, err := client.Encode(context.Background(), taskencoder.EncodeRequest{
		Reader: f,
		Output: output,
		Audio:  aacProfile(t),
	}, nil, func(e *taskschema.Event) error {
		if e.Name == taskschema.EventFinished {
			seenFinished = true
		}
		return nil
	})
	require.NoError(err)
	require.True(seenFinished, "did not observe a finished event")
	require.NotNil(status)
	require.Empty(status.Err)
	require.Equal(taskschema.StateDone, status.State())

	result, ok := status.Result.(map[string]any)
	require.True(ok, "expected a result object, got %T", status.Result)
	path, _ := result["path"].(string)
	require.NotEmpty(path)
	defer os.Remove(path)
}
