package media_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func Test_manager_002(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	assert.NotNil(mgr)
	path, err := ioutil.TempDir("", "media")
	assert.NoError(err)
	defer os.RemoveAll(path)
	media, err := mgr.CreateFile(filepath.Join(path, "XX.mp4"))
	assert.NoError(err)
	assert.NotNil(media)
	t.Log(media)
	assert.NoError(media.Close())
	assert.NoError(mgr.Close())
}
