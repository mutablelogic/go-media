/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	For Licensing and Usage information, please see LICENSE.md
*/

package mmal

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MMAL_SUCCESS   status = iota
	MMAL_ENOMEM           // Out of memory
	MMAL_ENOSPC           // Out of resources (other than memory)
	MMAL_EINVAL           // Argument is invalid
	MMAL_ENOSYS           // Function not implemented
	MMAL_ENOENT           // No such file or directory
	MMAL_ENXIO            // No such device or address
	MMAL_EIO              // I/O error
	MMAL_ESPIPE           // Illegal seek
	MMAL_ECORRUPT         // Data is corrupt
	MMAL_ENOTREADY        // Component is not ready
	MMAL_ECONFIG          // Component is not configured
	MMAL_EISCONN          // Port is already connected
	MMAL_ENOTCONN         // Port is disconnected
	MMAL_EAGAIN           // Resource temporarily unavailable. Try again later
	MMAL_EFAULT           // Bad address
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (s status) Error() string {
	switch s {
	case MMAL_SUCCESS:
		return "MMAL_SUCCESS"
	case MMAL_ENOMEM:
		return "MMAL_ENOMEM"
	case MMAL_ENOSPC:
		return "MMAL_ENOSPC"
	case MMAL_EINVAL:
		return "MMAL_EINVAL"
	case MMAL_ENOSYS:
		return "MMAL_ENOSYS"
	case MMAL_ENOENT:
		return "MMAL_ENOENT"
	case MMAL_ENXIO:
		return "MMAL_ENXIO"
	case MMAL_EIO:
		return "MMAL_EIO"
	case MMAL_ESPIPE:
		return "MMAL_ESPIPE"
	case MMAL_ECORRUPT:
		return "MMAL_ECORRUPT"
	case MMAL_ENOTREADY:
		return "MMAL_ENOTREADY"
	case MMAL_ECONFIG:
		return "MMAL_ECONFIG"
	case MMAL_EISCONN:
		return "MMAL_EISCONN"
	case MMAL_ENOTCONN:
		return "MMAL_ENOTCONN"
	case MMAL_EAGAIN:
		return "MMAL_EAGAIN"
	case MMAL_EFAULT:
		return "MMAL_EFAULT"
	default:
		return "[?? Invalid status value]"
	}
}
