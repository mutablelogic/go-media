package metadata_test

import (
	"os"
	"path/filepath"
	"testing"

	// Packages
	metadata "github.com/mutablelogic/go-media/task/metadata"
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

////////////////////////////////////////////////////////////////////////////////
// TESTS

// TestMetadataTask runs a MetadataRequest through the real task manager -
// Add, Start, Wait - rather than calling Run directly, to exercise the same
// path a server would use.
func TestMetadataTask(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	require.NoError(err)
	defer f.Close()

	id, err := mgr.Add(ctx, "metadata", &metadata.MetadataRequest{Reader: f})
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	status, err := mgr.Wait(ctx, id)
	require.NoError(err)
	require.Equal(schema.StateDone, status.State())

	result, ok := status.Result.(*metadata.MetadataResponse)
	require.True(ok, "expected *metadata.MetadataResponse, got %T", status.Result)
	require.Equal(f.Name(), result.Name)
	require.Equal("audio/mpeg", result.Type)
}

func TestMetadataTask_NilReader(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	id, err := mgr.Add(ctx, "metadata", &metadata.MetadataRequest{Reader: nil})
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	status, err := mgr.Wait(ctx, id)
	require.Error(err)
	require.Equal(schema.StateError, status.State())
}
