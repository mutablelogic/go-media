package ffmpeg_test

import (
	"testing"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	assert "github.com/stretchr/testify/assert"
)

func Test_manager_001(t *testing.T) {
	assert := assert.New(t)

	// Create a manager
	manager, err := ffmpeg.NewManager(ffmpeg.OptLog(true, func(v string) {
		t.Log(v)
	}))
	if !assert.NoError(err) {
		t.FailNow()
	}

	manager.Infof("INFO test")
	manager.Warningf("WARNING test")
	manager.Errorf("ERROR test")
}

func Test_manager_002(t *testing.T) {
	assert := assert.New(t)

	// Create a manager
	manager, err := ffmpeg.NewManager(ffmpeg.OptLog(true, func(v string) {
		t.Log(v)
	}))
	if !assert.NoError(err) {
		t.FailNow()
	}

	for _, v := range manager.Version() {
		t.Log(v)
	}
}
