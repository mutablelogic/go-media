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

func TestListChannelLayoutsAll(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListChannelLayouts(ctx, schema.ChannelLayoutListRequest{})
	require.NoError(err)
	require.NotEmpty(resp)
}

func TestListChannelLayoutsFilterName(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	all, err := mgr.ListChannelLayouts(ctx, schema.ChannelLayoutListRequest{})
	require.NoError(err)
	require.NotEmpty(all)

	name := all[0].Name
	resp, err := mgr.ListChannelLayouts(ctx, schema.ChannelLayoutListRequest{Name: types.Ptr(name)})
	require.NoError(err)
	require.NotEmpty(resp)
	for _, cl := range resp {
		require.Equal(name, cl.Name)
	}
}

func TestListChannelLayoutsFilterNumChannels(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListChannelLayouts(ctx, schema.ChannelLayoutListRequest{NumChannels: types.Ptr(uint64(2))})
	require.NoError(err)
	require.NotEmpty(resp)
	for _, cl := range resp {
		require.Equal(2, cl.NumChannels)
	}
}

func TestListChannelLayoutsFilterNoMatch(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListChannelLayouts(ctx, schema.ChannelLayoutListRequest{Name: types.Ptr("not-a-real-channel-layout")})
	require.NoError(err)
	require.Empty(resp)
}
