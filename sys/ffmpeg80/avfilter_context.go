package ffmpeg

import "encoding/json"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavfilter
#include <libavfilter/avfilter.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type AVFilterContext C.AVFilterContext

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVFilterContext) MarshalJSON() ([]byte, error) {
	type j struct {
		Name    string    `json:"name"`
		Filter  *AVFilter `json:"filter"`
		Inputs  uint      `json:"num_inputs,omitempty"`
		Outputs uint      `json:"num_outputs,omitempty"`
	}
	if ctx == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(j{
		Name:    ctx.Name(),
		Filter:  ctx.Filter(),
		Inputs:  ctx.NumInputs(),
		Outputs: ctx.NumOutputs(),
	})
}

func (ctx *AVFilterContext) String() string {
	return marshalToString(ctx)
}

////////////////////////////////////////////////////////////////////////////////
// AVFilterContext

func (c *AVFilterContext) Name() string {
	return C.GoString(c.name)
}

func (c *AVFilterContext) Filter() *AVFilter {
	return (*AVFilter)(c.filter)
}

func (c *AVFilterContext) NumInputs() uint {
	return uint(c.nb_inputs)
}

func (c *AVFilterContext) NumOutputs() uint {
	return uint(c.nb_outputs)
}

// Link two filters together.
func AVFilterContext_link(src *AVFilterContext, srcpad uint, dst *AVFilterContext, dstpad uint) error {
	if ret := C.avfilter_link(
		(*C.AVFilterContext)(src), C.uint(srcpad),
		(*C.AVFilterContext)(dst), C.uint(dstpad),
	); ret < 0 {
		return AVError(ret)
	}
	return nil
}

// Free a filter context. This will also remove the filter from the graph's list of filters.
func AVFilterContext_free(ctx *AVFilterContext) {
	C.avfilter_free((*C.AVFilterContext)(ctx))
}
