package manager_test

import (
	"strings"
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
		// Name is comma-joined when a format has more than one (e.g. the mp4
		// demuxer is "mov,mp4,m4a,3gp,3g2,mj2"), so the filter matches any one
		// token rather than the field as a whole - same as Ext/Type below.
		tokens := strings.Split(format.Name, ",")
		found := false
		for _, token := range tokens {
			if strings.EqualFold(strings.TrimSpace(token), "mp4") {
				found = true
				break
			}
		}
		require.True(found, "format name %q does not contain token \"mp4\"", format.Name)
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

func TestListFormats_ExcludesDevices(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListFormats(ctx, schema.FormatListRequest{})
	require.NoError(err)
	require.NotNil(resp)
	for _, format := range resp.Body {
		devices, err := mgr.ListDevices(ctx, schema.DeviceListRequest{Format: types.Ptr(format.Name)})
		require.NoError(err)
		require.Empty(devices, "format %q should not be a device format", format.Name)
	}
}
