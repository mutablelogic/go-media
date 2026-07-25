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

func TestListDevicesAll(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	// Device availability is host-dependent (CI runners may have none), so
	// this only checks the call succeeds and returns internally-consistent data.
	resp, err := mgr.ListDevices(ctx, schema.DeviceListRequest{})
	require.NoError(err)
	for _, d := range resp {
		require.NotEmpty(d.Format)
		require.NotEmpty(d.Name)
		require.True(d.IsInput || d.IsOutput)
	}
}

func TestListDevicesFilterFormat(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	all, err := mgr.ListDevices(ctx, schema.DeviceListRequest{})
	require.NoError(err)
	if len(all) == 0 {
		t.Skip("no devices available on this host")
	}

	format := all[0].Format
	resp, err := mgr.ListDevices(ctx, schema.DeviceListRequest{Format: types.Ptr(format)})
	require.NoError(err)
	require.NotEmpty(resp)
	for _, d := range resp {
		require.Equal(format, d.Format)
	}
}

func TestListDevicesFilterNoMatch(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	resp, err := mgr.ListDevices(ctx, schema.DeviceListRequest{Name: types.Ptr("not-a-real-device")})
	require.NoError(err)
	require.Empty(resp)
}
