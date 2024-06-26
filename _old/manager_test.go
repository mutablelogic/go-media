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

	manager, err := NewManager()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	formats := manager.InputFormats(ANY)
	assert.NotNil(formats)
	t.Log(formats)
}

func Test_manager_002(t *testing.T) {
	assert := assert.New(t)

	manager, err := NewManager()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	formats := manager.OutputFormats(ANY)
	assert.NotNil(formats)
	t.Log(formats)
}

func Test_manager_004(t *testing.T) {
	assert := assert.New(t)

	manager, err := NewManager()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	channel_layouts := manager.ChannelLayouts()
	assert.NotNil(channel_layouts)

	tablewriter.New(os.Stderr, tablewriter.OptHeader(), tablewriter.OptOutputText()).Write(channel_layouts)
}

func Test_manager_005(t *testing.T) {
	assert := assert.New(t)

	manager, err := NewManager()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	sample_formats := manager.SampleFormats()
	assert.NotNil(sample_formats)

	tablewriter.New(os.Stderr, tablewriter.OptHeader(), tablewriter.OptOutputText()).Write(sample_formats)
}

func Test_manager_006(t *testing.T) {
	assert := assert.New(t)

	manager, err := NewManager()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	pixel_formats := manager.PixelFormats()
	assert.NotNil(pixel_formats)

	tablewriter.New(os.Stderr, tablewriter.OptHeader(), tablewriter.OptOutputText()).Write(pixel_formats)
}

func Test_manager_007(t *testing.T) {
	assert := assert.New(t)

	manager, err := NewManager()
	if !assert.NoError(err) {
		t.SkipNow()
	}

	codecs := manager.Codecs()
	assert.NotNil(codecs)

	tablewriter.New(os.Stderr, tablewriter.OptHeader(), tablewriter.OptOutputText()).Write(codecs)
}
