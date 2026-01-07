package ffmpeg

import (
	"encoding/json"
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
// TYPES

type AVFilter C.AVFilter

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVFilter) MarshalJSON() ([]byte, error) {
	type j struct {
		Name        string       `json:"name"`
		Description string       `json:"description"`
		Flags       AVFilterFlag `json:"flags,omitempty"`
		Inputs      uint         `json:"num_inputs"`
		Outputs     uint         `json:"num_outputs"`
	}
	if ctx == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(j{
		Name:        ctx.Name(),
		Description: ctx.Description(),
		Flags:       ctx.Flags(),
		Inputs:      ctx.NumInputs(),
		Outputs:     ctx.NumOutputs(),
	})
}

func (ctx *AVFilter) String() string {
	return marshalToString(ctx)
}

////////////////////////////////////////////////////////////////////////////////
// AVFilter

func (c *AVFilter) Name() string {
	return C.GoString(c.name)
}

func (c *AVFilter) Description() string {
	return C.GoString(c.description)
}

func (c *AVFilter) Flags() AVFilterFlag {
	return AVFilterFlag(c.flags)
}

func (c *AVFilter) NumInputs() uint {
	return AVFilter_inputs(c)
}

func (c *AVFilter) NumOutputs() uint {
	return AVFilter_outputs(c)
}

func (c *AVFilter) PrivClass() *AVClass {
	return (*AVClass)(c.priv_class)
}

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
