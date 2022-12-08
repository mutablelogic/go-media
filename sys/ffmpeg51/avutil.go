package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/dict.h>
#include <libavutil/rational.h>
#include <libavutil/samplefmt.h>
#include <libavutil/channel_layout.h>
#include <libavutil/pixfmt.h>

AVChannelLayout _AV_CHANNEL_LAYOUT_MONO = AV_CHANNEL_LAYOUT_MONO;
AVChannelLayout _AV_CHANNEL_LAYOUT_STEREO = AV_CHANNEL_LAYOUT_STEREO;
AVChannelLayout _AV_CHANNEL_LAYOUT_2POINT1 = AV_CHANNEL_LAYOUT_2POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_2_1 = AV_CHANNEL_LAYOUT_2_1;
AVChannelLayout _AV_CHANNEL_LAYOUT_SURROUND = AV_CHANNEL_LAYOUT_SURROUND;
AVChannelLayout _AV_CHANNEL_LAYOUT_3POINT1 = AV_CHANNEL_LAYOUT_3POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_4POINT0 = AV_CHANNEL_LAYOUT_4POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_4POINT1 = AV_CHANNEL_LAYOUT_4POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_2_2 = AV_CHANNEL_LAYOUT_2_2;
AVChannelLayout _AV_CHANNEL_LAYOUT_QUAD = AV_CHANNEL_LAYOUT_QUAD;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT0 = AV_CHANNEL_LAYOUT_5POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT1 = AV_CHANNEL_LAYOUT_5POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT0_BACK = AV_CHANNEL_LAYOUT_5POINT0_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT1_BACK = AV_CHANNEL_LAYOUT_5POINT1_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT0 = AV_CHANNEL_LAYOUT_6POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT0_FRONT = AV_CHANNEL_LAYOUT_6POINT0_FRONT;
AVChannelLayout _AV_CHANNEL_LAYOUT_HEXAGONAL = AV_CHANNEL_LAYOUT_HEXAGONAL;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT1 = AV_CHANNEL_LAYOUT_6POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT1_BACK = AV_CHANNEL_LAYOUT_6POINT1_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT1_FRONT = AV_CHANNEL_LAYOUT_6POINT1_FRONT;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT0 = AV_CHANNEL_LAYOUT_7POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT0_FRONT = AV_CHANNEL_LAYOUT_7POINT0_FRONT;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT1 = AV_CHANNEL_LAYOUT_7POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT1_WIDE = AV_CHANNEL_LAYOUT_7POINT1_WIDE;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK = AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_OCTAGONAL = AV_CHANNEL_LAYOUT_OCTAGONAL;
AVChannelLayout _AV_CHANNEL_LAYOUT_HEXADECAGONAL = AV_CHANNEL_LAYOUT_HEXADECAGONAL;
AVChannelLayout _AV_CHANNEL_LAYOUT_STEREO_DOWNMIX = AV_CHANNEL_LAYOUT_STEREO_DOWNMIX;
AVChannelLayout _AV_CHANNEL_LAYOUT_22POINT2 = AV_CHANNEL_LAYOUT_22POINT2;
AVChannelLayout _AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER = AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER;
*/
import "C"

type (
	AVError           C.int
	AVClass           C.struct_AVClass
	AVLogLevel        C.int
	AVLogCallback     func(AVLogLevel, string, uintptr)
	AVDictionaryEntry C.struct_AVDictionaryEntry
	AVDictionaryFlag  int
	AVDictionary      struct {
		ctx *C.struct_AVDictionary
	}
	AVRational      C.struct_AVRational
	AVSampleFormat  C.enum_AVSampleFormat
	AVChannelOrder  C.enum_AVChannelOrder
	AVChannelCustom C.struct_AVChannelCustom
	AVChannel       C.enum_AVChannel
	AVChannelLayout C.struct_AVChannelLayout
	AVPixelFormat   C.enum_AVPixelFormat
	AVRounding      C.enum_AVRounding
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_LOG_QUIET   AVLogLevel = C.AV_LOG_QUIET
	AV_LOG_PANIC   AVLogLevel = C.AV_LOG_PANIC
	AV_LOG_FATAL   AVLogLevel = C.AV_LOG_FATAL
	AV_LOG_ERROR   AVLogLevel = C.AV_LOG_ERROR
	AV_LOG_WARNING AVLogLevel = C.AV_LOG_WARNING
	AV_LOG_INFO    AVLogLevel = C.AV_LOG_INFO
	AV_LOG_VERBOSE AVLogLevel = C.AV_LOG_VERBOSE
	AV_LOG_DEBUG   AVLogLevel = C.AV_LOG_DEBUG
	AV_LOG_TRACE   AVLogLevel = C.AV_LOG_TRACE
)

const (
	AV_DICT_MATCH_CASE      AVDictionaryFlag = C.AV_DICT_MATCH_CASE
	AV_DICT_IGNORE_SUFFIX   AVDictionaryFlag = C.AV_DICT_IGNORE_SUFFIX
	AV_DICT_DONT_STRDUP_KEY AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_KEY
	AV_DICT_DONT_STRDUP_VAL AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_VAL
	AV_DICT_DONT_OVERWRITE  AVDictionaryFlag = C.AV_DICT_DONT_OVERWRITE
	AV_DICT_APPEND          AVDictionaryFlag = C.AV_DICT_APPEND
	AV_DICT_MULTIKEY        AVDictionaryFlag = C.AV_DICT_MULTIKEY
)

const (
	AV_SAMPLE_FMT_NONE AVSampleFormat = C.AV_SAMPLE_FMT_NONE
	AV_SAMPLE_FMT_U8   AVSampleFormat = C.AV_SAMPLE_FMT_U8
	AV_SAMPLE_FMT_S16  AVSampleFormat = C.AV_SAMPLE_FMT_S16
	AV_SAMPLE_FMT_S32  AVSampleFormat = C.AV_SAMPLE_FMT_S32
	AV_SAMPLE_FMT_FLT  AVSampleFormat = C.AV_SAMPLE_FMT_FLT
	AV_SAMPLE_FMT_DBL  AVSampleFormat = C.AV_SAMPLE_FMT_DBL
	AV_SAMPLE_FMT_U8P  AVSampleFormat = C.AV_SAMPLE_FMT_U8P
	AV_SAMPLE_FMT_S16P AVSampleFormat = C.AV_SAMPLE_FMT_S16P
	AV_SAMPLE_FMT_S32P AVSampleFormat = C.AV_SAMPLE_FMT_S32P
	AV_SAMPLE_FMT_FLTP AVSampleFormat = C.AV_SAMPLE_FMT_FLTP
	AV_SAMPLE_FMT_DBLP AVSampleFormat = C.AV_SAMPLE_FMT_DBLP
	AV_SAMPLE_FMT_S64  AVSampleFormat = C.AV_SAMPLE_FMT_S64
	AV_SAMPLE_FMT_S64P AVSampleFormat = C.AV_SAMPLE_FMT_S64P
	AV_SAMPLE_FMT_NB   AVSampleFormat = C.AV_SAMPLE_FMT_NB
)

const (
	AV_CHANNEL_ORDER_UNSPEC    AVChannelOrder = C.AV_CHANNEL_ORDER_UNSPEC
	AV_CHANNEL_ORDER_NATIVE    AVChannelOrder = C.AV_CHANNEL_ORDER_NATIVE
	AV_CHANNEL_ORDER_CUSTOM    AVChannelOrder = C.AV_CHANNEL_ORDER_CUSTOM
	AV_CHANNEL_ORDER_AMBISONIC AVChannelOrder = C.AV_CHANNEL_ORDER_AMBISONIC
)

const (
	AV_CHAN_NONE                  AVChannel = C.AV_CHAN_NONE
	AV_CHAN_FRONT_LEFT            AVChannel = C.AV_CHAN_FRONT_LEFT
	AV_CHAN_FRONT_RIGHT           AVChannel = C.AV_CHAN_FRONT_RIGHT
	AV_CHAN_FRONT_CENTER          AVChannel = C.AV_CHAN_FRONT_CENTER
	AV_CHAN_LOW_FREQUENCY         AVChannel = C.AV_CHAN_LOW_FREQUENCY
	AV_CHAN_BACK_LEFT             AVChannel = C.AV_CHAN_BACK_LEFT
	AV_CHAN_BACK_RIGHT            AVChannel = C.AV_CHAN_BACK_RIGHT
	AV_CHAN_FRONT_LEFT_OF_CENTER  AVChannel = C.AV_CHAN_FRONT_LEFT_OF_CENTER
	AV_CHAN_FRONT_RIGHT_OF_CENTER AVChannel = C.AV_CHAN_FRONT_RIGHT_OF_CENTER
	AV_CHAN_BACK_CENTER           AVChannel = C.AV_CHAN_BACK_CENTER
	AV_CHAN_SIDE_LEFT             AVChannel = C.AV_CHAN_SIDE_LEFT
	AV_CHAN_SIDE_RIGHT            AVChannel = C.AV_CHAN_SIDE_RIGHT
	AV_CHAN_TOP_CENTER            AVChannel = C.AV_CHAN_TOP_CENTER
	AV_CHAN_TOP_FRONT_LEFT        AVChannel = C.AV_CHAN_TOP_FRONT_LEFT
	AV_CHAN_TOP_FRONT_CENTER      AVChannel = C.AV_CHAN_TOP_FRONT_CENTER
	AV_CHAN_TOP_FRONT_RIGHT       AVChannel = C.AV_CHAN_TOP_FRONT_RIGHT
	AV_CHAN_TOP_BACK_LEFT         AVChannel = C.AV_CHAN_TOP_BACK_LEFT
	AV_CHAN_TOP_BACK_CENTER       AVChannel = C.AV_CHAN_TOP_BACK_CENTER
	AV_CHAN_TOP_BACK_RIGHT        AVChannel = C.AV_CHAN_TOP_BACK_RIGHT
	AV_CHAN_STEREO_LEFT           AVChannel = C.AV_CHAN_STEREO_LEFT
	AV_CHAN_STEREO_RIGHT          AVChannel = C.AV_CHAN_STEREO_RIGHT
	AV_CHAN_WIDE_LEFT             AVChannel = C.AV_CHAN_WIDE_LEFT
	AV_CHAN_WIDE_RIGHT            AVChannel = C.AV_CHAN_WIDE_RIGHT
	AV_CHAN_SURROUND_DIRECT_LEFT  AVChannel = C.AV_CHAN_SURROUND_DIRECT_LEFT
	AV_CHAN_SURROUND_DIRECT_RIGHT AVChannel = C.AV_CHAN_SURROUND_DIRECT_RIGHT
	AV_CHAN_LOW_FREQUENCY_2       AVChannel = C.AV_CHAN_LOW_FREQUENCY_2
	AV_CHAN_TOP_SIDE_LEFT         AVChannel = C.AV_CHAN_TOP_SIDE_LEFT
	AV_CHAN_TOP_SIDE_RIGHT        AVChannel = C.AV_CHAN_TOP_SIDE_RIGHT
	AV_CHAN_BOTTOM_FRONT_CENTER   AVChannel = C.AV_CHAN_BOTTOM_FRONT_CENTER
	AV_CHAN_BOTTOM_FRONT_LEFT     AVChannel = C.AV_CHAN_BOTTOM_FRONT_LEFT
	AV_CHAN_BOTTOM_FRONT_RIGHT    AVChannel = C.AV_CHAN_BOTTOM_FRONT_RIGHT
	AV_CHAN_UNUSED                AVChannel = C.AV_CHAN_UNUSED
	AV_CHAN_UNKNOWN               AVChannel = C.AV_CHAN_UNKNOWN
	AV_CHAN_AMBISONIC_BASE        AVChannel = C.AV_CHAN_AMBISONIC_BASE
	AV_CHAN_AMBISONIC_END         AVChannel = C.AV_CHAN_AMBISONIC_END
)

const (
	AV_PIX_FMT_YUV420P        AVPixelFormat = C.AV_PIX_FMT_YUV420P        // planar YUV 4:2:0, 12bpp, (1 Cr & Cb sample per 2x2 Y samples)
	AV_PIX_FMT_YUYV422        AVPixelFormat = C.AV_PIX_FMT_YUYV422        // packed YUV 4:2:2, 16bpp, Y0 Cb Y1 Cr
	AV_PIX_FMT_RGB24          AVPixelFormat = C.AV_PIX_FMT_RGB24          // packed RGB 8:8:8, 24bpp, RGBRGB...
	AV_PIX_FMT_BGR24          AVPixelFormat = C.AV_PIX_FMT_BGR24          // packed RGB 8:8:8, 24bpp, BGRBGR...
	AV_PIX_FMT_YUV422P        AVPixelFormat = C.AV_PIX_FMT_YUV422P        // planar YUV 4:2:2, 16bpp, (1 Cr & Cb sample per 2x1 Y samples)
	AV_PIX_FMT_YUV444P        AVPixelFormat = C.AV_PIX_FMT_YUV444P        // planar YUV 4:4:4, 24bpp, (1 Cr & Cb sample per 1x1 Y samples)
	AV_PIX_FMT_YUV410P        AVPixelFormat = C.AV_PIX_FMT_YUV410P        // planar YUV 4:1:0, 9bpp, (1 Cr & Cb sample per 4x4 Y samples)
	AV_PIX_FMT_YUV411P        AVPixelFormat = C.AV_PIX_FMT_YUV411P        // planar YUV 4:1:1, 12bpp, (1 Cr & Cb sample per 4x1 Y samples)
	AV_PIX_FMT_GRAY8          AVPixelFormat = C.AV_PIX_FMT_GRAY8          // 8bpp.
	AV_PIX_FMT_MONOWHITE      AVPixelFormat = C.AV_PIX_FMT_MONOWHITE      // 1bpp, 0 is white, 1 is black, in each byte pixels are ordered from the msb to the lsb.
	AV_PIX_FMT_MONOBLACK      AVPixelFormat = C.AV_PIX_FMT_MONOBLACK      // 1bpp, 0 is black, 1 is white, in each byte pixels are ordered from the msb to the lsb.
	AV_PIX_FMT_PAL8           AVPixelFormat = C.AV_PIX_FMT_PAL8           // 8 bits with AV_PIX_FMT_RGB32alette
	AV_PIX_FMT_YUVJ420P       AVPixelFormat = C.AV_PIX_FMT_YUVJ420P       // planar YUV 4:2:0, 12bpp, full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV420P and setting color_range
	AV_PIX_FMT_YUVJ422P       AVPixelFormat = C.AV_PIX_FMT_YUVJ422P       // planar YUV 4:2:2, 16bpp, full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV422P and setting color_range
	AV_PIX_FMT_YUVJ444P       AVPixelFormat = C.AV_PIX_FMT_YUVJ444P       // planar YUV 4:4:4, 24bpp, full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV444P and setting color_range
	AV_PIX_FMT_UYVY422        AVPixelFormat = C.AV_PIX_FMT_UYVY422        // packed YUV 4:2:2, 16bpp, Cb Y0 Cr Y1
	AV_PIX_FMT_UYYVYY411      AVPixelFormat = C.AV_PIX_FMT_UYYVYY411      // packed YUV 4:1:1, 12bpp, Cb Y0 Y1 Cr Y2 Y3
	AV_PIX_FMT_BGR8           AVPixelFormat = C.AV_PIX_FMT_BGR8           // packed RGB 3:3:2, 8bpp, (msb)2B 3G 3R(lsb)
	AV_PIX_FMT_BGR4           AVPixelFormat = C.AV_PIX_FMT_BGR4           // packed RGB 1:2:1 bitstream, 4bpp, (msb)1B 2G 1R(lsb), a byte contains two pixels, the first pixel in the byte is the one composed by the 4 msb bits
	AV_PIX_FMT_BGR4_BYTE      AVPixelFormat = C.AV_PIX_FMT_BGR4_BYTE      // packed RGB 1:2:1, 8bpp, (msb)1B 2G 1R(lsb)
	AV_PIX_FMT_RGB8           AVPixelFormat = C.AV_PIX_FMT_RGB8           // packed RGB 3:3:2, 8bpp, (msb)2R 3G 3B(lsb)
	AV_PIX_FMT_RGB4           AVPixelFormat = C.AV_PIX_FMT_RGB4           // packed RGB 1:2:1 bitstream, 4bpp, (msb)1R 2G 1B(lsb), a byte contains two pixels, the first pixel in the byte is the one composed by the 4 msb bits
	AV_PIX_FMT_RGB4_BYTE      AVPixelFormat = C.AV_PIX_FMT_RGB4_BYTE      // packed RGB 1:2:1, 8bpp, (msb)1R 2G 1B(lsb)
	AV_PIX_FMT_NV12           AVPixelFormat = C.AV_PIX_FMT_NV12           // planar YUV 4:2:0, 12bpp, 1 plane for Y and 1 plane for the UV components, which are interleaved (first byte U and the following byte V)
	AV_PIX_FMT_NV21           AVPixelFormat = C.AV_PIX_FMT_NV21           // as above, but U and V bytes are swapped
	AV_PIX_FMT_ARGB           AVPixelFormat = C.AV_PIX_FMT_ARGB           // packed ARGB 8:8:8:8, 32bpp, ARGBARGB...
	AV_PIX_FMT_RGBA           AVPixelFormat = C.AV_PIX_FMT_RGBA           // packed RGBA 8:8:8:8, 32bpp, RGBARGBA...
	AV_PIX_FMT_ABGR           AVPixelFormat = C.AV_PIX_FMT_ABGR           // packed ABGR 8:8:8:8, 32bpp, ABGRABGR...
	AV_PIX_FMT_BGRA           AVPixelFormat = C.AV_PIX_FMT_BGRA           // packed BGRA 8:8:8:8, 32bpp, BGRABGRA...
	AV_PIX_FMT_GRAY16BE       AVPixelFormat = C.AV_PIX_FMT_GRAY16BE       // 16bpp, big-endian.
	AV_PIX_FMT_GRAY16LE       AVPixelFormat = C.AV_PIX_FMT_GRAY16LE       // 16bpp, little-endian.
	AV_PIX_FMT_YUV440P        AVPixelFormat = C.AV_PIX_FMT_YUV440P        // planar YUV 4:4:0 (1 Cr & Cb sample per 1x2 Y samples)
	AV_PIX_FMT_YUVJ440P       AVPixelFormat = C.AV_PIX_FMT_YUVJ440P       // planar YUV 4:4:0 full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV440P  and setting color_range
	AV_PIX_FMT_YUVA420P       AVPixelFormat = C.AV_PIX_FMT_YUVA420P       // planar YUV 4:2:0, 20bpp, (1 Cr & Cb sample per 2x2 Y & A samples)
	AV_PIX_FMT_RGB48BE        AVPixelFormat = C.AV_PIX_FMT_RGB48BE        // packed RGB 16:16:16, 48bpp, 16R, 16G, 16B, the 2-byte value for each R/G/B component is stored as big-endian
	AV_PIX_FMT_RGB48LE        AVPixelFormat = C.AV_PIX_FMT_RGB48LE        // packed RGB 16:16:16, 48bpp, 16R, 16G, 16B, the 2-byte value for each R/G/B component is stored as little-endian
	AV_PIX_FMT_RGB565BE       AVPixelFormat = C.AV_PIX_FMT_RGB565BE       // packed RGB 5:6:5, 16bpp, (msb) 5R 6G 5B(lsb), big-endian
	AV_PIX_FMT_RGB565LE       AVPixelFormat = C.AV_PIX_FMT_RGB565LE       // packed RGB 5:6:5, 16bpp, (msb) 5R 6G 5B(lsb), little-endian
	AV_PIX_FMT_RGB555BE       AVPixelFormat = C.AV_PIX_FMT_RGB555BE       // packed RGB 5:5:5, 16bpp, (msb)1X 5R 5G 5B(lsb), big-endian , X=unused/undefined
	AV_PIX_FMT_RGB555LE       AVPixelFormat = C.AV_PIX_FMT_RGB555LE       // packed RGB 5:5:5, 16bpp, (msb)1X 5R 5G 5B(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_BGR565BE       AVPixelFormat = C.AV_PIX_FMT_BGR565BE       // packed BGR 5:6:5, 16bpp, (msb) 5B 6G 5R(lsb), big-endian
	AV_PIX_FMT_BGR565LE       AVPixelFormat = C.AV_PIX_FMT_BGR565LE       // packed BGR 5:6:5, 16bpp, (msb) 5B 6G 5R(lsb), little-endian
	AV_PIX_FMT_BGR555BE       AVPixelFormat = C.AV_PIX_FMT_BGR555BE       // packed BGR 5:5:5, 16bpp, (msb)1X 5B 5G 5R(lsb), big-endian , X=unused/undefined
	AV_PIX_FMT_BGR555LE       AVPixelFormat = C.AV_PIX_FMT_BGR555LE       // packed BGR 5:5:5, 16bpp, (msb)1X 5B 5G 5R(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_VAAPI          AVPixelFormat = C.AV_PIX_FMT_VAAPI          //
	AV_PIX_FMT_YUV420P16LE    AVPixelFormat = C.AV_PIX_FMT_YUV420P16LE    // planar YUV 4:2:0, 24bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV420P16BE    AVPixelFormat = C.AV_PIX_FMT_YUV420P16BE    // planar YUV 4:2:0, 24bpp, (1 Cr & Cb sample per 2x2 Y samples), big-endian
	AV_PIX_FMT_YUV422P16LE    AVPixelFormat = C.AV_PIX_FMT_YUV422P16LE    // planar YUV 4:2:2, 32bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_YUV422P16BE    AVPixelFormat = C.AV_PIX_FMT_YUV422P16BE    // planar YUV 4:2:2, 32bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P16LE    AVPixelFormat = C.AV_PIX_FMT_YUV444P16LE    // planar YUV 4:4:4, 48bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P16BE    AVPixelFormat = C.AV_PIX_FMT_YUV444P16BE    // planar YUV 4:4:4, 48bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_DXVA2_VLD      AVPixelFormat = C.AV_PIX_FMT_DXVA2_VLD      // HW decoding through DXVA2, Picture.data[3] contains a LPDIRECT3DSURFACE9 pointer.
	AV_PIX_FMT_RGB444LE       AVPixelFormat = C.AV_PIX_FMT_RGB444LE       // packed RGB 4:4:4, 16bpp, (msb)4X 4R 4G 4B(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_RGB444BE       AVPixelFormat = C.AV_PIX_FMT_RGB444BE       // packed RGB 4:4:4, 16bpp, (msb)4X 4R 4G 4B(lsb), big-endian, X=unused/undefined
	AV_PIX_FMT_BGR444LE       AVPixelFormat = C.AV_PIX_FMT_BGR444LE       // packed BGR 4:4:4, 16bpp, (msb)4X 4B 4G 4R(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_BGR444BE       AVPixelFormat = C.AV_PIX_FMT_BGR444BE       // packed BGR 4:4:4, 16bpp, (msb)4X 4B 4G 4R(lsb), big-endian, X=unused/undefined
	AV_PIX_FMT_YA8            AVPixelFormat = C.AV_PIX_FMT_YA8            // 8 bits gray, 8 bits alpha
	AV_PIX_FMT_Y400A          AVPixelFormat = C.AV_PIX_FMT_Y400A          // alias for AV_PIX_FMT_YA8
	AV_PIX_FMT_GRAY8A         AVPixelFormat = C.AV_PIX_FMT_GRAY8A         // alias for AV_PIX_FMT_YA8
	AV_PIX_FMT_BGR48BE        AVPixelFormat = C.AV_PIX_FMT_BGR48BE        // packed RGB 16:16:16, 48bpp, 16B, 16G, 16R, the 2-byte value for each R/G/B component is stored as big-endian
	AV_PIX_FMT_BGR48LE        AVPixelFormat = C.AV_PIX_FMT_BGR48LE        // packed RGB 16:16:16, 48bpp, 16B, 16G, 16R, the 2-byte value for each R/G/B component is stored as little-endian
	AV_PIX_FMT_YUV420P9BE     AVPixelFormat = C.AV_PIX_FMT_YUV420P9BE     // The following 12 formats have the disadvantage of needing 1 format for each bit depth.
	AV_PIX_FMT_YUV420P9LE     AVPixelFormat = C.AV_PIX_FMT_YUV420P9LE     // planar YUV 4:2:0, 13.5bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV420P10BE    AVPixelFormat = C.AV_PIX_FMT_YUV420P10BE    // planar YUV 4:2:0, 15bpp, (1 Cr & Cb sample per 2x2 Y samples), big-endian
	AV_PIX_FMT_YUV420P10LE    AVPixelFormat = C.AV_PIX_FMT_YUV420P10LE    // planar YUV 4:2:0, 15bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV422P10BE    AVPixelFormat = C.AV_PIX_FMT_YUV422P10BE    // planar YUV 4:2:2, 20bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV422P10LE    AVPixelFormat = C.AV_PIX_FMT_YUV422P10LE    // planar YUV 4:2:2, 20bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P9BE     AVPixelFormat = C.AV_PIX_FMT_YUV444P9BE     // planar YUV 4:4:4, 27bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P9LE     AVPixelFormat = C.AV_PIX_FMT_YUV444P9LE     // planar YUV 4:4:4, 27bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P10BE    AVPixelFormat = C.AV_PIX_FMT_YUV444P10BE    // planar YUV 4:4:4, 30bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P10LE    AVPixelFormat = C.AV_PIX_FMT_YUV444P10LE    // planar YUV 4:4:4, 30bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_YUV422P9BE     AVPixelFormat = C.AV_PIX_FMT_YUV422P9BE     // planar YUV 4:2:2, 18bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV422P9LE     AVPixelFormat = C.AV_PIX_FMT_YUV422P9LE     // planar YUV 4:2:2, 18bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_GBRP           AVPixelFormat = C.AV_PIX_FMT_GBRP           //
	AV_PIX_FMT_GBR24P         AVPixelFormat = C.AV_PIX_FMT_GBR24P         // planar GBR 4:4:4 24bpp
	AV_PIX_FMT_GBRP9BE        AVPixelFormat = C.AV_PIX_FMT_GBRP9BE        // planar GBR 4:4:4 27bpp, big-endian
	AV_PIX_FMT_GBRP9LE        AVPixelFormat = C.AV_PIX_FMT_GBRP9LE        // planar GBR 4:4:4 27bpp, little-endian
	AV_PIX_FMT_GBRP10BE       AVPixelFormat = C.AV_PIX_FMT_GBRP10BE       // planar GBR 4:4:4 30bpp, big-endian
	AV_PIX_FMT_GBRP10LE       AVPixelFormat = C.AV_PIX_FMT_GBRP10LE       // planar GBR 4:4:4 30bpp, little-endian
	AV_PIX_FMT_GBRP16BE       AVPixelFormat = C.AV_PIX_FMT_GBRP16BE       // planar GBR 4:4:4 48bpp, big-endian
	AV_PIX_FMT_GBRP16LE       AVPixelFormat = C.AV_PIX_FMT_GBRP16LE       // planar GBR 4:4:4 48bpp, little-endian
	AV_PIX_FMT_YUVA422P       AVPixelFormat = C.AV_PIX_FMT_YUVA422P       // planar YUV 4:2:2 24bpp, (1 Cr & Cb sample per 2x1 Y & A samples)
	AV_PIX_FMT_YUVA444P       AVPixelFormat = C.AV_PIX_FMT_YUVA444P       // planar YUV 4:4:4 32bpp, (1 Cr & Cb sample per 1x1 Y & A samples)
	AV_PIX_FMT_YUVA420P9BE    AVPixelFormat = C.AV_PIX_FMT_YUVA420P9BE    // planar YUV 4:2:0 22.5bpp, (1 Cr & Cb sample per 2x2 Y & A samples), big-endian
	AV_PIX_FMT_YUVA420P9LE    AVPixelFormat = C.AV_PIX_FMT_YUVA420P9LE    // planar YUV 4:2:0 22.5bpp, (1 Cr & Cb sample per 2x2 Y & A samples), little-endian
	AV_PIX_FMT_YUVA422P9BE    AVPixelFormat = C.AV_PIX_FMT_YUVA422P9BE    // planar YUV 4:2:2 27bpp, (1 Cr & Cb sample per 2x1 Y & A samples), big-endian
	AV_PIX_FMT_YUVA422P9LE    AVPixelFormat = C.AV_PIX_FMT_YUVA422P9LE    // planar YUV 4:2:2 27bpp, (1 Cr & Cb sample per 2x1 Y & A samples), little-endian
	AV_PIX_FMT_YUVA444P9BE    AVPixelFormat = C.AV_PIX_FMT_YUVA444P9BE    // planar YUV 4:4:4 36bpp, (1 Cr & Cb sample per 1x1 Y & A samples), big-endian
	AV_PIX_FMT_YUVA444P9LE    AVPixelFormat = C.AV_PIX_FMT_YUVA444P9LE    // planar YUV 4:4:4 36bpp, (1 Cr & Cb sample per 1x1 Y & A samples), little-endian
	AV_PIX_FMT_YUVA420P10BE   AVPixelFormat = C.AV_PIX_FMT_YUVA420P10BE   // planar YUV 4:2:0 25bpp, (1 Cr & Cb sample per 2x2 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA420P10LE   AVPixelFormat = C.AV_PIX_FMT_YUVA420P10LE   // planar YUV 4:2:0 25bpp, (1 Cr & Cb sample per 2x2 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA422P10BE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P10BE   // planar YUV 4:2:2 30bpp, (1 Cr & Cb sample per 2x1 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA422P10LE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P10LE   // planar YUV 4:2:2 30bpp, (1 Cr & Cb sample per 2x1 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA444P10BE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P10BE   // planar YUV 4:4:4 40bpp, (1 Cr & Cb sample per 1x1 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA444P10LE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P10LE   // planar YUV 4:4:4 40bpp, (1 Cr & Cb sample per 1x1 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA420P16BE   AVPixelFormat = C.AV_PIX_FMT_YUVA420P16BE   // planar YUV 4:2:0 40bpp, (1 Cr & Cb sample per 2x2 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA420P16LE   AVPixelFormat = C.AV_PIX_FMT_YUVA420P16LE   // planar YUV 4:2:0 40bpp, (1 Cr & Cb sample per 2x2 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA422P16BE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P16BE   // planar YUV 4:2:2 48bpp, (1 Cr & Cb sample per 2x1 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA422P16LE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P16LE   // planar YUV 4:2:2 48bpp, (1 Cr & Cb sample per 2x1 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA444P16BE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P16BE   // planar YUV 4:4:4 64bpp, (1 Cr & Cb sample per 1x1 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA444P16LE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P16LE   // planar YUV 4:4:4 64bpp, (1 Cr & Cb sample per 1x1 Y & A samples, little-endian)
	AV_PIX_FMT_VDPAU          AVPixelFormat = C.AV_PIX_FMT_VDPAU          // HW acceleration through VDPAU, Picture.data[3] contains a VdpVideoSurface.
	AV_PIX_FMT_XYZ12LE        AVPixelFormat = C.AV_PIX_FMT_XYZ12LE        // packed XYZ 4:4:4, 36 bpp, (msb) 12X, 12Y, 12Z (lsb), the 2-byte value for each X/Y/Z is stored as little-endian, the 4 lower bits are set to 0
	AV_PIX_FMT_XYZ12BE        AVPixelFormat = C.AV_PIX_FMT_XYZ12BE        // packed XYZ 4:4:4, 36 bpp, (msb) 12X, 12Y, 12Z (lsb), the 2-byte value for each X/Y/Z is stored as big-endian, the 4 lower bits are set to 0
	AV_PIX_FMT_NV16           AVPixelFormat = C.AV_PIX_FMT_NV16           // interleaved chroma YUV 4:2:2, 16bpp, (1 Cr & Cb sample per 2x1 Y samples)
	AV_PIX_FMT_NV20LE         AVPixelFormat = C.AV_PIX_FMT_NV20LE         // interleaved chroma YUV 4:2:2, 20bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_NV20BE         AVPixelFormat = C.AV_PIX_FMT_NV20BE         // interleaved chroma YUV 4:2:2, 20bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_RGBA64BE       AVPixelFormat = C.AV_PIX_FMT_RGBA64BE       // packed RGBA 16:16:16:16, 64bpp, 16R, 16G, 16B, 16A, the 2-byte value for each R/G/B/A component is stored as big-endian
	AV_PIX_FMT_RGBA64LE       AVPixelFormat = C.AV_PIX_FMT_RGBA64LE       // packed RGBA 16:16:16:16, 64bpp, 16R, 16G, 16B, 16A, the 2-byte value for each R/G/B/A component is stored as little-endian
	AV_PIX_FMT_BGRA64BE       AVPixelFormat = C.AV_PIX_FMT_BGRA64BE       // packed RGBA 16:16:16:16, 64bpp, 16B, 16G, 16R, 16A, the 2-byte value for each R/G/B/A component is stored as big-endian
	AV_PIX_FMT_BGRA64LE       AVPixelFormat = C.AV_PIX_FMT_BGRA64LE       // packed RGBA 16:16:16:16, 64bpp, 16B, 16G, 16R, 16A, the 2-byte value for each R/G/B/A component is stored as little-endian
	AV_PIX_FMT_YVYU422        AVPixelFormat = C.AV_PIX_FMT_YVYU422        // packed YUV 4:2:2, 16bpp, Y0 Cr Y1 Cb
	AV_PIX_FMT_YA16BE         AVPixelFormat = C.AV_PIX_FMT_YA16BE         // 16 bits gray, 16 bits alpha (big-endian)
	AV_PIX_FMT_YA16LE         AVPixelFormat = C.AV_PIX_FMT_YA16LE         // 16 bits gray, 16 bits alpha (little-endian)
	AV_PIX_FMT_GBRAP          AVPixelFormat = C.AV_PIX_FMT_GBRAP          // planar GBRA 4:4:4:4 32bpp
	AV_PIX_FMT_GBRAP16BE      AVPixelFormat = C.AV_PIX_FMT_GBRAP16BE      // planar GBRA 4:4:4:4 64bpp, big-endian
	AV_PIX_FMT_GBRAP16LE      AVPixelFormat = C.AV_PIX_FMT_GBRAP16LE      // planar GBRA 4:4:4:4 64bpp, little-endian
	AV_PIX_FMT_QSV            AVPixelFormat = C.AV_PIX_FMT_QSV            // HW acceleration through QSV, data[3] contains a pointer to the mfxFrameSurface1 structure.
	AV_PIX_FMT_MMAL           AVPixelFormat = C.AV_PIX_FMT_MMAL           // HW acceleration though MMAL, data[3] contains a pointer to the MMAL_BUFFER_HEADER_T structure.
	AV_PIX_FMT_D3D11VA_VLD    AVPixelFormat = C.AV_PIX_FMT_D3D11VA_VLD    // HW decoding through Direct3D11 via old API, Picture.data[3] contains a ID3D11VideoDecoderOutputView pointer.
	AV_PIX_FMT_CUDA           AVPixelFormat = C.AV_PIX_FMT_CUDA           // HW acceleration through CUDA.
	AV_PIX_FMT_0RGB           AVPixelFormat = C.AV_PIX_FMT_0RGB           // packed RGB 8:8:8, 32bpp, XRGBXRGB... X=unused/undefined
	AV_PIX_FMT_RGB0           AVPixelFormat = C.AV_PIX_FMT_RGB0           // packed RGB 8:8:8, 32bpp, RGBXRGBX... X=unused/undefined
	AV_PIX_FMT_0BGR           AVPixelFormat = C.AV_PIX_FMT_0BGR           // packed BGR 8:8:8, 32bpp, XBGRXBGR... X=unused/undefined
	AV_PIX_FMT_BGR0           AVPixelFormat = C.AV_PIX_FMT_BGR0           // packed BGR 8:8:8, 32bpp, BGRXBGRX... X=unused/undefined
	AV_PIX_FMT_YUV420P12BE    AVPixelFormat = C.AV_PIX_FMT_YUV420P12BE    // planar YUV 4:2:0,18bpp, (1 Cr & Cb sample per 2x2 Y samples), big-endian
	AV_PIX_FMT_YUV420P12LE    AVPixelFormat = C.AV_PIX_FMT_YUV420P12LE    // planar YUV 4:2:0,18bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV420P14BE    AVPixelFormat = C.AV_PIX_FMT_YUV420P14BE    // planar YUV 4:2:0,21bpp, (1 Cr & Cb sample per 2x2 Y samples), big-endian
	AV_PIX_FMT_YUV420P14LE    AVPixelFormat = C.AV_PIX_FMT_YUV420P14LE    // planar YUV 4:2:0,21bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV422P12BE    AVPixelFormat = C.AV_PIX_FMT_YUV422P12BE    // planar YUV 4:2:2,24bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV422P12LE    AVPixelFormat = C.AV_PIX_FMT_YUV422P12LE    // planar YUV 4:2:2,24bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_YUV422P14BE    AVPixelFormat = C.AV_PIX_FMT_YUV422P14BE    // planar YUV 4:2:2,28bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV422P14LE    AVPixelFormat = C.AV_PIX_FMT_YUV422P14LE    // planar YUV 4:2:2,28bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P12BE    AVPixelFormat = C.AV_PIX_FMT_YUV444P12BE    // planar YUV 4:4:4,36bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P12LE    AVPixelFormat = C.AV_PIX_FMT_YUV444P12LE    // planar YUV 4:4:4,36bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P14BE    AVPixelFormat = C.AV_PIX_FMT_YUV444P14BE    // planar YUV 4:4:4,42bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P14LE    AVPixelFormat = C.AV_PIX_FMT_YUV444P14LE    // planar YUV 4:4:4,42bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_GBRP12BE       AVPixelFormat = C.AV_PIX_FMT_GBRP12BE       // planar GBR 4:4:4 36bpp, big-endian
	AV_PIX_FMT_GBRP12LE       AVPixelFormat = C.AV_PIX_FMT_GBRP12LE       // planar GBR 4:4:4 36bpp, little-endian
	AV_PIX_FMT_GBRP14BE       AVPixelFormat = C.AV_PIX_FMT_GBRP14BE       // planar GBR 4:4:4 42bpp, big-endian
	AV_PIX_FMT_GBRP14LE       AVPixelFormat = C.AV_PIX_FMT_GBRP14LE       // planar GBR 4:4:4 42bpp, little-endian
	AV_PIX_FMT_YUVJ411P       AVPixelFormat = C.AV_PIX_FMT_YUVJ411P       // planar YUV 4:1:1, 12bpp, (1 Cr & Cb sample per 4x1 Y samples) full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV411P AVPixelFormat = C.AV_PIX_FMT_YUV411P and setting color_range
	AV_PIX_FMT_BAYER_BGGR8    AVPixelFormat = C.AV_PIX_FMT_BAYER_BGGR8    // bayer, BGBG..(odd line), GRGR..(even line), 8-bit samples
	AV_PIX_FMT_BAYER_RGGB8    AVPixelFormat = C.AV_PIX_FMT_BAYER_RGGB8    // bayer, RGRG..(odd line), GBGB..(even line), 8-bit samples
	AV_PIX_FMT_BAYER_GBRG8    AVPixelFormat = C.AV_PIX_FMT_BAYER_GBRG8    // bayer, GBGB..(odd line), RGRG..(even line), 8-bit samples
	AV_PIX_FMT_BAYER_GRBG8    AVPixelFormat = C.AV_PIX_FMT_BAYER_GRBG8    // bayer, GRGR..(odd line), BGBG..(even line), 8-bit samples
	AV_PIX_FMT_BAYER_BGGR16LE AVPixelFormat = C.AV_PIX_FMT_BAYER_BGGR16LE // bayer, BGBG..(odd line), GRGR..(even line), 16-bit samples, little-endian
	AV_PIX_FMT_BAYER_BGGR16BE AVPixelFormat = C.AV_PIX_FMT_BAYER_BGGR16BE // bayer, BGBG..(odd line), GRGR..(even line), 16-bit samples, big-endian
	AV_PIX_FMT_BAYER_RGGB16LE AVPixelFormat = C.AV_PIX_FMT_BAYER_RGGB16LE // bayer, RGRG..(odd line), GBGB..(even line), 16-bit samples, little-endian
	AV_PIX_FMT_BAYER_RGGB16BE AVPixelFormat = C.AV_PIX_FMT_BAYER_RGGB16BE // bayer, RGRG..(odd line), GBGB..(even line), 16-bit samples, big-endian
	AV_PIX_FMT_BAYER_GBRG16LE AVPixelFormat = C.AV_PIX_FMT_BAYER_GBRG16LE // bayer, GBGB..(odd line), RGRG..(even line), 16-bit samples, little-endian
	AV_PIX_FMT_BAYER_GBRG16BE AVPixelFormat = C.AV_PIX_FMT_BAYER_GBRG16BE // bayer, GBGB..(odd line), RGRG..(even line), 16-bit samples, big-endian
	AV_PIX_FMT_BAYER_GRBG16LE AVPixelFormat = C.AV_PIX_FMT_BAYER_GRBG16LE // bayer, GRGR..(odd line), BGBG..(even line), 16-bit samples, little-endian
	AV_PIX_FMT_BAYER_GRBG16BE AVPixelFormat = C.AV_PIX_FMT_BAYER_GRBG16BE // bayer, GRGR..(odd line), BGBG..(even line), 16-bit samples, big-endian
	AV_PIX_FMT_XVMC           AVPixelFormat = C.AV_PIX_FMT_XVMC           // XVideo Motion Acceleration via common packet passing.
	AV_PIX_FMT_YUV440P10LE    AVPixelFormat = C.AV_PIX_FMT_YUV440P10LE    // planar YUV 4:4:0,20bpp, (1 Cr & Cb sample per 1x2 Y samples), little-endian
	AV_PIX_FMT_YUV440P10BE    AVPixelFormat = C.AV_PIX_FMT_YUV440P10BE    // planar YUV 4:4:0,20bpp, (1 Cr & Cb sample per 1x2 Y samples), big-endian
	AV_PIX_FMT_YUV440P12LE    AVPixelFormat = C.AV_PIX_FMT_YUV440P12LE    // planar YUV 4:4:0,24bpp, (1 Cr & Cb sample per 1x2 Y samples), little-endian
	AV_PIX_FMT_YUV440P12BE    AVPixelFormat = C.AV_PIX_FMT_YUV440P12BE    // planar YUV 4:4:0,24bpp, (1 Cr & Cb sample per 1x2 Y samples), big-endian
	AV_PIX_FMT_AYUV64LE       AVPixelFormat = C.AV_PIX_FMT_AYUV64LE       // packed AYUV 4:4:4,64bpp (1 Cr & Cb sample per 1x1 Y & A samples), little-endian
	AV_PIX_FMT_AYUV64BE       AVPixelFormat = C.AV_PIX_FMT_AYUV64BE       // packed AYUV 4:4:4,64bpp (1 Cr & Cb sample per 1x1 Y & A samples), big-endian
	AV_PIX_FMT_VIDEOTOOLBOX   AVPixelFormat = C.AV_PIX_FMT_VIDEOTOOLBOX   // hardware decoding through Videotoolbox
	AV_PIX_FMT_P010LE         AVPixelFormat = C.AV_PIX_FMT_P010LE         // like NV12, with 10bpp per component, data in the high bits, zeros in the low bits, little-endian
	AV_PIX_FMT_P010BE         AVPixelFormat = C.AV_PIX_FMT_P010BE         // like NV12, with 10bpp per component, data in the high bits, zeros in the low bits, big-endian
	AV_PIX_FMT_GBRAP12BE      AVPixelFormat = C.AV_PIX_FMT_GBRAP12BE      // planar GBR 4:4:4:4 48bpp, big-endian
	AV_PIX_FMT_GBRAP12LE      AVPixelFormat = C.AV_PIX_FMT_GBRAP12LE      // planar GBR 4:4:4:4 48bpp, little-endian
	AV_PIX_FMT_GBRAP10BE      AVPixelFormat = C.AV_PIX_FMT_GBRAP10BE      // planar GBR 4:4:4:4 40bpp, big-endian
	AV_PIX_FMT_GBRAP10LE      AVPixelFormat = C.AV_PIX_FMT_GBRAP10LE      // planar GBR 4:4:4:4 40bpp, little-endian
	AV_PIX_FMT_MEDIACODEC     AVPixelFormat = C.AV_PIX_FMT_MEDIACODEC     // hardware decoding through MediaCodec
	AV_PIX_FMT_GRAY12BE       AVPixelFormat = C.AV_PIX_FMT_GRAY12BE       // Y , 12bpp, big-endian.
	AV_PIX_FMT_GRAY12LE       AVPixelFormat = C.AV_PIX_FMT_GRAY12LE       // Y , 12bpp, little-endian.
	AV_PIX_FMT_GRAY10BE       AVPixelFormat = C.AV_PIX_FMT_GRAY10BE       // Y , 10bpp, big-endian.
	AV_PIX_FMT_GRAY10LE       AVPixelFormat = C.AV_PIX_FMT_GRAY10LE       // Y , 10bpp, little-endian.
	AV_PIX_FMT_P016LE         AVPixelFormat = C.AV_PIX_FMT_P016LE         // like NV12, with 16bpp per component, little-endian
	AV_PIX_FMT_P016BE         AVPixelFormat = C.AV_PIX_FMT_P016BE         // like NV12, with 16bpp per component, big-endian
	AV_PIX_FMT_D3D11          AVPixelFormat = C.AV_PIX_FMT_D3D11          // Hardware surfaces for Direct3D11.
	AV_PIX_FMT_GRAY9BE        AVPixelFormat = C.AV_PIX_FMT_GRAY9BE        // Y , 9bpp, big-endian.
	AV_PIX_FMT_GRAY9LE        AVPixelFormat = C.AV_PIX_FMT_GRAY9LE        // Y , 9bpp, little-endian.
	AV_PIX_FMT_GBRPF32BE      AVPixelFormat = C.AV_PIX_FMT_GBRPF32BE      // IEEE-754 single precision planar GBR 4:4:4, 96bpp, big-endian.
	AV_PIX_FMT_GBRPF32LE      AVPixelFormat = C.AV_PIX_FMT_GBRPF32LE      // IEEE-754 single precision planar GBR 4:4:4, 96bpp, little-endian.
	AV_PIX_FMT_GBRAPF32BE     AVPixelFormat = C.AV_PIX_FMT_GBRAPF32BE     // IEEE-754 single precision planar GBRA 4:4:4:4, 128bpp, big-endian.
	AV_PIX_FMT_GBRAPF32LE     AVPixelFormat = C.AV_PIX_FMT_GBRAPF32LE     // IEEE-754 single precision planar GBRA 4:4:4:4, 128bpp, little-endian.
	AV_PIX_FMT_DRM_PRIME      AVPixelFormat = C.AV_PIX_FMT_DRM_PRIME      // DRM-managed buffers exposed through PRIME buffer sharing.
	AV_PIX_FMT_OPENCL         AVPixelFormat = C.AV_PIX_FMT_OPENCL         // Hardware surfaces for OpenCL.
	AV_PIX_FMT_GRAY14BE       AVPixelFormat = C.AV_PIX_FMT_GRAY14BE       // Y , 14bpp, big-endian.
	AV_PIX_FMT_GRAY14LE       AVPixelFormat = C.AV_PIX_FMT_GRAY14LE       // Y , 14bpp, little-endian.
	AV_PIX_FMT_GRAYF32BE      AVPixelFormat = C.AV_PIX_FMT_GRAYF32BE      // IEEE-754 single precision Y, 32bpp, big-endian.
	AV_PIX_FMT_GRAYF32LE      AVPixelFormat = C.AV_PIX_FMT_GRAYF32LE      // IEEE-754 single precision Y, 32bpp, little-endian.
	AV_PIX_FMT_NONE           AVPixelFormat = C.AV_PIX_FMT_NONE
)

var (
	AV_CHANNEL_LAYOUT_MONO                  = AVChannelLayout(C._AV_CHANNEL_LAYOUT_MONO)
	AV_CHANNEL_LAYOUT_STEREO                = AVChannelLayout(C._AV_CHANNEL_LAYOUT_STEREO)
	AV_CHANNEL_LAYOUT_2POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_2POINT1)
	AV_CHANNEL_LAYOUT_2_1                   = AVChannelLayout(C._AV_CHANNEL_LAYOUT_2_1)
	AV_CHANNEL_LAYOUT_SURROUND              = AVChannelLayout(C._AV_CHANNEL_LAYOUT_SURROUND)
	AV_CHANNEL_LAYOUT_3POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_3POINT1)
	AV_CHANNEL_LAYOUT_4POINT0               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_4POINT0)
	AV_CHANNEL_LAYOUT_4POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_4POINT1)
	AV_CHANNEL_LAYOUT_2_2                   = AVChannelLayout(C._AV_CHANNEL_LAYOUT_2_2)
	AV_CHANNEL_LAYOUT_QUAD                  = AVChannelLayout(C._AV_CHANNEL_LAYOUT_QUAD)
	AV_CHANNEL_LAYOUT_5POINT0               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT0)
	AV_CHANNEL_LAYOUT_5POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT1)
	AV_CHANNEL_LAYOUT_5POINT0_BACK          = AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT0_BACK)
	AV_CHANNEL_LAYOUT_5POINT1_BACK          = AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT1_BACK)
	AV_CHANNEL_LAYOUT_6POINT0               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT0)
	AV_CHANNEL_LAYOUT_6POINT0_FRONT         = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT0_FRONT)
	AV_CHANNEL_LAYOUT_HEXAGONAL             = AVChannelLayout(C._AV_CHANNEL_LAYOUT_HEXAGONAL)
	AV_CHANNEL_LAYOUT_6POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT1)
	AV_CHANNEL_LAYOUT_6POINT1_BACK          = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT1_BACK)
	AV_CHANNEL_LAYOUT_6POINT1_FRONT         = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT1_FRONT)
	AV_CHANNEL_LAYOUT_7POINT0               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT0)
	AV_CHANNEL_LAYOUT_7POINT0_FRONT         = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT0_FRONT)
	AV_CHANNEL_LAYOUT_7POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT1)
	AV_CHANNEL_LAYOUT_7POINT1_WIDE          = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT1_WIDE)
	AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK     = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK)
	AV_CHANNEL_LAYOUT_OCTAGONAL             = AVChannelLayout(C._AV_CHANNEL_LAYOUT_OCTAGONAL)
	AV_CHANNEL_LAYOUT_HEXADECAGONAL         = AVChannelLayout(C._AV_CHANNEL_LAYOUT_HEXADECAGONAL)
	AV_CHANNEL_LAYOUT_STEREO_DOWNMIX        = AVChannelLayout(C._AV_CHANNEL_LAYOUT_STEREO_DOWNMIX)
	AV_CHANNEL_LAYOUT_22POINT2              = AVChannelLayout(C._AV_CHANNEL_LAYOUT_22POINT2)
	AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER = AVChannelLayout(C._AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER)
)

const (
	AV_ROUND_ZERO        = C.AV_ROUND_ZERO        ///< Round toward zero.
	AV_ROUND_INF         = C.AV_ROUNT_INF         ///< Round away from zero.
	AV_ROUND_DOWN        = C.AV_ROUND_DOWN        ///< Round toward -infinity.
	AV_ROUND_UP          = C.AV_ROUND_UP          ///< Round toward +infinity.
	AV_ROUND_NEAR_INF    = C.AV_ROUND_NEAR_INF    ///< Round to nearest and halfway cases away from zero.
	AV_ROUND_PASS_MINMAX = C.AV_ROUND_PASS_MINMAX ///< Flag telling rescaling functions to pass INT64_MIN/MAX through unchanged
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVLogLevel) String() string {
	switch v {
	case AV_LOG_QUIET:
		return "AV_LOG_QUIET"
	case AV_LOG_PANIC:
		return "AV_LOG_PANIC"
	case AV_LOG_FATAL:
		return "AV_LOG_FATAL"
	case AV_LOG_ERROR:
		return "AV_LOG_ERROR"
	case AV_LOG_WARNING:
		return "AV_LOG_WARNING"
	case AV_LOG_INFO:
		return "AV_LOG_INFO"
	case AV_LOG_VERBOSE:
		return "AV_LOG_VERBOSE"
	case AV_LOG_DEBUG:
		return "AV_LOG_DEBUG"
	case AV_LOG_TRACE:
		return "AV_LOG_TRACE"
	default:
		return "[?? Invalid AVLogLevel value]"
	}
}

func (v AVSampleFormat) String() string {
	switch v {
	case AV_SAMPLE_FMT_NONE:
		return "AV_SAMPLE_FMT_NONE"
	case AV_SAMPLE_FMT_U8:
		return "AV_SAMPLE_FMT_U8"
	case AV_SAMPLE_FMT_S16:
		return "AV_SAMPLE_FMT_S16"
	case AV_SAMPLE_FMT_S32:
		return "AV_SAMPLE_FMT_S32"
	case AV_SAMPLE_FMT_FLT:
		return "AV_SAMPLE_FMT_FLT"
	case AV_SAMPLE_FMT_DBL:
		return "AV_SAMPLE_FMT_DBL"
	case AV_SAMPLE_FMT_U8P:
		return "AV_SAMPLE_FMT_U8P"
	case AV_SAMPLE_FMT_S16P:
		return "AV_SAMPLE_FMT_S16P"
	case AV_SAMPLE_FMT_S32P:
		return "AV_SAMPLE_FMT_S32P"
	case AV_SAMPLE_FMT_FLTP:
		return "AV_SAMPLE_FMT_FLTP"
	case AV_SAMPLE_FMT_DBLP:
		return "AV_SAMPLE_FMT_DBLP"
	case AV_SAMPLE_FMT_S64:
		return "AV_SAMPLE_FMT_S64"
	case AV_SAMPLE_FMT_S64P:
		return "AV_SAMPLE_FMT_S64P"
	case AV_SAMPLE_FMT_NB:
		return "AV_SAMPLE_FMT_NB"
	default:
		return "[?? Invalid AVSampleFormat value]"
	}
}

func (v AVChannelOrder) String() string {
	switch v {
	case AV_CHANNEL_ORDER_UNSPEC:
		return "AV_CHANNEL_ORDER_UNSPEC"
	case AV_CHANNEL_ORDER_NATIVE:
		return "AV_CHANNEL_ORDER_NATIVE"
	case AV_CHANNEL_ORDER_CUSTOM:
		return "AV_CHANNEL_ORDER_CUSTOM"
	case AV_CHANNEL_ORDER_AMBISONIC:
		return "AV_CHANNEL_ORDER_AMBISONIC"
	default:
		return "[?? Invalid AVChannelOrder value]"
	}
}

func (v AVChannel) String() string {
	switch v {
	case AV_CHAN_NONE:
		return "AV_CHAN_NONE"
	case AV_CHAN_FRONT_LEFT:
		return "AV_CHAN_FRONT_LEFT"
	case AV_CHAN_FRONT_RIGHT:
		return "AV_CHAN_FRONT_RIGHT"
	case AV_CHAN_FRONT_CENTER:
		return "AV_CHAN_FRONT_CENTER"
	case AV_CHAN_LOW_FREQUENCY:
		return "AV_CHAN_LOW_FREQUENCY"
	case AV_CHAN_BACK_LEFT:
		return "AV_CHAN_BACK_LEFT"
	case AV_CHAN_BACK_RIGHT:
		return "AV_CHAN_BACK_RIGHT"
	case AV_CHAN_FRONT_LEFT_OF_CENTER:
		return "AV_CHAN_FRONT_LEFT_OF_CENTER"
	case AV_CHAN_FRONT_RIGHT_OF_CENTER:
		return "AV_CHAN_FRONT_RIGHT_OF_CENTER"
	case AV_CHAN_BACK_CENTER:
		return "AV_CHAN_BACK_CENTER"
	case AV_CHAN_SIDE_LEFT:
		return "AV_CHAN_SIDE_LEFT"
	case AV_CHAN_SIDE_RIGHT:
		return "AV_CHAN_SIDE_RIGHT"
	case AV_CHAN_TOP_CENTER:
		return "AV_CHAN_TOP_CENTER"
	case AV_CHAN_TOP_FRONT_LEFT:
		return "AV_CHAN_TOP_FRONT_LEFT"
	case AV_CHAN_TOP_FRONT_CENTER:
		return "AV_CHAN_TOP_FRONT_CENTER"
	case AV_CHAN_TOP_FRONT_RIGHT:
		return "AV_CHAN_TOP_FRONT_RIGHT"
	case AV_CHAN_TOP_BACK_LEFT:
		return "AV_CHAN_TOP_BACK_LEFT"
	case AV_CHAN_TOP_BACK_CENTER:
		return "AV_CHAN_TOP_BACK_CENTER"
	case AV_CHAN_TOP_BACK_RIGHT:
		return "AV_CHAN_TOP_BACK_RIGHT"
	case AV_CHAN_STEREO_LEFT:
		return "AV_CHAN_STEREO_LEFT"
	case AV_CHAN_STEREO_RIGHT:
		return "AV_CHAN_STEREO_RIGHT"
	case AV_CHAN_WIDE_LEFT:
		return "AV_CHAN_WIDE_LEFT"
	case AV_CHAN_WIDE_RIGHT:
		return "AV_CHAN_WIDE_RIGHT"
	case AV_CHAN_SURROUND_DIRECT_LEFT:
		return "AV_CHAN_SURROUND_DIRECT_LEFT"
	case AV_CHAN_SURROUND_DIRECT_RIGHT:
		return "AV_CHAN_SURROUND_DIRECT_RIGHT"
	case AV_CHAN_LOW_FREQUENCY_2:
		return "AV_CHAN_LOW_FREQUENCY_2"
	case AV_CHAN_TOP_SIDE_LEFT:
		return "AV_CHAN_TOP_SIDE_LEFT"
	case AV_CHAN_TOP_SIDE_RIGHT:
		return "AV_CHAN_TOP_SIDE_RIGHT"
	case AV_CHAN_BOTTOM_FRONT_CENTER:
		return "AV_CHAN_BOTTOM_FRONT_CENTER"
	case AV_CHAN_BOTTOM_FRONT_LEFT:
		return "AV_CHAN_BOTTOM_FRONT_LEFT"
	case AV_CHAN_BOTTOM_FRONT_RIGHT:
		return "AV_CHAN_BOTTOM_FRONT_RIGHT"
	case AV_CHAN_UNUSED:
		return "AV_CHAN_UNUSED"
	case AV_CHAN_UNKNOWN:
		return "AV_CHAN_UNKNOWN"
	case AV_CHAN_AMBISONIC_BASE:
		return "AV_CHAN_AMBISONIC_BASE"
	case AV_CHAN_AMBISONIC_END:
		return "AV_CHAN_AMBISONIC_END"
	default:
		return "[?? Invalid AVChannel value]"
	}
}
