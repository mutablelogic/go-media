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

func TestListPixelFormatsAll(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListPixelFormats(ctx, schema.PixelFormatListRequest{})
	require.NoError(err)
	require.NotEmpty(resp)
}

func TestListPixelFormatsFilterName(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	all, err := mgr.ListPixelFormats(ctx, schema.PixelFormatListRequest{})
	require.NoError(err)
	require.NotEmpty(all)

	name := all[0].Name
	resp, err := mgr.ListPixelFormats(ctx, schema.PixelFormatListRequest{Name: types.Ptr(name)})
	require.NoError(err)
	require.NotEmpty(resp)
	for _, pf := range resp {
		require.Equal(name, pf.Name)
	}
}

func TestListPixelFormatsFilterNumPlanes(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListPixelFormats(ctx, schema.PixelFormatListRequest{NumPlanes: types.Ptr(uint64(1))})
	require.NoError(err)
	require.NotEmpty(resp)
	for _, pf := range resp {
		require.Equal(1, pf.NumPlanes)
	}
}

func TestListPixelFormatsFilterNoMatch(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListPixelFormats(ctx, schema.PixelFormatListRequest{Name: types.Ptr("not-a-real-pixel-format")})
	require.NoError(err)
	require.Empty(resp)
}
