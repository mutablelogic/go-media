package media_test

import (
	// Import namespaces
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

func Test_manager_001(t *testing.T) {
	assert := assert.New(t)

	manager := NewManager()
	assert.NotNil(manager)

	formats := manager.InputFormats()
	assert.NotNil(formats)
	t.Log(formats)
}

func Test_manager_002(t *testing.T) {
	assert := assert.New(t)

	manager := NewManager()
	assert.NotNil(manager)

	formats := manager.OutputFormats()
	assert.NotNil(formats)
	t.Log(formats)
}
