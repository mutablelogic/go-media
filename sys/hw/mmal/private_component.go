/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package mmal

import (
	"reflect"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
	#cgo CFLAGS: -I/opt/vc/include
	#include <interface/mmal/mmal.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - COMPONENTS

func mmal_component_create(name string, handle *componentHandle) error {
	cName := C.CString(name)
	var cHandle (*C.MMAL_COMPONENT_T)
	defer C.free(unsafe.Pointer(cName))
	if status := status(C.mmal_component_create(cName, &cHandle)); status == MMAL_SUCCESS {
		*handle = componentHandle(cHandle)
		return nil
	} else {
		return status
	}
}

func mmal_component_destroy(handle componentHandle) error {
	if status := status(C.mmal_component_destroy(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_component_enable(handle componentHandle) error {
	if status := status(C.mmal_component_enable(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_component_disable(handle componentHandle) error {
	if status := status(C.mmal_component_disable(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_component_acquire(handle componentHandle) {
	C.mmal_component_acquire(handle)
}

func mmal_component_release(handle componentHandle) error {
	if status := status(C.mmal_component_release(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_component_name(handle componentHandle) string {
	return C.GoString(handle.name)
}

func mmal_component_id(handle componentHandle) uint32 {
	return uint32(handle.id)
}

func mmal_component_is_enabled(handle componentHandle) bool {
	return handle.is_enabled != 0
}

func mmal_component_control_port(handle componentHandle) portHandle {
	return handle.control
}

func mmal_component_input_port(handle componentHandle, index uint) portHandle {
	if index >= uint(handle.input_num) {
		return nil
	} else {
		return mmal_component_port_at_index(handle.input, uint(handle.input_num), index)
	}
}

func mmal_component_output_port(handle componentHandle, index uint) portHandle {
	if index >= uint(handle.output_num) {
		return nil
	} else {
		return mmal_component_port_at_index(handle.output, uint(handle.output_num), index)
	}
}

func mmal_component_clock_port(handle componentHandle, index uint) portHandle {
	if index >= uint(handle.clock_num) {
		return nil
	} else {
		return mmal_component_port_at_index(handle.clock, uint(handle.clock_num), index)
	}
}

func mmal_component_port_at_index(array **C.MMAL_PORT_T, num, index uint) portHandle {
	var handles []portHandle
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&handles)))
	sliceHeader.Cap = int(num)
	sliceHeader.Len = int(num)
	sliceHeader.Data = uintptr(unsafe.Pointer(array))
	return handles[index]

}

func mmal_component_num_input_port(handle componentHandle) uint {
	return uint(handle.input_num)
}

func mmal_component_num_output_port(handle componentHandle) uint {
	return uint(handle.output_num)
}

func mmal_component_num_clock_port(handle componentHandle) uint {
	return uint(handle.clock_num)
}

func mmal_component_num_port(handle componentHandle) uint {
	return uint(handle.port_num)
}
