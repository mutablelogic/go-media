package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavfilter
#include <libavfilter/avfilter.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Iterate over all registered filters.
func AVFilter_iterate(opaque *uintptr) *AVFilter {
	return (*AVFilter)(C.av_filter_iterate((*unsafe.Pointer)(unsafe.Pointer(opaque))))
}

// Get a filter definition matching the given name.
func AVFilter_get_by_name(name string) *AVFilter {
	cStr := C.CString(name)
	defer C.free(unsafe.Pointer(cStr))
	return (*AVFilter)(C.avfilter_get_by_name(cStr))
}

// Return number of input pads for a filter.
func AVFilter_inputs(filter *AVFilter) uint {
	ctx := (*C.AVFilter)(filter)
	return uint(C.avfilter_filter_pad_count(ctx, 0))
}

// Return number of output pads for a filter.
func AVFilter_outputs(filter *AVFilter) uint {
	ctx := (*C.AVFilter)(filter)
	return uint(C.avfilter_filter_pad_count(ctx, 1))
}

// Free a filter context. This will also remove the filter from graph's list of filters.
func AVFilter_free(filter *AVFilterContext) {
	ctx := (*C.AVFilterContext)(filter)
	C.avfilter_free(ctx)
}
