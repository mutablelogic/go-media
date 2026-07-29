package manager_test

import (
	"testing"

	// Packages
	uuid "github.com/google/uuid"
	schema "github.com/mutablelogic/go-media/task/schema"
	test "github.com/mutablelogic/go-media/task/test"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
	require "github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func TestManager_ListTasksFiltersByState(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	// The manager is shared across every test in this package, so measure
	// deltas against a baseline rather than assuming an absolute count.
	notStarted := schema.StateNotStarted
	before, err := mgr.ListTasks(ctx, schema.TaskListRequest{})
	require.NoError(err)
	beforeNotStarted, err := mgr.ListTasks(ctx, schema.TaskListRequest{State: &notStarted})
	require.NoError(err)

	const n = 3
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ft := &fakeTask{done: make(chan struct{})}
		defer close(ft.done)
		id, err := mgr.Add(ctx, "listed", ft)
		require.NoError(err)
		ids[i] = id
	}

	all, err := mgr.ListTasks(ctx, schema.TaskListRequest{})
	require.NoError(err)
	require.Equal(before.Count+n, all.Count)

	filtered, err := mgr.ListTasks(ctx, schema.TaskListRequest{State: &notStarted})
	require.NoError(err)
	require.Equal(beforeNotStarted.Count+n, filtered.Count)

	seen := make(map[uuid.UUID]bool, len(filtered.Body))
	for _, status := range filtered.Body {
		require.Equal(schema.StateNotStarted, status.State())
		seen[status.UUID] = true
	}
	for _, id := range ids {
		require.True(seen[id], "expected task %v in filtered results", id)
	}
}

func TestManager_ListTasksPaging(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	notStarted := schema.StateNotStarted
	before, err := mgr.ListTasks(ctx, schema.TaskListRequest{State: &notStarted})
	require.NoError(err)

	const n = 3
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ft := &fakeTask{done: make(chan struct{})}
		defer close(ft.done)
		id, err := mgr.Add(ctx, "paged", ft)
		require.NoError(err)
		ids[i] = id
	}

	// Skip every pre-existing not-started task, then page through just the
	// ones added above, two at a time.
	limit := uint64(2)
	page, err := mgr.ListTasks(ctx, schema.TaskListRequest{
		State:       &notStarted,
		OffsetLimit: pg.OffsetLimit{Offset: before.Count, Limit: &limit},
	})
	require.NoError(err)
	require.Equal(before.Count+n, page.Count)
	require.Len(page.Body, 2)
	require.Equal(ids[0], page.Body[0].UUID)
	require.Equal(ids[1], page.Body[1].UUID)

	offset := before.Count + 2
	rest, err := mgr.ListTasks(ctx, schema.TaskListRequest{
		State:       &notStarted,
		OffsetLimit: pg.OffsetLimit{Offset: offset, Limit: &limit},
	})
	require.NoError(err)
	require.Len(rest.Body, 1)
	require.Equal(ids[2], rest.Body[0].UUID)
	require.Equal(uint64(1), types.Value(rest.Limit), "limit should be clamped to the items available after offset")
}
