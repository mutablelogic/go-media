package ffmpeg

import (
	"bytes"
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <errno.h>
#include <libavutil/error.h>

// Takes the errno value as a C int, resolved by cgo from <errno.h> at
// compile time (C.EAGAIN etc, see IsEAGAIN/IsEINVAL/IsENOSYS below) - NOT
// a numeric value computed on the Go side. On Windows, Go's syscall.EAGAIN
// etc are synthetic values with no relation to the real errno.h EAGAIN that
// mingw-compiled FFmpeg actually returns (Go deliberately places them in
// the unused APPLICATION_ERROR bit-flag range so they can't collide with
// real Windows error codes), so comparing a C-returned AVERROR(EAGAIN)
// against AVERROR(<Go's EAGAIN value>) silently never matches there. Using
// the same compiler's own errno.h macro on both sides is correct by
// construction regardless of platform.
static int av_error_is(int av, int en) {
	return av == AVERROR(en);
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVError C.int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	errBufferSize = C.AV_ERROR_MAX_STRING_SIZE
)

// Exposed for tests to construct representative AVError values (e.g.
// AVError(-testErrnoEAGAIN)) without a tautological comparison against
// syscall.EAGAIN - and without a cgo import of their own, since a package
// can't split cgo compilation between its regular files and test files.
const (
	testErrnoEAGAIN = C.EAGAIN
	testErrnoEINVAL = C.EINVAL
	testErrnoENOSYS = C.ENOSYS
)

const (
	AVERROR_BSF_NOT_FOUND      = C.AVERROR_BSF_NOT_FOUND      ///< Bitstream filter not found
	AVERROR_BUG                = C.AVERROR_BUG                ///< Internal bug, also see AVERROR_BUG2
	AVERROR_BUFFER_TOO_SMALL   = C.AVERROR_BUFFER_TOO_SMALL   ///< Buffer too small
	AVERROR_DECODER_NOT_FOUND  = C.AVERROR_DECODER_NOT_FOUND  ///< Decoder not found
	AVERROR_DEMUXER_NOT_FOUND  = C.AVERROR_DEMUXER_NOT_FOUND  ///< Demuxer not found
	AVERROR_ENCODER_NOT_FOUND  = C.AVERROR_ENCODER_NOT_FOUND  ///< Encoder not found
	AVERROR_EOF                = C.AVERROR_EOF                ///< End of file
	AVERROR_EXIT               = C.AVERROR_EXIT               ///< Immediate exit was requested; the called function should not be restarted
	AVERROR_EXTERNAL           = C.AVERROR_EXTERNAL           ///< Generic error in an external library
	AVERROR_FILTER_NOT_FOUND   = C.AVERROR_FILTER_NOT_FOUND   ///< Filter not found
	AVERROR_INVALIDDATA        = C.AVERROR_INVALIDDATA        ///< Invalid data found when processing input
	AVERROR_MUXER_NOT_FOUND    = C.AVERROR_MUXER_NOT_FOUND    ///< Muxer not found
	AVERROR_OPTION_NOT_FOUND   = C.AVERROR_OPTION_NOT_FOUND   ///< Option not found
	AVERROR_PATCHWELCOME       = C.AVERROR_PATCHWELCOME       ///< Not yet implemented in FFmpeg, patches welcome
	AVERROR_PROTOCOL_NOT_FOUND = C.AVERROR_PROTOCOL_NOT_FOUND ///< Protocol not found
	AVERROR_STREAM_NOT_FOUND   = C.AVERROR_STREAM_NOT_FOUND   ///< Stream not found
	AVERROR_BUG2               = C.AVERROR_BUG2               // This is semantically identical to AVERROR_BUG, it has been introduced in Libav after our AVERROR_BUG and with a modified value
	AVERROR_UNKNOWN            = C.AVERROR_UNKNOWN            ///< Unknown error, typically from an external library
	AVERROR_EXPERIMENTAL       = C.AVERROR_EXPERIMENTAL       ///< Requested feature is flagged experimental. Set strict_std_compliance if you really want to use it.
	AVERROR_INPUT_CHANGED      = C.AVERROR_INPUT_CHANGED      ///< Input changed between calls. Reconfiguration is required. (can be OR-ed with AVERROR_OUTPUT_CHANGED)
	AVERROR_OUTPUT_CHANGED     = C.AVERROR_OUTPUT_CHANGED     ///< Output changed between calls. Reconfiguration is required. (can be OR-ed with AVERROR_INPUT_CHANGED)
	AVERROR_HTTP_BAD_REQUEST   = C.AVERROR_HTTP_BAD_REQUEST   // HTTP & RTSP errors
	AVERROR_HTTP_UNAUTHORIZED  = C.AVERROR_HTTP_UNAUTHORIZED  // HTTP & RTSP errors
	AVERROR_HTTP_FORBIDDEN     = C.AVERROR_HTTP_FORBIDDEN     // HTTP & RTSP errors
	AVERROR_HTTP_NOT_FOUND     = C.AVERROR_HTTP_NOT_FOUND     // HTTP & RTSP errors
	AVERROR_HTTP_OTHER_4XX     = C.AVERROR_HTTP_OTHER_4XX     // HTTP & RTSP errors
	AVERROR_HTTP_SERVER_ERROR  = C.AVERROR_HTTP_SERVER_ERROR  // HTTP & RTSP errors
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (err AVError) Error() string {
	if err == 0 {
		return ""
	}
	cBuffer := make([]byte, errBufferSize)
	if ret := C.av_strerror(C.int(err), (*C.char)(unsafe.Pointer(&cBuffer[0])), errBufferSize); ret == 0 {
		if n := bytes.IndexByte(cBuffer, 0); n >= 0 {
			return string(cBuffer[:n])
		} else {
			return string(cBuffer)
		}
	} else {
		// av_strerror itself failed (err isn't a code it recognizes) -
		// report the original value, not av_strerror's own status (ret).
		return fmt.Sprintf("Error code: %v", int(err))
	}
}

// IsEAGAIN reports whether err represents EAGAIN/EWOULDBLOCK.
func (err AVError) IsEAGAIN() bool {
	return C.av_error_is(C.int(err), C.EAGAIN) == 1
}

// IsEINVAL reports whether err represents EINVAL.
func (err AVError) IsEINVAL() bool {
	return C.av_error_is(C.int(err), C.EINVAL) == 1
}

// IsENOSYS reports whether err represents ENOSYS.
func (err AVError) IsENOSYS() bool {
	return C.av_error_is(C.int(err), C.ENOSYS) == 1
}
