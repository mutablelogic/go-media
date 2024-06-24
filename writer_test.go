package media_test

import (
	"path/filepath"
	"strings"
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

func Test_writer_001(t *testing.T) {
	assert := assert.New(t)
	manager, err := NewManager(OptLog(true, func(v string) {
		t.Log(strings.TrimSpace(v))
	}))
	if !assert.NoError(err) {
		t.SkipNow()
	}

	// Write audio file
	filename := filepath.Join(t.TempDir(), t.Name()+".sw")
	stream, err := manager.AudioParameters("mono", "s16", 22050)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	writer, err := manager.Create(filename, nil, nil, stream)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer writer.Close()
	t.Log(writer)
}
