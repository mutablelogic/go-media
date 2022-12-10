package media_test

import (
	"testing"

	// Package imports
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	//. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/pkg/media"
)

const (
	SAMPLE_MP4 = "../../etc/sample.mp4"
)

func Test_manager_000(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	assert.NotNil(mgr)
	assert.NoError(mgr.Close())
}

func Test_manager_001(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	assert.NotNil(mgr)
	media, err := mgr.OpenFile(SAMPLE_MP4)
	assert.NoError(err)
	t.Log(media)
	assert.NoError(media.Close())
	assert.NoError(mgr.Close())
}
