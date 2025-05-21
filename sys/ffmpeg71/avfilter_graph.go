package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavfilter
#include <libavfilter/avfilter.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func AVFilterGraph_alloc() *AVFilterGraph {
	return (*AVFilterGraph)(C.avfilter_graph_alloc())
}

func AVFilterGraph_free(graph *AVFilterGraph) {
	ctx := (*C.AVFilterGraph)(graph)
	C.avfilter_graph_free(&ctx)
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
