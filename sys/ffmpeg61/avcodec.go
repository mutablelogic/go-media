package ffmpeg

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec libavutil
#include <libavcodec/avcodec.h>
#include <libavutil/opt.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVPacket                      C.AVPacket
	AVCodec                       C.AVCodec
	AVCodecCap                    C.uint32_t
	AVCodecContext                C.AVCodecContext
	AVCodecFlag                   C.uint32_t
	AVCodecFlag2                  C.uint32_t
	AVCodecMacroblockDecisionMode C.int
	AVCodecParameters             C.AVCodecParameters
	AVCodecParser                 C.AVCodecParser
	AVCodecParserContext          C.AVCodecParserContext
	AVProfile                     C.AVProfile
	AVCodecID                     C.enum_AVCodecID
)

type jsonAVCodec struct {
	Type           AVMediaType       `json:"type"`
	Name           string            `json:"name,omitempty"`
	LongName       string            `json:"long_name,omitempty"`
	ID             AVCodecID         `json:"id,omitempty"`
	Capabilities   AVCodecCap        `json:"capabilities,omitempty"`
	Framerates     []AVRational      `json:"supported_framerates,omitempty"`
	SampleFormats  []AVSampleFormat  `json:"sample_formats,omitempty"`
	PixelFormats   []AVPixelFormat   `json:"pixel_formats,omitempty"`
	Samplerates    []int             `json:"samplerates,omitempty"`
	Profiles       []AVProfile       `json:"profiles,omitempty"`
	ChannelLayouts []AVChannelLayout `json:"channel_layouts,omitempty"`
}

type jsonAVCodecContext struct {
	CodecType         AVMediaType     `json:"codec_type,omitempty"`
	Codec             *AVCodec        `json:"codec,omitempty"`
	BitRate           int64           `json:"bit_rate,omitempty"`
	BitRateTolerance  int             `json:"bit_rate_tolerance,omitempty"`
	PixelFormat       AVPixelFormat   `json:"pix_fmt,omitempty"`
	Width             int             `json:"width,omitempty"`
	Height            int             `json:"height,omitempty"`
	SampleAspectRatio AVRational      `json:"sample_aspect_ratio,omitempty"`
	Framerate         AVRational      `json:"framerate,omitempty"`
	SampleFormat      AVSampleFormat  `json:"sample_fmt,omitempty"`
	SampleRate        int             `json:"sample_rate,omitempty"`
	ChannelLayout     AVChannelLayout `json:"channel_layout,omitempty"`
	FrameSize         int             `json:"frame_size,omitempty"`
	TimeBase          AVRational      `json:"time_base,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_CODEC_ID_NONE       AVCodecID = C.AV_CODEC_ID_NONE
	AV_CODEC_ID_MP2        AVCodecID = C.AV_CODEC_ID_MP2
	AV_CODEC_ID_H264       AVCodecID = C.AV_CODEC_ID_H264
	AV_CODEC_ID_MPEG1VIDEO AVCodecID = C.AV_CODEC_ID_MPEG1VIDEO
	AV_CODEC_ID_MPEG2VIDEO AVCodecID = C.AV_CODEC_ID_MPEG2VIDEO
)

/**
 * Required number of additionally allocated bytes at the end of the input bitstream for decoding.
 * This is mainly needed because some optimized bitstream readers read
 * 32 or 64 bit at once and could read over the end.
 * Note: If the first 23 bits of the additional bytes are not 0, then damaged
 * MPEG bitstreams could cause overread and segfault.
 */
const (
	AV_INPUT_BUFFER_PADDING_SIZE int = C.AV_INPUT_BUFFER_PADDING_SIZE
)

/**
 * macroblock decision mode
 * - encoding: Set by user.
 * - decoding: unused
 */
const (
	FF_MB_DECISION_SIMPLE AVCodecMacroblockDecisionMode = C.FF_MB_DECISION_SIMPLE ///< uses mb_cmp
	FF_MB_DECISION_BITS   AVCodecMacroblockDecisionMode = C.FF_MB_DECISION_BITS   ///< chooses the one which needs the fewest bits
	FF_MB_DECISION_RD     AVCodecMacroblockDecisionMode = C.FF_MB_DECISION_RD     ///< rate distortion
)

const (
	AV_CODEC_FLAG_UNALIGNED      AVCodecFlag  = C.AV_CODEC_FLAG_UNALIGNED      // Allow decoders to produce frames with data planes that are not aligned to CPU requirements
	AV_CODEC_FLAG_QSCALE         AVCodecFlag  = C.AV_CODEC_FLAG_QSCALE         // Use fixed qscale
	AV_CODEC_FLAG_4MV            AVCodecFlag  = C.AV_CODEC_FLAG_4MV            // 4 MV per MB allowed / advanced prediction for H.263.
	AV_CODEC_FLAG_OUTPUT_CORRUPT AVCodecFlag  = C.AV_CODEC_FLAG_OUTPUT_CORRUPT // Output even those frames that might be corrupted.
	AV_CODEC_FLAG_QPEL           AVCodecFlag  = C.AV_CODEC_FLAG_QPEL           // Use qpel MC.
	AV_CODEC_FLAG_RECON_FRAME    AVCodecFlag  = C.AV_CODEC_FLAG_RECON_FRAME    // Request the encoder to output reconstructed frames
	AV_CODEC_FLAG_COPY_OPAQUE    AVCodecFlag  = C.AV_CODEC_FLAG_COPY_OPAQUE    // Request the decoder to propagate each packet's AVPacket.opaque and AVPacket.opaque_ref to its corresponding output AVFrame.
	AV_CODEC_FLAG_FRAME_DURATION AVCodecFlag  = C.AV_CODEC_FLAG_FRAME_DURATION // Signal to the encoder that the values of AVFrame.duration are valid and should be used
	AV_CODEC_FLAG_PASS1          AVCodecFlag  = C.AV_CODEC_FLAG_PASS1          // Use internal 2pass ratecontrol in first pass mode.
	AV_CODEC_FLAG_PASS2          AVCodecFlag  = C.AV_CODEC_FLAG_PASS2          // Use internal 2pass ratecontrol in second pass mode.
	AV_CODEC_FLAG_LOOP_FILTER    AVCodecFlag  = C.AV_CODEC_FLAG_LOOP_FILTER    // loop filter.
	AV_CODEC_FLAG_GRAY           AVCodecFlag  = C.AV_CODEC_FLAG_GRAY           // Only decode/encode grayscale.
	AV_CODEC_FLAG_PSNR           AVCodecFlag  = C.AV_CODEC_FLAG_PSNR           // error[?] variables will be set during encoding.
	AV_CODEC_FLAG_INTERLACED_DCT AVCodecFlag  = C.AV_CODEC_FLAG_INTERLACED_DCT // Use interlaced DCT.
	AV_CODEC_FLAG_LOW_DELAY      AVCodecFlag  = C.AV_CODEC_FLAG_LOW_DELAY      // Force low delay.
	AV_CODEC_FLAG_GLOBAL_HEADER  AVCodecFlag  = C.AV_CODEC_FLAG_GLOBAL_HEADER  // Place global headers in extradata instead of every keyframe.
	AV_CODEC_FLAG_BITEXACT       AVCodecFlag  = C.AV_CODEC_FLAG_BITEXACT       // Use only bitexact stuff (except (I)DCT).
	AV_CODEC_FLAG_AC_PRED        AVCodecFlag  = C.AV_CODEC_FLAG_AC_PRED        // H.263 advanced intra coding / MPEG-4 AC prediction
	AV_CODEC_FLAG_INTERLACED_ME  AVCodecFlag  = C.AV_CODEC_FLAG_INTERLACED_ME  // interlaced motion estimation
	AV_CODEC_FLAG_CLOSED_GOP     AVCodecFlag  = C.AV_CODEC_FLAG_CLOSED_GOP
	AV_CODEC_FLAG2_FAST          AVCodecFlag2 = C.AV_CODEC_FLAG2_FAST          // Allow non spec compliant speedup tricks.
	AV_CODEC_FLAG2_NO_OUTPUT     AVCodecFlag2 = C.AV_CODEC_FLAG2_NO_OUTPUT     // Skip bitstream encoding.
	AV_CODEC_FLAG2_LOCAL_HEADER  AVCodecFlag2 = C.AV_CODEC_FLAG2_LOCAL_HEADER  // Place global headers at every keyframe instead of in extradata.
	AV_CODEC_FLAG2_CHUNKS        AVCodecFlag2 = C.AV_CODEC_FLAG2_CHUNKS        // Input bitstream might be truncated at a packet boundaries instead of only at frame boundaries.
	AV_CODEC_FLAG2_IGNORE_CROP   AVCodecFlag2 = C.AV_CODEC_FLAG2_IGNORE_CROP   // Discard cropping information from SPS.
	AV_CODEC_FLAG2_SHOW_ALL      AVCodecFlag2 = C.AV_CODEC_FLAG2_SHOW_ALL      // Show all frames before the first keyframe
	AV_CODEC_FLAG2_EXPORT_MVS    AVCodecFlag2 = C.AV_CODEC_FLAG2_EXPORT_MVS    // Export motion vectors through frame side data
	AV_CODEC_FLAG2_SKIP_MANUAL   AVCodecFlag2 = C.AV_CODEC_FLAG2_SKIP_MANUAL   // Do not skip samples and export skip information as frame side data
	AV_CODEC_FLAG2_RO_FLUSH_NOOP AVCodecFlag2 = C.AV_CODEC_FLAG2_RO_FLUSH_NOOP // Do not reset ASS ReadOrder field on flush (subtitles decoding)
	AV_CODEC_FLAG2_ICC_PROFILES  AVCodecFlag2 = C.AV_CODEC_FLAG2_ICC_PROFILES  // Generate/parse ICC profiles on encode/decode, as appropriate for the type of file
)

const (
	AV_CODEC_CAP_NONE                     AVCodecCap = 0
	AV_CODEC_CAP_DRAW_HORIZ_BAND          AVCodecCap = C.AV_CODEC_CAP_DRAW_HORIZ_BAND          // Decoder can use draw_horiz_band callback
	AV_CODEC_CAP_DR1                      AVCodecCap = C.AV_CODEC_CAP_DR1                      // Codec uses get_buffer() for allocating buffers and supports custom allocators
	AV_CODEC_CAP_DELAY                    AVCodecCap = C.AV_CODEC_CAP_DELAY                    // Encoder or decoder requires flushing with NULL input at the end in order to give the complete and correct output
	AV_CODEC_CAP_SMALL_LAST_FRAME         AVCodecCap = C.AV_CODEC_CAP_SMALL_LAST_FRAME         // Codec can be fed a final frame with a smaller size
	AV_CODEC_CAP_SUBFRAMES                AVCodecCap = C.AV_CODEC_CAP_SUBFRAMES                // Codec can output multiple frames per AVPacket Normally demuxers return one frame at a time, demuxers which do not do are connected to a parser to split what they return into proper frames
	AV_CODEC_CAP_EXPERIMENTAL             AVCodecCap = C.AV_CODEC_CAP_EXPERIMENTAL             // Codec is experimental and is thus avoided in favor of non experimental encoders
	AV_CODEC_CAP_CHANNEL_CONF             AVCodecCap = C.AV_CODEC_CAP_CHANNEL_CONF             // Codec should fill in channel configuration and samplerate instead of container
	AV_CODEC_CAP_FRAME_THREADS            AVCodecCap = C.AV_CODEC_CAP_FRAME_THREADS            // Codec supports frame-level multithreading
	AV_CODEC_CAP_SLICE_THREADS            AVCodecCap = C.AV_CODEC_CAP_SLICE_THREADS            // Codec supports slice-based (or partition-based) multithreading
	AV_CODEC_CAP_PARAM_CHANGE             AVCodecCap = C.AV_CODEC_CAP_PARAM_CHANGE             // Codec supports changed parameters at any point
	AV_CODEC_CAP_OTHER_THREADS            AVCodecCap = C.AV_CODEC_CAP_OTHER_THREADS            // Codec supports multithreading through a method other than slice
	AV_CODEC_CAP_VARIABLE_FRAME_SIZE      AVCodecCap = C.AV_CODEC_CAP_VARIABLE_FRAME_SIZE      // Audio encoder supports receiving a different number of samples in each call
	AV_CODEC_CAP_AVOID_PROBING            AVCodecCap = C.AV_CODEC_CAP_AVOID_PROBING            // Decoder is not a preferred choice for probing
	AV_CODEC_CAP_HARDWARE                 AVCodecCap = C.AV_CODEC_CAP_HARDWARE                 // Codec is backed by a hardware implementation
	AV_CODEC_CAP_HYBRID                   AVCodecCap = C.AV_CODEC_CAP_HYBRID                   // Codec is potentially backed by a hardware implementation, but not necessarily
	AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE AVCodecCap = C.AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE // This encoder can reorder user opaque values from input AVFrames and return them with corresponding output packets.
	AV_CODEC_CAP_ENCODER_FLUSH            AVCodecCap = C.AV_CODEC_CAP_ENCODER_FLUSH            //  This encoder can be flushed using avcodec_flush_buffers()
	AV_CODEC_CAP_ENCODER_RECON_FRAME      AVCodecCap = C.AV_CODEC_CAP_ENCODER_RECON_FRAME      // The encoder is able to output reconstructed frame data
	AV_CODEC_CAP_MAX                                 = AV_CODEC_CAP_ENCODER_RECON_FRAME
)

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx *AVCodec) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVCodec{
		Name:           C.GoString(ctx.name),
		LongName:       C.GoString(ctx.long_name),
		Type:           AVMediaType(ctx._type),
		ID:             AVCodecID(ctx.id),
		Capabilities:   AVCodecCap(ctx.capabilities),
		Framerates:     ctx.SupportedFramerates(),
		SampleFormats:  ctx.SampleFormats(),
		PixelFormats:   ctx.PixelFormats(),
		Samplerates:    ctx.SupportedSamplerates(),
		Profiles:       ctx.Profiles(),
		ChannelLayouts: ctx.ChannelLayouts(),
	})
}

func (ctx *AVCodecContext) MarshalJSON() ([]byte, error) {
	switch ctx.codec_type {
	case C.AVMEDIA_TYPE_VIDEO:
		return json.Marshal(jsonAVCodecContext{
			CodecType:         AVMediaType(ctx.codec_type),
			Codec:             (*AVCodec)(ctx.codec),
			BitRate:           int64(ctx.bit_rate),
			BitRateTolerance:  int(ctx.bit_rate_tolerance),
			PixelFormat:       AVPixelFormat(ctx.pix_fmt),
			Width:             int(ctx.width),
			Height:            int(ctx.height),
			SampleAspectRatio: AVRational(ctx.sample_aspect_ratio),
			Framerate:         AVRational(ctx.framerate),
		})
	case C.AVMEDIA_TYPE_AUDIO:
		return json.Marshal(jsonAVCodecContext{
			CodecType:        AVMediaType(ctx.codec_type),
			Codec:            (*AVCodec)(ctx.codec),
			BitRate:          int64(ctx.bit_rate),
			BitRateTolerance: int(ctx.bit_rate_tolerance),
			TimeBase:         (AVRational)(ctx.time_base),
			SampleFormat:     AVSampleFormat(ctx.sample_fmt),
			SampleRate:       int(ctx.sample_rate),
			ChannelLayout:    AVChannelLayout(ctx.ch_layout),
			FrameSize:        int(ctx.frame_size),
		})
	default:
		return json.Marshal(jsonAVCodecContext{
			CodecType:        AVMediaType(ctx.codec_type),
			Codec:            (*AVCodec)(ctx.codec),
			BitRate:          int64(ctx.bit_rate),
			BitRateTolerance: int(ctx.bit_rate_tolerance),
			TimeBase:         (AVRational)(ctx.time_base),
		})
	}
}

func (ctx AVProfile) MarshalJSON() ([]byte, error) {
	return json.Marshal(ctx.Name())
}

func (ctx AVMediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ctx.String())
}

func (v AVCodecCap) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v AVCodecID) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVCodec) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx *AVCodecContext) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx AVProfile) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVCodec

func (c *AVCodec) Name() string {
	return C.GoString(c.name)
}

func (c *AVCodec) LongName() string {
	return C.GoString(c.long_name)
}

func (c *AVCodec) Type() AVMediaType {
	return AVMediaType(c._type)
}

func (c *AVCodec) ID() AVCodecID {
	return AVCodecID(c.id)
}

func (c *AVCodec) Capabilities() AVCodecCap {
	return AVCodecCap(c.capabilities)
}

func (c *AVCodec) SupportedFramerates() []AVRational {
	var result []AVRational
	ptr := uintptr(unsafe.Pointer(c.supported_framerates))
	if ptr == 0 {
		return nil
	}
	for {
		v := AVRational(*(*C.struct_AVRational)(unsafe.Pointer(ptr)))
		if v.IsZero() {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AVRational{})
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

func (c *AVCodec) SupportedSamplerates() []int {
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

func (c *AVCodec) Profiles() []AVProfile {
	var result []AVProfile
	ptr := uintptr(unsafe.Pointer(c.profiles))
	if ptr == 0 {
		return nil
	}
	for {
		v := (AVProfile)(*(*C.struct_AVProfile)(unsafe.Pointer(ptr)))
		if v.profile == C.FF_PROFILE_UNKNOWN {
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

////////////////////////////////////////////////////////////////////////////////
// AVCodecContext

func (ctx *AVCodecContext) Codec() *AVCodec {
	return (*AVCodec)(ctx.codec)
}

func (ctx *AVCodecContext) BitRate() int64 {
	return int64(ctx.bit_rate)
}

func (ctx *AVCodecContext) SetBitRate(bit_rate int64) {
	ctx.bit_rate = C.int64_t(bit_rate)
}

func (ctx *AVCodecContext) Width() int {
	return int(ctx.width)
}

func (ctx *AVCodecContext) SetWidth(width int) {
	ctx.width = C.int(width)
}

func (ctx *AVCodecContext) Height() int {
	return int(ctx.height)
}

func (ctx *AVCodecContext) SetHeight(height int) {
	ctx.height = C.int(height)
}

func (ctx *AVCodecContext) SampleAspectRatio() AVRational {
	return (AVRational)(ctx.sample_aspect_ratio)
}

func (ctx *AVCodecContext) SetSampleAspectRatio(sample_aspect_ratio AVRational) {
	ctx.sample_aspect_ratio = C.struct_AVRational(sample_aspect_ratio)
}

func (ctx *AVCodecContext) Framerate() AVRational {
	return (AVRational)(ctx.framerate)
}

func (ctx *AVCodecContext) SetFramerate(framerate AVRational) {
	ctx.framerate = C.struct_AVRational(framerate)
}

func (ctx *AVCodecContext) TimeBase() AVRational {
	return (AVRational)(ctx.time_base)
}

func (ctx *AVCodecContext) SetTimeBase(time_base AVRational) {
	ctx.time_base = C.struct_AVRational(time_base)
}

// Audio sample format.
func (ctx *AVCodecContext) SampleFormat() AVSampleFormat {
	return AVSampleFormat(ctx.sample_fmt)
}

// Audio sample format.
func (ctx *AVCodecContext) SetSampleFormat(sample_fmt AVSampleFormat) {
	ctx.sample_fmt = C.enum_AVSampleFormat(sample_fmt)
}

// Frame number.
func (ctx *AVCodecContext) FrameNum() int {
	return int(ctx.frame_num)
}

// Audio sample rate.
func (ctx *AVCodecContext) SampleRate() int {
	return int(ctx.sample_rate)
}

// Audio sample rate.
func (ctx *AVCodecContext) SetSampleRate(sample_rate int) {
	ctx.sample_rate = C.int(sample_rate)
}

// Number of samples per channel in an audio frame.
func (ctx *AVCodecContext) FrameSize() int {
	return int(ctx.frame_size)
}

// Audio channel layout.
func (ctx *AVCodecContext) ChannelLayout() AVChannelLayout {
	return AVChannelLayout(ctx.ch_layout)
}

// Audio channel layout.
func (ctx *AVCodecContext) SetChannelLayout(src AVChannelLayout) error {
	if ret := AVError(C.av_channel_layout_copy((*C.struct_AVChannelLayout)(&ctx.ch_layout), (*C.struct_AVChannelLayout)(&src))); ret != 0 {
		return ret
	}
	return nil
}

// Group-of-pictures (GOP) size.
func (ctx *AVCodecContext) GopSize() int {
	return int(ctx.gop_size)
}

// Group-of-pictures (GOP) size.
func (ctx *AVCodecContext) SetGopSize(gop_size int) {
	ctx.gop_size = C.int(gop_size)
}

// Maximum number of B-frames between non-B-frames.
func (ctx *AVCodecContext) MaxBFrames() int {
	return int(ctx.max_b_frames)
}

// Maximum number of B-frames between non-B-frames.
func (ctx *AVCodecContext) SetMaxBFrames(max_b_frames int) {
	ctx.max_b_frames = C.int(max_b_frames)
}

// Pixel format.
func (ctx *AVCodecContext) PixFmt() AVPixelFormat {
	return AVPixelFormat(ctx.pix_fmt)
}

// Pixel format.
func (ctx *AVCodecContext) SetPixFmt(pix_fmt AVPixelFormat) {
	ctx.pix_fmt = C.enum_AVPixelFormat(pix_fmt)
}

// Private data, set key/value pair
func (ctx *AVCodecContext) SetPrivDataKV(name, value string) error {
	cName, cValue := C.CString(name), C.CString(value)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cValue))
	if ret := AVError(C.av_opt_set(ctx.priv_data, cName, cValue, 0)); ret != 0 {
		return ret
	}
	return nil
}

// Set Macroblock decision mode.
func (ctx *AVCodecContext) SetMbDecision(mode AVCodecMacroblockDecisionMode) {
	ctx.mb_decision = C.int(mode)
}

// Get Macroblock decision mode.
func (ctx *AVCodecContext) MbDecision() AVCodecMacroblockDecisionMode {
	return AVCodecMacroblockDecisionMode(ctx.mb_decision)
}

// Get flags
func (ctx *AVCodecContext) Flags() AVCodecFlag {
	return AVCodecFlag(ctx.flags)

}

// Set flags
func (ctx *AVCodecContext) SetFlags(flags AVCodecFlag) {
	ctx.flags = C.int(flags)
}

// Get flags2
func (ctx *AVCodecContext) Flags2() AVCodecFlag2 {
	return AVCodecFlag2(ctx.flags2)
}

// Set flags2
func (ctx *AVCodecContext) SetFlags2(flags2 AVCodecFlag2) {
	ctx.flags2 = C.int(flags2)
}

////////////////////////////////////////////////////////////////////////////////
// AVProfile

func (c *AVProfile) ID() int {
	return int(c.profile)
}

func (c *AVProfile) Name() string {
	return C.GoString(c.name)
}

////////////////////////////////////////////////////////////////////////////////
// AVCodecCap

func (v AVCodecCap) Is(cap AVCodecCap) bool {
	return v&cap == cap
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
	case AV_CODEC_CAP_OTHER_THREADS:
		return "AV_CODEC_CAP_OTHER_THREADS"
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
	case AV_CODEC_CAP_ENCODER_RECON_FRAME:
		return "AV_CODEC_CAP_ENCODER_RECON_FRAME"
	default:
		return fmt.Sprintf("AVCodecCap(0x%08X)", uint32(v))
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVCodecID

func (v AVCodecID) String() string {
	return v.Name()
}

func (v AVCodecID) Name() string {
	return C.GoString(C.avcodec_get_name(C.enum_AVCodecID(v)))
}

func (v AVCodecID) Type() AVMediaType {
	return AVMediaType(C.avcodec_get_type(C.enum_AVCodecID(v)))
}
