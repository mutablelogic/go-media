package media_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Package imports
	assert "github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/pkg/media"
)

const (
	SAMPLE_MP4 = "../../etc/sample.mp4"
	SAMPLE_HLS = "http://a.files.bbci.co.uk/media/live/manifesto/audio/simulcast/hls/nonuk/sbr_vlow/ak/bbc_radio_fourfm.m3u8"
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
	media, err := mgr.OpenFile(SAMPLE_MP4, nil)
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
	err = media.Set(MEDIA_KEY_CREATED, time.Now())
	assert.NoError(err)
	t.Log(media)
	assert.NoError(media.Close())
	assert.NoError(mgr.Close())
}

func Test_manager_003(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	assert.NotNil(mgr)

	formats := mgr.MediaFormats(MEDIA_FLAG_NONE)
	assert.True(len(formats) > 0)
	for _, format := range formats {
		t.Log(format)
	}
	assert.NoError(mgr.Close())
}

func Test_manager_004(t *testing.T) {
	assert := assert.New(t)
	mgr := New()
	assert.NotNil(mgr)
	media, err := mgr.OpenURL(SAMPLE_HLS, nil)
	assert.NoError(err)
	t.Log(media)
	assert.NoError(media.Close())
	assert.NoError(mgr.Close())
}
