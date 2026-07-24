package manager_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/profile/schema"
	test "github.com/mutablelogic/go-media/profile/test"
	types "github.com/mutablelogic/go-server/pkg/types"
	require "github.com/stretchr/testify/require"
)

///////////////////////////////////////////////////////////////////////////////
// TESTS

func TestListFormats_FilterByName(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListFormats(ctx, schema.FormatListRequest{Name: types.Ptr("mp4")})
	require.NoError(err)
	require.NotNil(resp)
	require.Greater(resp.Count, uint64(0))
	require.Len(resp.Body, int(resp.Count))
	for _, format := range resp.Body {
		require.Equal("mp4", format.Name)
	}
}

func TestListFormats_FilterByExt(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListFormats(ctx, schema.FormatListRequest{Ext: types.Ptr("mp4")})
	require.NoError(err)
	require.NotNil(resp)
	require.Greater(resp.Count, uint64(0))
	for _, format := range resp.Body {
		require.Contains(format.Ext, "mp4")
	}
}

func TestListFormats_FilterByType(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListFormats(ctx, schema.FormatListRequest{Type: types.Ptr("video/mp4")})
	require.NoError(err)
	require.NotNil(resp)
	require.Greater(resp.Count, uint64(0))
	for _, format := range resp.Body {
		require.Contains(format.Type, "video/mp4")
	}
}

func TestListFormats_FilterNoMatch(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListFormats(ctx, schema.FormatListRequest{Name: types.Ptr("not-a-real-format")})
	require.NoError(err)
	require.NotNil(resp)
	require.Equal(uint64(0), resp.Count)
	require.Empty(resp.Body)
}

func TestListFormats_NoFilterReturnsMultiple(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListFormats(ctx, schema.FormatListRequest{})
	require.NoError(err)
	require.NotNil(resp)
	require.Greater(resp.Count, uint64(1))
}
