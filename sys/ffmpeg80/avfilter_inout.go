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

type AVFilterInOut C.AVFilterInOut

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

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

func (ctx *AVFilterInOut) String() string {
	return marshalToString(ctx)
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

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocate a single AVFilterInOut entry, with name, filter context and pad index.
func AVFilterInOut_alloc(name string, filter *AVFilterContext, pad int) *AVFilterInOut {
	inout := (*C.AVFilterInOut)(C.avfilter_inout_alloc())
	if inout == nil {
		return nil
	}
	inout.name = C.CString(name)
	inout.filter_ctx = (*C.AVFilterContext)(filter)
	inout.pad_idx = C.int(pad)
	inout.next = nil
	return (*AVFilterInOut)(inout)
}

// Free a single AVFilterInOut entry, including its name.
func AVFilterInOut_free(inout *AVFilterInOut) {
	ctx := (*C.AVFilterInOut)(inout)
	C.avfilter_inout_free(&ctx)
}

// Link an array of AVFilterInOut entries together, and return the first entry.
// If the array is empty, or the first entry is nil, nil is returned. A nil entry
// after the first acts as a terminator for the chain and subsequent entries are ignored.
func AVFilterInOut_link(inout ...*AVFilterInOut) *AVFilterInOut {
	if len(inout) == 0 {
		return nil
	}
	// If the first element is nil, there is no valid head to return.
	if inout[0] == nil {
		return nil
	}
	for i := 0; i < len(inout)-1; i++ {
		// A nil entry terminates the chain; do not attempt to dereference it.
		if inout[i] == nil {
			break
		}
		inout[i].SetNext(inout[i+1])
	}
	// Set the last element's next to nil, but only if it's not nil itself.
	if inout[len(inout)-1] != nil {
		inout[len(inout)-1].SetNext(nil)
	}
	return inout[0]
}

// Return an array of AVFilterInOut entries, given the first entry.
// Returns nil if the first entry is nil.
func AVFilterInOut_list(head *AVFilterInOut) []*AVFilterInOut {
	if head == nil {
		return nil
	}
	var result []*AVFilterInOut
	for inout := head; inout != nil; inout = inout.Next() {
		result = append(result, inout)
	}
	return result
}

// Free an array of AVFilterInOut entries, given the first entry.
func AVFilterInOut_list_free(list []*AVFilterInOut) {
	if len(list) == 0 {
		return
	}
	AVFilterInOut_free(list[0])
}
