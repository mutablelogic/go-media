package ffmpeg

import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavfilter
#include <libavfilter/avfilter.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func AVFilterGraph_alloc() *AVFilterGraph {
	return (*AVFilterGraph)(C.avfilter_graph_alloc())
}

func AVFilterGraph_free(graph *AVFilterGraph) {
	ctx := (*C.AVFilterGraph)(graph)
	C.avfilter_graph_free(&ctx)
}

// Check validity and configure all the links and formats in the graph.
func AVFilterGraph_config(graph *AVFilterGraph) error {
	if err := AVError(C.avfilter_graph_config((*C.AVFilterGraph)(graph), nil)); err != 0 {
		return fmt.Errorf("avfilter_graph_config: %w", err)
	}
	return nil
}

// Dump the graph out.
func AVFilterGraph_dump(graph *AVFilterGraph) string {
	if graph == nil {
		return ""
	}

	// Check if the graph is valid
	if err := AVFilterGraph_config(graph); err != nil {
		return err.Error()
	}

	// Dump the graph
	cStr := C.avfilter_graph_dump((*C.AVFilterGraph)(graph), nil)
	defer C.free(unsafe.Pointer(cStr))

	// Return the string
	return C.GoString(cStr)
}

// Allocates and initializes a filter in a single step.
// The filter instance is created from the filter and inited with the parameter args.
func AVFilterGraph_create_filter(graph *AVFilterGraph, filter *AVFilter, name, args string) (*AVFilterContext, error) {
	var ctx *C.AVFilterContext
	graph_ctx := (*C.AVFilterGraph)(graph)
	cName, cArgs := C.CString(name), C.CString(args)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cArgs))
	if err := AVError(C.avfilter_graph_create_filter(&ctx, (*C.AVFilter)(filter), cName, cArgs, nil, graph_ctx)); err != 0 {
		return nil, fmt.Errorf("avfilter_graph_create_filter: %w", err)
	}
	return (*AVFilterContext)(ctx), nil
}

// Add a graph described by a string to a graph,returning inputs and outputs. Will return an error
// if not all inputs and outputs are specified. The inputs and outputs should be freed
// with AVFilterInOut_free() when no longer needed.
func AVFilterGraph_parse(graph *AVFilterGraph, filters string) ([]*AVFilterInOut, []*AVFilterInOut, error) {
	var ins, outs *AVFilterInOut

	cFilters := C.CString(filters)
	defer C.free(unsafe.Pointer(cFilters))
	if err := AVError(C.avfilter_graph_parse_ptr((*C.AVFilterGraph)(graph), cFilters, (**C.AVFilterInOut)(unsafe.Pointer(&ins)), (**C.AVFilterInOut)(unsafe.Pointer(&outs)), nil)); err != 0 {
		AVFilterInOut_free(ins)
		AVFilterInOut_free(outs)
		return nil, nil, fmt.Errorf("avfilter_graph_parse: %w", err)
	}

	// TODO: If ins is 0 and outs is 0, we return the linked list of inputs and outputs
	// Or else we try again with the ins and outs

	// Return success
	return AVFilterInOut_list(ins), AVFilterInOut_list(outs), nil
}
