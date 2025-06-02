package avfilter_test

import (
	"testing"

	// Packages
	avfilter "github.com/mutablelogic/go-media/pkg/avfilter"
	assert "github.com/stretchr/testify/assert"
)

func Test_graph_001(t *testing.T) {
	assert := assert.New(t)

	graph := avfilter.NewGraph()
	assert.NotNil(graph)
	assert.NoError(graph.Close())
}

func Test_graph_002(t *testing.T) {
	assert := assert.New(t)

	graph, err := avfilter.ParseGraph("[a]null[b]")
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(graph)
	t.Log(graph)
	assert.NoError(graph.Close())
}
