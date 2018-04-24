/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package rpi

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
    #cgo CFLAGS: -I/opt/vc/include -I/opt/vc/include/interface/mmal
    #cgo LDFLAGS:  -L/opt/vc/lib -lmmal -lmmal_components -lmmal_core
	#include <mmal.h>
*/
import "C"
import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	mmalComponentHandle (*C.MMAL_COMPONENT_T)
	mmalPortHandle      (*C.MMAL_PORT_T)
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func mmal_component_create(name string, handle *mmalComponentHandle) error {
	cName := C.CString(name)
	var cHandle (*C.MMAL_COMPONENT_T)
	defer C.free(unsafe.Pointer(cName))
	if status := mmalStatus(C.mmal_component_create(cName, &cHandle)); status == MMAL_SUCCESS {
		*handle = mmalComponentHandle(cHandle)
		return nil
	} else {
		return status
	}
}

func mmal_component_destroy(handle mmalComponentHandle) error {
	if status := mmalStatus(C.mmal_component_destroy(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_component_enable(handle mmalComponentHandle) error {
	if status := mmalStatus(C.mmal_component_enable(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_component_disable(handle mmalComponentHandle) error {
	if status := mmalStatus(C.mmal_component_disable(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}
