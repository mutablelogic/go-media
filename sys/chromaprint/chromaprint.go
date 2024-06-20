package chromaprint

import (
	"fmt"
	"time"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libchromaprint
#include <chromaprint.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Error         int
	AlgorithmType C.int
	Context       C.ChromaprintContext
)

////////////////////////////////////////////////////////////////////////////////
// CONSTS

const (
	ALGORITHM_TEST1   AlgorithmType = C.CHROMAPRINT_ALGORITHM_TEST1
	ALGORITHM_TEST2   AlgorithmType = C.CHROMAPRINT_ALGORITHM_TEST2
	ALGORITHM_TEST3   AlgorithmType = C.CHROMAPRINT_ALGORITHM_TEST3
	ALGORITHM_TEST4   AlgorithmType = C.CHROMAPRINT_ALGORITHM_TEST4
	ALGORITHM_DEFAULT AlgorithmType = C.CHROMAPRINT_ALGORITHM_DEFAULT
)

const (
	errNone Error = iota
	errStart
	errFeed
	errFinish
	errFingerprint
)

////////////////////////////////////////////////////////////////////////////////
// METHODS

func Version() string {
	return C.GoString(C.chromaprint_get_version())
}

func NewChromaprint(algorithm AlgorithmType) *Context {
	ctx := C.chromaprint_new(C.int(algorithm))
	return (*Context)(ctx)
}

func (context *Context) Free() {
	C.chromaprint_free((*C.ChromaprintContext)(context))
}

func (context *Context) Start(rate, channels int) error {
	if res := C.chromaprint_start((*C.ChromaprintContext)(context), C.int(rate), C.int(channels)); res < 1 {
		return errStart
	}
	return nil
}

func (context *Context) Write(data []byte) error {
	ptr := (*C.int16_t)(unsafe.Pointer(&data[0]))
	if res := C.chromaprint_feed((*C.ChromaprintContext)(context), ptr, C.int(len(data)>>1)); res < 1 {
		return errFeed
	}
	return nil
}

func (context *Context) WritePtr(data uintptr, size int) error {
	ptr := (*C.int16_t)(unsafe.Pointer(data))
	if res := C.chromaprint_feed((*C.ChromaprintContext)(context), ptr, C.int(size)); res < 1 {
		return errFeed
	}
	return nil
}

func (context *Context) Finish() error {
	if res := C.chromaprint_finish((*C.ChromaprintContext)(context)); res < 1 {
		return errFinish
	}
	return nil
}

func (context *Context) Channels() int {
	return int(C.chromaprint_get_num_channels((*C.ChromaprintContext)(context)))
}

func (context *Context) Rate() int {
	return int(C.chromaprint_get_sample_rate((*C.ChromaprintContext)(context)))
}

/* Function not exported
func (this *Context) Algorithm() AlgorithmType {
	ctx := (*C.ChromaprintContext)(this)
	return AlgorithmType(C.chromaprint_get_algorithm(ctx))
}
*/

func (context *Context) Duration() int {
	return int(C.chromaprint_get_item_duration((*C.ChromaprintContext)(context)))
}

func (context *Context) DurationMs() time.Duration {
	return time.Duration((C.chromaprint_get_item_duration_ms((*C.ChromaprintContext)(context)))) * time.Millisecond
}

func (context *Context) Delay() int {
	return int(C.chromaprint_get_delay((*C.ChromaprintContext)(context)))
}

func (context *Context) DelayMs() time.Duration {
	return time.Duration(C.chromaprint_get_delay_ms((*C.ChromaprintContext)(context))) * time.Millisecond
}

func (context *Context) GetFingerprint() (string, error) {
	var ptr (*C.char)
	if res := C.chromaprint_get_fingerprint((*C.ChromaprintContext)(context), &ptr); res < 1 {
		return "", errFingerprint
	}
	defer C.chromaprint_dealloc(unsafe.Pointer(ptr))
	return C.GoString((*C.char)(ptr)), nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (context *Context) String() string {
	str := "<chromaprint.context"
	/*
		if a := this.Algorithm(); a >= 0 {
			str += " algorithm =" + fmt.Sprint(a)
		}
	*/
	if r := context.Rate(); r > 0 {
		str += " sample_rate=" + fmt.Sprint(r)
	}
	if ch := context.Channels(); ch > 0 {
		str += " channels=" + fmt.Sprint(ch)
	}
	if d := context.DurationMs(); d > 0 {
		str += " duration=" + fmt.Sprint(d)
	}
	if d := context.DelayMs(); d > 0 {
		str += " delay=" + fmt.Sprint(d)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e Error) Error() string {
	switch e {
	case errStart:
		return "Chromaprint Start() error"
	case errFeed:
		return "Chromaprint Feed() error"
	case errFinish:
		return "Chromaprint Finish() error"
	case errFingerprint:
		return "Chromaprint Fingerprinting error"
	default:
		return "Unknown Error"
	}
}

func (a AlgorithmType) String() string {
	switch a {
	case ALGORITHM_TEST1:
		return "ALGORITHM_TEST1"
	case ALGORITHM_TEST2:
		return "ALGORITHM_TEST2"
	case ALGORITHM_TEST3:
		return "ALGORITHM_TEST3"
	case ALGORITHM_TEST4:
		return "ALGORITHM_TEST4"
	default:
		return "[?? Invalid AlgorithmType value]"
	}
}
