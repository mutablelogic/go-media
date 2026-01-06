package ffmpeg

import (
	"encoding/json"
	"fmt"
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

type (
	AVFilterContext C.AVFilterContext
	AVFilter        C.AVFilter
	AVFilterFlag    C.int
	AVFilterGraph   C.AVFilterGraph
	AVFilterInOut   C.AVFilterInOut
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	AVFILTER_FLAG_NONE                      AVFilterFlag = 0
	AVFILTER_FLAG_DYNAMIC_INPUTS            AVFilterFlag = C.AVFILTER_FLAG_DYNAMIC_INPUTS
	AVFILTER_FLAG_DYNAMIC_OUTPUTS           AVFilterFlag = C.AVFILTER_FLAG_DYNAMIC_OUTPUTS
	AVFILTER_FLAG_SLICE_THREADS             AVFilterFlag = C.AVFILTER_FLAG_SLICE_THREADS
	AVFILTER_FLAG_METADATA_ONLY             AVFilterFlag = C.AVFILTER_FLAG_METADATA_ONLY
	AVFILTER_FLAG_HWDEVICE                  AVFilterFlag = C.AVFILTER_FLAG_HWDEVICE
	AVFILTER_FLAG_SUPPORT_TIMELINE          AVFilterFlag = C.AVFILTER_FLAG_SUPPORT_TIMELINE
	AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC  AVFilterFlag = C.AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC
	AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL AVFilterFlag = C.AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL
	AVFILTER_FLAG_MAX                                    = AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVFilter) MarshalJSON() ([]byte, error) {
	type j struct {
		Name        string       `json:"name"`
		Description string       `json:"description"`
		Flags       AVFilterFlag `json:"flags,omitzero"`
		Inputs      uint         `json:"num_inputs,omitempty"`
		Outputs     uint         `json:"num_outputs,omitempty"`
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

func (ctx *AVFilterInOut) MarshalJSON() ([]byte, error) {
	type j struct {
		Name   string           `json:"name"`
		Filter *AVFilterContext `json:"filter"`
		Pad    int              `json:"pad"`
	}
	if ctx == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(j{
		Name:   ctx.Name(),
		Filter: ctx.Filter(),
		Pad:    ctx.Pad(),
	})
}

func (ctx *AVFilterGraph) MarshalJSON() ([]byte, error) {
	type j struct {
		Graph string `json:"graph"`
	}
	if ctx == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(j{
		Graph: AVFilterGraph_dump(ctx),
	})
}

func (ctx *AVFilterGraph) String() string {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (ctx *AVFilter) String() string {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (ctx *AVFilterInOut) String() string {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
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

// Free a filter context. This will also remove the filter from the graph's list of filters.
func AVFilterContext_free(ctx *AVFilterContext) {
	C.avfilter_free((*C.AVFilterContext)(ctx))
}

////////////////////////////////////////////////////////////////////////////////
// AVFilterFlag

func (v AVFilterFlag) Is(f AVFilterFlag) bool {
	return v&f == f
}

func (v AVFilterFlag) String() string {
	if v == AVFILTER_FLAG_NONE {
		return v.FlagString()
	}
	str := ""
	for i := AVFilterFlag(C.int(1)); i <= AVFILTER_FLAG_MAX; i <<= 1 {
		if v&i == i {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (v AVFilterFlag) FlagString() string {
	switch v {
	case AVFILTER_FLAG_NONE:
		return "AVFILTER_FLAG_NONE"
	case AVFILTER_FLAG_DYNAMIC_INPUTS:
		return "AVFILTER_FLAG_DYNAMIC_INPUTS"
	case AVFILTER_FLAG_DYNAMIC_OUTPUTS:
		return "AVFILTER_FLAG_DYNAMIC_OUTPUTS"
	case AVFILTER_FLAG_SLICE_THREADS:
		return "AVFILTER_FLAG_SLICE_THREADS"
	case AVFILTER_FLAG_METADATA_ONLY:
		return "AVFILTER_FLAG_METADATA_ONLY"
	case AVFILTER_FLAG_HWDEVICE:
		return "AVFILTER_FLAG_HWDEVICE"
	case AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC:
		return "AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC"
	case AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL:
		return "AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL"
	default:
		return fmt.Sprintf("AVFilterFlag(0x%08X)", uint32(v))
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVFilterInOut

func (c *AVFilterInOut) Name() string {
	return C.GoString(c.name)
}

func (c *AVFilterInOut) Filter() *AVFilterContext {
	return (*AVFilterContext)(c.filter_ctx)
}

func (c *AVFilterInOut) Pad() int {
	return int(c.pad_idx)
}

func (c *AVFilterInOut) Next() *AVFilterInOut {
	return (*AVFilterInOut)(c.next)
}

func (c *AVFilterInOut) SetNext(next *AVFilterInOut) {
	c.next = (*C.AVFilterInOut)(next)
}
