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

func TestListSampleFormatsAll(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListSampleFormats(ctx, schema.SampleFormatListRequest{})
	require.NoError(err)
	require.NotEmpty(resp)
}

func TestListSampleFormatsFilterName(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	all, err := mgr.ListSampleFormats(ctx, schema.SampleFormatListRequest{})
	require.NoError(err)
	require.NotEmpty(all)

	name := all[0].Name
	resp, err := mgr.ListSampleFormats(ctx, schema.SampleFormatListRequest{Name: types.Ptr(name)})
	require.NoError(err)
	require.NotEmpty(resp)
	for _, sf := range resp {
		require.Equal(name, sf.Name)
	}
}

func TestListSampleFormatsFilterIsPlanar(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListSampleFormats(ctx, schema.SampleFormatListRequest{IsPlanar: types.Ptr(true)})
	require.NoError(err)
	require.NotEmpty(resp)
	for _, sf := range resp {
		require.True(sf.IsPlanar)
	}
}

func TestListSampleFormatsFilterNoMatch(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListSampleFormats(ctx, schema.SampleFormatListRequest{Name: types.Ptr("not-a-real-sample-format")})
	require.NoError(err)
	require.Empty(resp)
}
