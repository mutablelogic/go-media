package manager_test

import (
	"testing"

	// Packages
	test "github.com/mutablelogic/go-media/profile/test"
	require "github.com/stretchr/testify/require"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func TestManager(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	require.NotNil(mgr)
	require.NotNil(ctx)
}
