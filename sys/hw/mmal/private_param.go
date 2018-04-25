/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package mmal

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
    #cgo CFLAGS: -I/opt/vc/include
	#cgo LDFLAGS: -L/opt/vc/lib -lmmal -lmmal_core -lmmal_util
	#include <interface/mmal/mmal.h>
	#include <interface/mmal/util/mmal_util.h>
	#include <interface/mmal/util/mmal_util_params.h>
*/
import "C"
import (
	"reflect"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - PARAMETERS

func mmal_port_param_alloc_get(handle portHandle, param, size uint32) (paramHandle, error) {
	var err C.MMAL_STATUS_T
	if handle := C.mmal_port_parameter_alloc_get(handle, C.uint(param), C.uint(size), &err); status(err) != MMAL_SUCCESS {
		return handle, status(err)
	} else {
		return handle, nil
	}
}

func mmal_port_param_free(handle paramHandle) {
	C.mmal_port_parameter_free(handle)
}

func mmal_port_param_get_bool(handle portHandle, param uint32) (bool, error) {
	var value C.MMAL_BOOL_T
	if err := C.mmal_port_parameter_get_boolean(handle, C.uint(param), &value); status(err) != MMAL_SUCCESS {
		return false, status(err)
	} else {
		return uint32(value) != 0, nil
	}
}

func mmal_port_param_set_bool(handle portHandle, param uint32, value bool) error {
	var value_ C.MMAL_BOOL_T
	if value {
		value_ = C.MMAL_BOOL_T(1)
	} else {
		value_ = C.MMAL_BOOL_T(0)
	}
	if err := C.mmal_port_parameter_set_boolean(handle, C.uint(param), value_); status(err) != MMAL_SUCCESS {
		return status(err)
	} else {
		return nil
	}
}

func mmal_param_get_array_uint32(handle paramHandle) []C.uint32_t {
	var array []C.uint32_t

	// Data and length of the arrat
	data := uintptr(unsafe.Pointer(handle)) + unsafe.Sizeof(*handle)
	len := (uintptr(handle.size) - unsafe.Sizeof(*handle)) / C.sizeof_uint32_t

	// Make a fake slice
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&array)))
	sliceHeader.Cap = int(len)
	sliceHeader.Len = int(len)
	sliceHeader.Data = data

	// Return the array
	return array
}
