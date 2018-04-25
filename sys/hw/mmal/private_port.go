/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package mmal

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
    #cgo CFLAGS: -I/opt/vc/include
	#include <interface/mmal/mmal.h>

	void mmal_port_callback(MMAL_PORT_T* port, MMAL_BUFFER_HEADER_T* buffer);
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - PORTS

func mmal_port_enable(handle portHandle) error {
	if status := status(C.mmal_port_enable(handle, C.MMAL_PORT_BH_CB_T(C.mmal_port_callback))); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_port_disable(handle portHandle) error {
	if status := status(C.mmal_port_disable(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_port_flush(handle portHandle) error {
	if status := status(C.mmal_port_flush(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_port_name(handle portHandle) string {
	return C.GoString(handle.name)
}

func mmal_port_type(handle portHandle) portType {
	return portType(handle._type)
}

func mmal_port_index(handle portHandle) uint {
	return uint(handle.index)
}

func mmal_port_is_enabled(handle portHandle) bool {
	return (handle.is_enabled != 0)
}

func mmal_port_capabilities(handle portHandle) uint32 {
	return uint32(handle.capabilities)
}

func mmal_port_disconnect(handle portHandle) error {
	if status := status(C.mmal_port_disconnect(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_port_connect(this, other portHandle) error {
	if status := status(C.mmal_port_connect(this, other)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_port_format_commit(handle portHandle) error {
	if status := status(C.mmal_port_format_commit(handle)); status == MMAL_SUCCESS {
		return nil
	} else {
		return status
	}
}

func mmal_port_format(handle portHandle) esFormatHandle {
	return handle.format
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - CALLBACKS

//export mmal_port_callback
func mmal_port_callback(port *C.MMAL_PORT_T, buffer *C.MMAL_BUFFER_HEADER_T) {
	fmt.Printf("mmal_port_callback port=%v buffer=%v\n", port, buffer)
}
