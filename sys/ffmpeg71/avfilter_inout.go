package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavfilter
#include <libavfilter/avfilter.h>
*/
import "C"

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

// Free the entire linked list of AVFilterInOut entries starting from the given entry.
// This calls avfilter_inout_free() which frees the supplied list and all entries
// linked via the 'next' field. Only call this on the head of a list or on an unlinked
// entry to avoid double-free issues.
func AVFilterInOut_free(inout *AVFilterInOut) {
	ctx := (*C.AVFilterInOut)(inout)
	C.avfilter_inout_free(&ctx)
}

// Link an array of AVFilterInOut entries together, and return the first entry.
// If the array is empty, or any entry is nil, nil is returned.
func AVFilterInOut_link(inout ...*AVFilterInOut) *AVFilterInOut {
	if len(inout) == 0 {
		return nil
	}
	// If the first element is nil, there is no valid head to return.
	if inout[0] == nil {
		return nil
	}
	for i := 0; i < len(inout)-1; i++ {
		// A nil entry after the first acts as a terminator for the chain.
		// Set the previous entry's next to nil and stop processing.
		if inout[i+1] == nil {
			inout[i].SetNext(nil)
			break
		}
		inout[i].SetNext(inout[i+1])
	}
	// Set the last element's next to nil if we processed all entries.
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
