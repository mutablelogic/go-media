package manager_test

import (
	"testing"

	// Packages
	uuid "github.com/google/uuid"
	schema "github.com/mutablelogic/go-media/task/schema"
	test "github.com/mutablelogic/go-media/task/test"
	require "github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func TestManager_GetTask(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	id, err := mgr.Add(ctx, "probe", &fakeTask{done: make(chan struct{})})
	require.NoError(err)

	status, err := mgr.GetTask(ctx, id)
	require.NoError(err)
	require.Equal(id, status.UUID)
	require.Equal("probe", status.Name)
	require.Equal("fake", status.Task)
	require.Equal(schema.StateNotStarted, status.State())
}

func TestManager_GetTaskNotFound(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	_, err := mgr.GetTask(ctx, uuid.New())
	require.Error(err)
}
