package ffmpeg

import (
	"bytes"
	"fmt"
	"syscall"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/error.h>

static int av_error_matches(int av,int en) {
	return av == AVERROR(en);
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	errBufferSize = C.AV_ERROR_MAX_STRING_SIZE
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
	cBuffer := make([]byte, errBufferSize)
	if err := C.av_strerror(C.int(err), (*C.char)(unsafe.Pointer(&cBuffer[0])), errBufferSize); err == 0 {
		if n := bytes.IndexByte(cBuffer, 0); n >= 0 {
			return string(cBuffer[:n])
		} else {
			return string(cBuffer)
		}
	} else {
		return fmt.Sprintf("Error code: %v", int(err))
	}
}

func (err AVError) IsErrno(v syscall.Errno) bool {
	c := int(C.av_error_matches(C.int(err), C.int(v)))
	return c == 1
}
