package ffmpeg

import (
	"errors"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/samplefmt.h>
*/
import "C"

const (
	AV_NUM_PLANES = 8
)

type AVSamples struct {
	nb_samples  C.int
	nb_channels C.int
	sample_fmt  C.enum_AVSampleFormat
	plane_size  C.int
	buffer_size C.int
	planes      [AV_NUM_PLANES]*C.uint8_t
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (data *AVSamples) Bytes(plane int) []byte {
	if plane < 0 || plane >= AV_NUM_PLANES {
		return nil
	}
	if ptr := data.planes[plane]; ptr == nil {
		return nil
	} else {
		return cByteSlice(unsafe.Pointer(ptr), data.plane_size)
	}
}

func (data *AVSamples) Int16(plane int) []int16 {
	if plane < 0 || plane >= AV_NUM_PLANES {
		return nil
	}
	if ptr := data.planes[plane]; ptr == nil {
		return nil
	} else {
		return cInt16Slice(unsafe.Pointer(ptr), data.plane_size>>1)
	}
}

func (data *AVSamples) Float64(plane int) []float64 {
	if plane < 0 || plane >= AV_NUM_PLANES {
		return nil
	}
	if ptr := data.planes[plane]; ptr == nil {
		return nil
	} else {
		return cFloat64Slice(unsafe.Pointer(ptr), data.plane_size>>3)
	}
}

func (data *AVSamples) NumPlanes() int {
	if AVUtil_sample_fmt_is_planar(AVSampleFormat(data.sample_fmt)) {
		return int(data.nb_channels)
	} else {
		return 1
	}
}

func (data *AVSamples) NumChannels() int {
	return int(data.nb_channels)
}

func (data *AVSamples) NumSamples() int {
	return int(data.nb_samples)
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Get AVSamples from an AVFrame
func AVUtil_samples_frame(frame *AVFrame) *AVSamples {
	return &AVSamples{
		nb_samples:  frame.nb_samples,
		nb_channels: frame.channels,
		sample_fmt:  C.enum_AVSampleFormat(frame.format),
		plane_size:  frame.linesize[0],
		buffer_size: frame.linesize[0] * frame.nb_samples,
		planes:      frame.data,
	}
}

// Allocate a samples buffer for nb_samples samples. Return allocated data for each plane, and the stride.
func AVUtil_samples_alloc(nb_samples, nb_channels int, sample_fmt AVSampleFormat, align bool) (*AVSamples, error) {
	if nb_channels < 1 {
		return nil, errors.New("too few channels")
	}
	if AVUtil_sample_fmt_is_planar(sample_fmt) && nb_channels > AV_NUM_PLANES {
		return nil, errors.New("too many channels")
	}
	data := &AVSamples{nb_samples: C.int(nb_samples), nb_channels: C.int(nb_channels), sample_fmt: C.enum_AVSampleFormat(sample_fmt)}

	// Get the size
	if size := C.av_samples_get_buffer_size(nil, data.nb_channels, data.nb_samples, data.sample_fmt, boolToInt(align)); size < 0 {
		return nil, AVError(size)
	} else {
		data.buffer_size = size
	}

	// Allocate the buffer
	if err := AVError(C.av_samples_alloc(&data.planes[0], &data.plane_size, data.nb_channels, data.nb_samples, data.sample_fmt, boolToInt(align))); err < 0 {
		return nil, err
	}

	// Return success
	return data, nil
}

// Free the samples
func AVUtil_samples_free(samples *AVSamples) {
	C.av_freep(unsafe.Pointer(&samples.planes[0]))
}

// Get the required buffer size for the given audio parameters. Returns the calculated buffer size and plane size.
func AVUtil_samples_get_buffer_size(nb_samples, nb_channels int, sample_fmt AVSampleFormat, align bool) (int, int, error) {
	var plane_size C.int
	ret := int(C.av_samples_get_buffer_size(&plane_size, C.int(nb_channels), C.int(nb_samples), C.enum_AVSampleFormat(sample_fmt), boolToInt(align)))
	if ret < 0 {
		return 0, 0, AVError(ret)
	} else {
		return ret, int(plane_size), nil
	}
}

// Copy samples - dst and src channels and formats need to match
func AVUtil_samples_copy(dst, src *AVSamples, dst_offset, src_offset, nb_samples int) error {
	if dst.sample_fmt != src.sample_fmt {
		return errors.New("sample formats do not match")
	}
	if dst.nb_channels != src.nb_channels {
		return errors.New("sample channels do not match")
	}

	// Perform the copy
	C.av_samples_copy(&dst.planes[0], &src.planes[0], C.int(dst_offset), C.int(src_offset), C.int(nb_samples), dst.nb_channels, dst.sample_fmt)

	// Return success
	return nil
}

// Fill an audio buffer with silence
func AVUtil_samples_set_silence(data *AVSamples, offset int, nb_samples int) {
	C.av_samples_set_silence(&data.planes[0], C.int(offset), C.int(nb_samples), data.nb_channels, data.sample_fmt)
}
