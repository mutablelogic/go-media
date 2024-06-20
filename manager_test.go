package media_test

import (
	// Import namespaces
	"os"
	"testing"

	// Package imports
	"github.com/djthorpe/go-tablewriter"
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

func Test_manager_001(t *testing.T) {
	assert := assert.New(t)

	manager := NewManager()
	assert.NotNil(manager)

	formats := manager.InputFormats(ANY)
	assert.NotNil(formats)
	t.Log(formats)
}

func Test_manager_002(t *testing.T) {
	assert := assert.New(t)

	manager := NewManager()
	assert.NotNil(manager)

	formats := manager.OutputFormats(ANY)
	assert.NotNil(formats)
	t.Log(formats)
}

func Test_manager_003(t *testing.T) {
	assert := assert.New(t)

	manager := NewManager()
	assert.NotNil(manager)

	version := manager.Version()
	assert.NotNil(version)

	tablewriter.New(os.Stderr, tablewriter.OptHeader(), tablewriter.OptOutputText()).Write(version)
}
