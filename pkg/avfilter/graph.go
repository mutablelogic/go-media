package avfilter

import (
	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Graph struct {
	ctx *ff.AVFilterGraph
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Allocates a new filter graph and returns it.
func NewGraph() *Graph {
	graph := new(Graph)
	if ctx := ff.AVFilterGraph_alloc(); ctx == nil {
		return nil
	} else {
		graph.ctx = ctx
	}
	// Return the graph
	return graph
}

// Frees the filter graph and all its resources.
func (g *Graph) Close() error {
	if g.ctx != nil {
		ff.AVFilterGraph_free(g.ctx)
		g.ctx = nil
	}
	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Add a graph described by the input, with optional inputs and outputs
func (g *Graph) Parse(desc string, opts ...Opt) error {
	// TODO
	return media.ErrNotImplemented
}
