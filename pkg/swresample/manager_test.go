package swresample_test

import (
	"testing"

	// Package imports
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/pkg/swresample"
)

func Test_swresample_000(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	assert.NotNil(mgr)
	assert.NoError(mgr.Close())
}
