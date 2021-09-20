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
	#cgo LDFLAGS: -L/opt/vc/lib -lmmal -lmmal_core
	#include <interface/mmal/mmal.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	status          int
	componentHandle (*C.MMAL_COMPONENT_T)
	portHandle      (*C.MMAL_PORT_T)
	portType        (C.MMAL_PORT_TYPE_T)
	portCallback    (C.MMAL_PORT_BH_CB_T)
	esFormatHandle  (*C.MMAL_ES_FORMAT_T)
	paramHandle     (*C.MMAL_PARAMETER_HEADER_T)
)
