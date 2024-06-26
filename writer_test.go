package media_test

import (
	"context"
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
	filename := filepath.Join(t.TempDir(), t.Name()+".mp3")
	stream, err := manager.AudioParameters("mono", "fltp", 22050)
	if !assert.NoError(err) {
		t.SkipNow()
	}

	writer, err := manager.Create(filename, nil, nil, stream)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer writer.Close()

	t.Log(writer, "=>", filename)

	// Perform muxing of packets
	writer.Mux(context.Background(), func(stream int) (Packet, error) {
		t.Log("Muxing packet for stream", stream)
		return nil, nil
	})
}
