package ffmpeg

import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
#include <libavutil/samplefmt.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVCodec           C.struct_AVCodec
	AVCodecContext    C.struct_AVCodecContext
	AVCodecParameters C.struct_AVCodecParameters
	AVCodecID         C.enum_AVCodecID
	AVCodecCap        C.int
	AVProfile         C.struct_AVProfile
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_CODEC_CAP_NONE                     AVCodecCap = 0
	AV_CODEC_CAP_DRAW_HORIZ_BAND          AVCodecCap = C.AV_CODEC_CAP_DRAW_HORIZ_BAND     // Decoder can use draw_horiz_band callback
	AV_CODEC_CAP_DR1                      AVCodecCap = C.AV_CODEC_CAP_DR1                 // Codec uses get_buffer() for allocating buffers and supports custom allocators
	AV_CODEC_CAP_DELAY                    AVCodecCap = C.AV_CODEC_CAP_DELAY               // Encoder or decoder requires flushing with NULL input at the end in order to give the complete and correct output
	AV_CODEC_CAP_SMALL_LAST_FRAME         AVCodecCap = C.AV_CODEC_CAP_SMALL_LAST_FRAME    // Codec can be fed a final frame with a smaller size
	AV_CODEC_CAP_SUBFRAMES                AVCodecCap = C.AV_CODEC_CAP_SUBFRAMES           // Codec can output multiple frames per AVPacket Normally demuxers return one frame at a time, demuxers which do not do are connected to a parser to split what they return into proper frames
	AV_CODEC_CAP_EXPERIMENTAL             AVCodecCap = C.AV_CODEC_CAP_EXPERIMENTAL        // Codec is experimental and is thus avoided in favor of non experimental encoders
	AV_CODEC_CAP_CHANNEL_CONF             AVCodecCap = C.AV_CODEC_CAP_CHANNEL_CONF        // Codec should fill in channel configuration and samplerate instead of container
	AV_CODEC_CAP_FRAME_THREADS            AVCodecCap = C.AV_CODEC_CAP_FRAME_THREADS       // Codec supports frame-level multithreading
	AV_CODEC_CAP_SLICE_THREADS            AVCodecCap = C.AV_CODEC_CAP_SLICE_THREADS       // Codec supports slice-based (or partition-based) multithreading
	AV_CODEC_CAP_PARAM_CHANGE             AVCodecCap = C.AV_CODEC_CAP_PARAM_CHANGE        // Codec supports changed parameters at any point
	AV_CODEC_CAP_AUTO_THREADS             AVCodecCap = C.AV_CODEC_CAP_AUTO_THREADS        // Codec supports avctx->thread_count == 0 (auto)
	AV_CODEC_CAP_VARIABLE_FRAME_SIZE      AVCodecCap = C.AV_CODEC_CAP_VARIABLE_FRAME_SIZE // Audio encoder supports receiving a different number of samples in each call
	AV_CODEC_CAP_AVOID_PROBING            AVCodecCap = C.AV_CODEC_CAP_AVOID_PROBING       // Decoder is not a preferred choice for probing
	AV_CODEC_CAP_HARDWARE                 AVCodecCap = C.AV_CODEC_CAP_HARDWARE            // Codec is backed by a hardware implementation
	AV_CODEC_CAP_HYBRID                   AVCodecCap = C.AV_CODEC_CAP_HYBRID              // Codec is potentially backed by a hardware implementation, but not necessarily
	AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE AVCodecCap = C.AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE
	AV_CODEC_CAP_ENCODER_FLUSH            AVCodecCap = C.AV_CODEC_CAP_ENCODER_FLUSH //  This encoder can be flushed using avcodec_flush_buffers()
	//AV_CODEC_CAP_ENCODER_RECON_FRAME      AVCodecCap = C.AV_CODEC_CAP_ENCODER_RECON_FRAME // The encoder is able to output reconstructed frame data
	AV_CODEC_CAP_MAX = AV_CODEC_CAP_ENCODER_FLUSH
)

const (
	FF_PROFILE_UNKNOWN = C.FF_PROFILE_UNKNOWN
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p AVProfile) String() string {
	str := "<AVProfile"
	str += fmt.Sprintf(" profile=%v", p.profile)
	str += fmt.Sprintf(" name=%q", C.GoString(p.name))
	return str + ">"
}

func (v AVCodecCap) String() string {
	if v == AV_CODEC_CAP_NONE {
		return v.FlagString()
	}
	str := ""
	for i := AVCodecCap(C.int(1)); i <= AV_CODEC_CAP_MAX; i <<= 1 {
		if v&i == i {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (v AVCodecCap) FlagString() string {
	switch v {
	case AV_CODEC_CAP_NONE:
		return "AV_CODEC_CAP_NONE"
	case AV_CODEC_CAP_DRAW_HORIZ_BAND:
		return "AV_CODEC_CAP_DRAW_HORIZ_BAND"
	case AV_CODEC_CAP_DR1:
		return "AV_CODEC_CAP_DR1"
	case AV_CODEC_CAP_DELAY:
		return "AV_CODEC_CAP_DELAY"
	case AV_CODEC_CAP_SMALL_LAST_FRAME:
		return "AV_CODEC_CAP_SMALL_LAST_FRAME"
	case AV_CODEC_CAP_SUBFRAMES:
		return "AV_CODEC_CAP_SUBFRAMES"
	case AV_CODEC_CAP_EXPERIMENTAL:
		return "AV_CODEC_CAP_EXPERIMENTAL"
	case AV_CODEC_CAP_CHANNEL_CONF:
		return "AV_CODEC_CAP_CHANNEL_CONF"
	case AV_CODEC_CAP_FRAME_THREADS:
		return "AV_CODEC_CAP_FRAME_THREADS"
	case AV_CODEC_CAP_SLICE_THREADS:
		return "AV_CODEC_CAP_SLICE_THREADS"
	case AV_CODEC_CAP_PARAM_CHANGE:
		return "AV_CODEC_CAP_PARAM_CHANGE"
	case AV_CODEC_CAP_AUTO_THREADS:
		return "AV_CODEC_CAP_AUTO_THREADS"
	case AV_CODEC_CAP_VARIABLE_FRAME_SIZE:
		return "AV_CODEC_CAP_VARIABLE_FRAME_SIZE"
	case AV_CODEC_CAP_AVOID_PROBING:
		return "AV_CODEC_CAP_AVOID_PROBING"
	case AV_CODEC_CAP_HARDWARE:
		return "AV_CODEC_CAP_HARDWARE"
	case AV_CODEC_CAP_HYBRID:
		return "AV_CODEC_CAP_HYBRID"
	case AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE:
		return "AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE"
	case AV_CODEC_CAP_ENCODER_FLUSH:
		return "AV_CODEC_CAP_ENCODER_FLUSH"
	//case AV_CODEC_CAP_ENCODER_RECON_FRAME:
	//	return "AV_CODEC_CAP_ENCODER_RECON_FRAME"
	default:
		return "[?? Invalid AVCodecCap value]"
	}
}

func (c *AVCodec) String() string {
	str := "<AVCodec"
	if c.AVCodec_is_encoder() {
		str += " encoder"
	}
	if c.AVCodec_is_decoder() {
		str += " decoder"
	}
	if id := c.ID(); id != 0 {
		str += fmt.Sprint(" id=", id)
	}
	if name := c.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if media_type := c.MediaType(); media_type != AVMEDIA_TYPE_UNKNOWN {
		str += fmt.Sprint(" media_type=", media_type)
	}
	if pixel_formats := c.PixelFormats(); len(pixel_formats) > 0 {
		str += fmt.Sprint(" pixel_formats=", pixel_formats)
	}
	if sample_rates := c.SampleRates(); len(sample_rates) > 0 {
		str += fmt.Sprint(" supported_samplerates=", sample_rates)
	}
	if sample_fmts := c.SampleFormats(); len(sample_fmts) > 0 {
		str += fmt.Sprint(" sample_fmts=", sample_fmts)
	}
	if profiles := c.Profiles(); len(profiles) > 0 {
		str += fmt.Sprint(" profiles=", profiles)
	}
	if ch_layouts := c.ChannelLayouts(); len(ch_layouts) > 0 {
		str += fmt.Sprint(" ch_layouts=", ch_layouts)
	}
	if cap := c.Cap(); cap != AV_CODEC_CAP_NONE {
		str += fmt.Sprint(" cap=", cap)
	}
	if description := c.Description(); description != "" {
		str += fmt.Sprintf(" description=%q", description)
	}
	if wrapper_name := c.WrapperName(); wrapper_name != "" {
		str += fmt.Sprintf(" wrapper_name=%q", wrapper_name)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (c *AVCodec) Name() string {
	return C.GoString(c.name)
}

func (c *AVCodec) Description() string {
	return C.GoString(c.long_name)
}

func (c *AVCodec) MediaType() AVMediaType {
	return AVMediaType(c._type)
}

func (c *AVCodec) Cap() AVCodecCap {
	return AVCodecCap(c.capabilities)
}

func (c *AVCodec) ID() AVCodecID {
	return AVCodecID(c.id)
}

func (c *AVCodec) PixelFormats() []AVPixelFormat {
	var result []AVPixelFormat
	ptr := uintptr(unsafe.Pointer(c.pix_fmts))
	if ptr == 0 {
		return nil
	}
	for {
		v := AVPixelFormat(*(*C.enum_AVPixelFormat)(unsafe.Pointer(ptr)))
		if v == AV_PIX_FMT_NONE {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AV_PIX_FMT_NONE)
	}
	return result
}

func (c *AVCodec) SampleRates() []int {
	var result []int
	ptr := uintptr(unsafe.Pointer(c.supported_samplerates))
	if ptr == 0 {
		return nil
	}
	for {
		v := int(*(*C.int)(unsafe.Pointer(ptr)))
		if v == 0 {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(C.int(0))
	}
	return result
}

func (c *AVCodec) SampleFormats() []AVSampleFormat {
	var result []AVSampleFormat
	ptr := uintptr(unsafe.Pointer(c.sample_fmts))
	if ptr == 0 {
		return nil
	}
	for {
		v := AVSampleFormat(*(*C.enum_AVSampleFormat)(unsafe.Pointer(ptr)))
		if v == AV_SAMPLE_FMT_NONE {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AV_SAMPLE_FMT_NONE)
	}
	return result
}

func (c *AVCodec) Profiles() []AVProfile {
	var result []AVProfile
	ptr := uintptr(unsafe.Pointer(c.profiles))
	if ptr == 0 {
		return nil
	}
	for {
		v := (AVProfile)(*(*C.struct_AVProfile)(unsafe.Pointer(ptr)))
		if v.profile == FF_PROFILE_UNKNOWN {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AVProfile{})
	}
	return result
}

func (c *AVCodec) ChannelLayouts() []AVChannelLayout {
	var result []AVChannelLayout
	ptr := uintptr(unsafe.Pointer(c.ch_layouts))
	if ptr == 0 {
		return nil
	}
	for {
		v := (AVChannelLayout)(*(*C.struct_AVChannelLayout)(unsafe.Pointer(ptr)))
		if v.nb_channels == 0 {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AVChannelLayout{})
	}
	return result
}

func (c *AVCodec) WrapperName() string {
	return C.GoString(c.wrapper_name)
}
