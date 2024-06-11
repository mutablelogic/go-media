package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// From fill the parameters based on the values from the supplied codec parameters
func AVCodec_parameters_copy(ctx *AVCodecParameters, codecpar *AVCodecParameters) error {
	if err := AVError(C.avcodec_parameters_copy((*C.AVCodecParameters)(ctx), (*C.AVCodecParameters)(codecpar))); err != 0 {
		return err
	} else {
		return nil
	}
}
