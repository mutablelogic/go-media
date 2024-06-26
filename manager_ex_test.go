//go:build !container

package media_test

import (
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

// These tests do not run in containers

func Test_manager_008(t *testing.T) {
	assert := assert.New(t)

	manager, err := NewManager()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	formats := manager.InputFormats(ANY)
	assert.NotNil(formats)
	for _, format := range formats {
		if format.Type().Is(DEVICE) {
			devices := manager.Devices(format)
			assert.NotNil(devices)
			t.Log(format, devices)
		}
	}
}
