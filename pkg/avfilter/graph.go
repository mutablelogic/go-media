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
	in  []*ff.AVFilterInOut
	out []*ff.AVFilterInOut
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

// Parse a graph description and return it.
func ParseGraph(desc string) (*Graph, error) {
	graph := NewGraph()
	if graph == nil {
		return nil, media.ErrInternalError.With("failed to allocate filter graph")
	}

	// Parse the graph, and set the inputs and outputs
	in, out, err := ff.AVFilterGraph_parse(graph.ctx, desc)
	if err != nil {
		return nil, graph.Close()
	} else {
		graph.in = in
		graph.out = out
	}

	// Validate the graph
	if err := ff.AVFilterGraph_config(graph.ctx); err != nil {
		return nil, graph.Close()
	}

	// Return the graph
	return graph, nil
}

// Frees the filter graph and all its resources.
func (g *Graph) Close() error {
	// Free the inputs and outputs
	if g.in != nil {
		ff.AVFilterInOut_list_free(g.in)
	}
	if g.out != nil {
		ff.AVFilterInOut_list_free(g.out)
	}
	if g.ctx != nil {
		ff.AVFilterGraph_free(g.ctx)
		g.ctx = nil
	}
	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS
