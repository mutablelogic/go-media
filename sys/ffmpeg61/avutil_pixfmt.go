package ffmpeg

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/pixdesc.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_PIX_FMT_NONE           AVPixelFormat = C.AV_PIX_FMT_NONE
	AV_PIX_FMT_YUV420P        AVPixelFormat = C.AV_PIX_FMT_YUV420P   ///< planar YUV 4:2:0, 12bpp, (1 Cr & Cb sample per 2x2 Y samples)
	AV_PIX_FMT_YUYV422        AVPixelFormat = C.AV_PIX_FMT_YUYV422   ///< packed YUV 4:2:2, 16bpp, Y0 Cb Y1 Cr
	AV_PIX_FMT_RGB24          AVPixelFormat = C.AV_PIX_FMT_RGB24     ///< packed RGB 8:8:8, 24bpp, RGBRGB...
	AV_PIX_FMT_BGR24          AVPixelFormat = C.AV_PIX_FMT_BGR24     ///< packed RGB 8:8:8, 24bpp, BGRBGR...
	AV_PIX_FMT_YUV422P        AVPixelFormat = C.AV_PIX_FMT_YUV422P   ///< planar YUV 4:2:2, 16bpp, (1 Cr & Cb sample per 2x1 Y samples)
	AV_PIX_FMT_YUV444P        AVPixelFormat = C.AV_PIX_FMT_YUV444P   ///< planar YUV 4:4:4, 24bpp, (1 Cr & Cb sample per 1x1 Y samples)
	AV_PIX_FMT_YUV410P        AVPixelFormat = C.AV_PIX_FMT_YUV410P   ///< planar YUV 4:1:0,  9bpp, (1 Cr & Cb sample per 4x4 Y samples)
	AV_PIX_FMT_YUV411P        AVPixelFormat = C.AV_PIX_FMT_YUV411P   ///< planar YUV 4:1:1, 12bpp, (1 Cr & Cb sample per 4x1 Y samples)
	AV_PIX_FMT_GRAY8          AVPixelFormat = C.AV_PIX_FMT_GRAY8     ///<        Y        ,  8bpp
	AV_PIX_FMT_MONOWHITE      AVPixelFormat = C.AV_PIX_FMT_MONOWHITE ///<        Y        ,  1bpp, 0 is white, 1 is black, in each byte pixels are ordered from the msb to the lsb
	AV_PIX_FMT_MONOBLACK      AVPixelFormat = C.AV_PIX_FMT_MONOBLACK ///<        Y        ,  1bpp, 0 is black, 1 is white, in each byte pixels are ordered from the msb to the lsb
	AV_PIX_FMT_PAL8           AVPixelFormat = C.AV_PIX_FMT_PAL8      ///< 8 bits with AV_PIX_FMT_RGB32 palette
	AV_PIX_FMT_YUVJ420P       AVPixelFormat = C.AV_PIX_FMT_YUVJ420P  ///< planar YUV 4:2:0, 12bpp, full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV420P and setting color_range
	AV_PIX_FMT_YUVJ422P       AVPixelFormat = C.AV_PIX_FMT_YUVJ422P  ///< planar YUV 4:2:2, 16bpp, full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV422P and setting color_range
	AV_PIX_FMT_YUVJ444P       AVPixelFormat = C.AV_PIX_FMT_YUVJ444P  ///< planar YUV 4:4:4, 24bpp, full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV444P and setting color_range
	AV_PIX_FMT_UYVY422        AVPixelFormat = C.AV_PIX_FMT_UYVY422   ///< packed YUV 4:2:2, 16bpp, Cb Y0 Cr Y1
	AV_PIX_FMT_UYYVYY411      AVPixelFormat = C.AV_PIX_FMT_UYYVYY411 ///< packed YUV 4:1:1, 12bpp, Cb Y0 Y1 Cr Y2 Y3
	AV_PIX_FMT_BGR8           AVPixelFormat = C.AV_PIX_FMT_BGR8      ///< packed RGB 3:3:2,  8bpp, (msb)2B 3G 3R(lsb)
	AV_PIX_FMT_BGR4           AVPixelFormat = C.AV_PIX_FMT_BGR4      ///< packed RGB 1:2:1 bitstream,  4bpp, (msb)1B 2G 1R(lsb), a byte contains two pixels, the first pixel in the byte is the one composed by the 4 msb bits
	AV_PIX_FMT_BGR4_BYTE      AVPixelFormat = C.AV_PIX_FMT_BGR4_BYTE ///< packed RGB 1:2:1,  8bpp, (msb)1B 2G 1R(lsb)
	AV_PIX_FMT_RGB8           AVPixelFormat = C.AV_PIX_FMT_RGB8      ///< packed RGB 3:3:2,  8bpp, (msb)2R 3G 3B(lsb)
	AV_PIX_FMT_RGB4           AVPixelFormat = C.AV_PIX_FMT_RGB4      ///< packed RGB 1:2:1 bitstream,  4bpp, (msb)1R 2G 1B(lsb), a byte contains two pixels, the first pixel in the byte is the one composed by the 4 msb bits
	AV_PIX_FMT_RGB4_BYTE      AVPixelFormat = C.AV_PIX_FMT_RGB4_BYTE ///< packed RGB 1:2:1,  8bpp, (msb)1R 2G 1B(lsb)
	AV_PIX_FMT_NV12           AVPixelFormat = C.AV_PIX_FMT_NV12      ///< planar YUV 4:2:0, 12bpp, 1 plane for Y and 1 plane for the UV components, which are interleaved (first byte U and the following byte V)
	AV_PIX_FMT_NV21           AVPixelFormat = C.AV_PIX_FMT_NV21      ///< as above, but U and V bytes are swapped
	AV_PIX_FMT_ARGB           AVPixelFormat = C.AV_PIX_FMT_ARGB      ///< packed ARGB 8:8:8:8, 32bpp, ARGBARGB...
	AV_PIX_FMT_RGBA           AVPixelFormat = C.AV_PIX_FMT_RGBA      ///< packed RGBA 8:8:8:8, 32bpp, RGBARGBA...
	AV_PIX_FMT_ABGR           AVPixelFormat = C.AV_PIX_FMT_ABGR      ///< packed ABGR 8:8:8:8, 32bpp, ABGRABGR...
	AV_PIX_FMT_BGRA           AVPixelFormat = C.AV_PIX_FMT_BGRA      ///< packed BGRA 8:8:8:8, 32bpp, BGRABGRA...
	AV_PIX_FMT_GRAY16BE       AVPixelFormat = C.AV_PIX_FMT_GRAY16BE  ///<        Y        , 16bpp, big-endian
	AV_PIX_FMT_GRAY16LE       AVPixelFormat = C.AV_PIX_FMT_GRAY16LE  ///<        Y        , 16bpp, little-endian
	AV_PIX_FMT_YUV440P        AVPixelFormat = C.AV_PIX_FMT_YUV440P   ///< planar YUV 4:4:0 (1 Cr & Cb sample per 1x2 Y samples)
	AV_PIX_FMT_YUVJ440P       AVPixelFormat = C.AV_PIX_FMT_YUVJ440P  ///< planar YUV 4:4:0 full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV440P and setting color_range
	AV_PIX_FMT_YUVA420P       AVPixelFormat = C.AV_PIX_FMT_YUVA420P  ///< planar YUV 4:2:0, 20bpp, (1 Cr & Cb sample per 2x2 Y & A samples)
	AV_PIX_FMT_RGB48BE        AVPixelFormat = C.AV_PIX_FMT_RGB48BE   ///< packed RGB 16:16:16, 48bpp, 16R, 16G, 16B, the 2-byte value for each R/G/B component is stored as big-endian
	AV_PIX_FMT_RGB48LE        AVPixelFormat = C.AV_PIX_FMT_RGB48LE   ///< packed RGB 16:16:16, 48bpp, 16R, 16G, 16B, the 2-byte value for each R/G/B component is stored as little-endian
	AV_PIX_FMT_RGB565BE       AVPixelFormat = C.AV_PIX_FMT_RGB565BE  ///< packed RGB 5:6:5, 16bpp, (msb)   5R 6G 5B(lsb), big-endian
	AV_PIX_FMT_RGB565LE       AVPixelFormat = C.AV_PIX_FMT_RGB565LE  ///< packed RGB 5:6:5, 16bpp, (msb)   5R 6G 5B(lsb), little-endian
	AV_PIX_FMT_RGB555BE       AVPixelFormat = C.AV_PIX_FMT_RGB555BE  ///< packed RGB 5:5:5, 16bpp, (msb)1X 5R 5G 5B(lsb), big-endian   , X=unused/undefined
	AV_PIX_FMT_RGB555LE       AVPixelFormat = C.AV_PIX_FMT_RGB555LE  ///< packed RGB 5:5:5, 16bpp, (msb)1X 5R 5G 5B(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_BGR565BE       AVPixelFormat = C.AV_PIX_FMT_BGR565BE  ///< packed BGR 5:6:5, 16bpp, (msb)   5B 6G 5R(lsb), big-endian
	AV_PIX_FMT_BGR565LE       AVPixelFormat = C.AV_PIX_FMT_BGR565LE  ///< packed BGR 5:6:5, 16bpp, (msb)   5B 6G 5R(lsb), little-endian
	AV_PIX_FMT_BGR555BE       AVPixelFormat = C.AV_PIX_FMT_BGR555BE  ///< packed BGR 5:5:5, 16bpp, (msb)1X 5B 5G 5R(lsb), big-endian   , X=unused/undefined
	AV_PIX_FMT_BGR555LE       AVPixelFormat = C.AV_PIX_FMT_BGR555LE  ///< packed BGR 5:5:5, 16bpp, (msb)1X 5B 5G 5R(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_VAAPI          AVPixelFormat = C.AV_PIX_FMT_VAAPI
	AV_PIX_FMT_YUV420P16LE    AVPixelFormat = C.AV_PIX_FMT_YUV420P16LE  ///< planar YUV 4:2:0, 24bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV420P16BE    AVPixelFormat = C.AV_PIX_FMT_YUV420P16BE  ///< planar YUV 4:2:0, 24bpp, (1 Cr & Cb sample per 2x2 Y samples), big-endian
	AV_PIX_FMT_YUV422P16LE    AVPixelFormat = C.AV_PIX_FMT_YUV422P16LE  ///< planar YUV 4:2:2, 32bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_YUV422P16BE    AVPixelFormat = C.AV_PIX_FMT_YUV422P16BE  ///< planar YUV 4:2:2, 32bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P16LE    AVPixelFormat = C.AV_PIX_FMT_YUV444P16LE  ///< planar YUV 4:4:4, 48bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P16BE    AVPixelFormat = C.AV_PIX_FMT_YUV444P16BE  ///< planar YUV 4:4:4, 48bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_DXVA2_VLD      AVPixelFormat = C.AV_PIX_FMT_DXVA2_VLD    ///< HW decoding through DXVA2, Picture.data[3] contains a LPDIRECT3DSURFACE9 pointer
	AV_PIX_FMT_RGB444LE       AVPixelFormat = C.AV_PIX_FMT_RGB444LE     ///< packed RGB 4:4:4, 16bpp, (msb)4X 4R 4G 4B(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_RGB444BE       AVPixelFormat = C.AV_PIX_FMT_RGB444BE     ///< packed RGB 4:4:4, 16bpp, (msb)4X 4R 4G 4B(lsb), big-endian,    X=unused/undefined
	AV_PIX_FMT_BGR444LE       AVPixelFormat = C.AV_PIX_FMT_BGR444LE     ///< packed BGR 4:4:4, 16bpp, (msb)4X 4B 4G 4R(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_BGR444BE       AVPixelFormat = C.AV_PIX_FMT_BGR444BE     ///< packed BGR 4:4:4, 16bpp, (msb)4X 4B 4G 4R(lsb), big-endian,    X=unused/undefined
	AV_PIX_FMT_YA8            AVPixelFormat = C.AV_PIX_FMT_YA8          ///< 8 bits gray, 8 bits alpha
	AV_PIX_FMT_Y400A          AVPixelFormat = C.AV_PIX_FMT_Y400A        ///< alias for AV_PIX_FMT_YA8
	AV_PIX_FMT_GRAY8A         AVPixelFormat = C.AV_PIX_FMT_GRAY8A       ///< alias for AV_PIX_FMT_YA8
	AV_PIX_FMT_BGR48BE        AVPixelFormat = C.AV_PIX_FMT_BGR48BE      ///< packed RGB 16:16:16, 48bpp, 16B, 16G, 16R, the 2-byte value for each R/G/B component is stored as big-endian
	AV_PIX_FMT_BGR48LE        AVPixelFormat = C.AV_PIX_FMT_BGR48LE      ///< packed RGB 16:16:16, 48bpp, 16B, 16G, 16R, the 2-byte value for each R/G/B component is stored as little-endian
	AV_PIX_FMT_YUV420P9BE     AVPixelFormat = C.AV_PIX_FMT_YUV420P9BE   ///< planar YUV 4:2:0, 13.5bpp, (1 Cr & Cb sample per 2x2 Y samples), big-endian
	AV_PIX_FMT_YUV420P9LE     AVPixelFormat = C.AV_PIX_FMT_YUV420P9LE   ///< planar YUV 4:2:0, 13.5bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV420P10BE    AVPixelFormat = C.AV_PIX_FMT_YUV420P10BE  ///< planar YUV 4:2:0, 15bpp, (1 Cr & Cb sample per 2x2 Y samples), big-endian
	AV_PIX_FMT_YUV420P10LE    AVPixelFormat = C.AV_PIX_FMT_YUV420P10LE  ///< planar YUV 4:2:0, 15bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV422P10BE    AVPixelFormat = C.AV_PIX_FMT_YUV422P10BE  ///< planar YUV 4:2:2, 20bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV422P10LE    AVPixelFormat = C.AV_PIX_FMT_YUV422P10LE  ///< planar YUV 4:2:2, 20bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P9BE     AVPixelFormat = C.AV_PIX_FMT_YUV444P9BE   ///< planar YUV 4:4:4, 27bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P9LE     AVPixelFormat = C.AV_PIX_FMT_YUV444P9LE   ///< planar YUV 4:4:4, 27bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P10BE    AVPixelFormat = C.AV_PIX_FMT_YUV444P10BE  ///< planar YUV 4:4:4, 30bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P10LE    AVPixelFormat = C.AV_PIX_FMT_YUV444P10LE  ///< planar YUV 4:4:4, 30bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_YUV422P9BE     AVPixelFormat = C.AV_PIX_FMT_YUV422P9BE   ///< planar YUV 4:2:2, 18bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV422P9LE     AVPixelFormat = C.AV_PIX_FMT_YUV422P9LE   ///< planar YUV 4:2:2, 18bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_GBRP           AVPixelFormat = C.AV_PIX_FMT_GBRP         ///< planar GBR 4:4:4 24bpp
	AV_PIX_FMT_GBR24P         AVPixelFormat = C.AV_PIX_FMT_GBR24P       // alias for #AV_PIX_FMT_GBRP
	AV_PIX_FMT_GBRP9BE        AVPixelFormat = C.AV_PIX_FMT_GBRP9BE      ///< planar GBR 4:4:4 27bpp, big-endian
	AV_PIX_FMT_GBRP9LE        AVPixelFormat = C.AV_PIX_FMT_GBRP9LE      ///< planar GBR 4:4:4 27bpp, little-endian
	AV_PIX_FMT_GBRP10BE       AVPixelFormat = C.AV_PIX_FMT_GBRP10BE     ///< planar GBR 4:4:4 30bpp, big-endian
	AV_PIX_FMT_GBRP10LE       AVPixelFormat = C.AV_PIX_FMT_GBRP10LE     ///< planar GBR 4:4:4 30bpp, little-endian
	AV_PIX_FMT_GBRP16BE       AVPixelFormat = C.AV_PIX_FMT_GBRP16BE     ///< planar GBR 4:4:4 48bpp, big-endian
	AV_PIX_FMT_GBRP16LE       AVPixelFormat = C.AV_PIX_FMT_GBRP16LE     ///< planar GBR 4:4:4 48bpp, little-endian
	AV_PIX_FMT_YUVA422P       AVPixelFormat = C.AV_PIX_FMT_YUVA422P     ///< planar YUV 4:2:2 24bpp, (1 Cr & Cb sample per 2x1 Y & A samples)
	AV_PIX_FMT_YUVA444P       AVPixelFormat = C.AV_PIX_FMT_YUVA444P     ///< planar YUV 4:4:4 32bpp, (1 Cr & Cb sample per 1x1 Y & A samples)
	AV_PIX_FMT_YUVA420P9BE    AVPixelFormat = C.AV_PIX_FMT_YUVA420P9BE  ///< planar YUV 4:2:0 22.5bpp, (1 Cr & Cb sample per 2x2 Y & A samples), big-endian
	AV_PIX_FMT_YUVA420P9LE    AVPixelFormat = C.AV_PIX_FMT_YUVA420P9LE  ///< planar YUV 4:2:0 22.5bpp, (1 Cr & Cb sample per 2x2 Y & A samples), little-endian
	AV_PIX_FMT_YUVA422P9BE    AVPixelFormat = C.AV_PIX_FMT_YUVA422P9BE  ///< planar YUV 4:2:2 27bpp, (1 Cr & Cb sample per 2x1 Y & A samples), big-endian
	AV_PIX_FMT_YUVA422P9LE    AVPixelFormat = C.AV_PIX_FMT_YUVA422P9LE  ///< planar YUV 4:2:2 27bpp, (1 Cr & Cb sample per 2x1 Y & A samples), little-endian
	AV_PIX_FMT_YUVA444P9BE    AVPixelFormat = C.AV_PIX_FMT_YUVA444P9BE  ///< planar YUV 4:4:4 36bpp, (1 Cr & Cb sample per 1x1 Y & A samples), big-endian
	AV_PIX_FMT_YUVA444P9LE    AVPixelFormat = C.AV_PIX_FMT_YUVA444P9LE  ///< planar YUV 4:4:4 36bpp, (1 Cr & Cb sample per 1x1 Y & A samples), little-endian
	AV_PIX_FMT_YUVA420P10BE   AVPixelFormat = C.AV_PIX_FMT_YUVA420P10BE ///< planar YUV 4:2:0 25bpp, (1 Cr & Cb sample per 2x2 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA420P10LE   AVPixelFormat = C.AV_PIX_FMT_YUVA420P10LE ///< planar YUV 4:2:0 25bpp, (1 Cr & Cb sample per 2x2 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA422P10BE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P10BE ///< planar YUV 4:2:2 30bpp, (1 Cr & Cb sample per 2x1 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA422P10LE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P10LE ///< planar YUV 4:2:2 30bpp, (1 Cr & Cb sample per 2x1 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA444P10BE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P10BE ///< planar YUV 4:4:4 40bpp, (1 Cr & Cb sample per 1x1 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA444P10LE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P10LE ///< planar YUV 4:4:4 40bpp, (1 Cr & Cb sample per 1x1 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA420P16BE   AVPixelFormat = C.AV_PIX_FMT_YUVA420P16BE ///< planar YUV 4:2:0 40bpp, (1 Cr & Cb sample per 2x2 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA420P16LE   AVPixelFormat = C.AV_PIX_FMT_YUVA420P16LE ///< planar YUV 4:2:0 40bpp, (1 Cr & Cb sample per 2x2 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA422P16BE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P16BE ///< planar YUV 4:2:2 48bpp, (1 Cr & Cb sample per 2x1 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA422P16LE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P16LE ///< planar YUV 4:2:2 48bpp, (1 Cr & Cb sample per 2x1 Y & A samples, little-endian)
	AV_PIX_FMT_YUVA444P16BE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P16BE ///< planar YUV 4:4:4 64bpp, (1 Cr & Cb sample per 1x1 Y & A samples, big-endian)
	AV_PIX_FMT_YUVA444P16LE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P16LE ///< planar YUV 4:4:4 64bpp, (1 Cr & Cb sample per 1x1 Y & A samples, little-endian)
	AV_PIX_FMT_VDPAU          AVPixelFormat = C.AV_PIX_FMT_VDPAU        ///< HW acceleration through VDPAU, Picture.data[3] contains a VdpVideoSurface
	AV_PIX_FMT_XYZ12LE        AVPixelFormat = C.AV_PIX_FMT_XYZ12LE      ///< packed XYZ 4:4:4, 36 bpp, (msb) 12X, 12Y, 12Z (lsb), the 2-byte value for each X/Y/Z is stored as little-endian, the 4 lower bits are set to 0
	AV_PIX_FMT_XYZ12BE        AVPixelFormat = C.AV_PIX_FMT_XYZ12BE      ///< packed XYZ 4:4:4, 36 bpp, (msb) 12X, 12Y, 12Z (lsb), the 2-byte value for each X/Y/Z is stored as big-endian, the 4 lower bits are set to 0
	AV_PIX_FMT_NV16           AVPixelFormat = C.AV_PIX_FMT_NV16         ///< interleaved chroma YUV 4:2:2, 16bpp, (1 Cr & Cb sample per 2x1 Y samples)
	AV_PIX_FMT_NV20LE         AVPixelFormat = C.AV_PIX_FMT_NV20LE       ///< interleaved chroma YUV 4:2:2, 20bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_NV20BE         AVPixelFormat = C.AV_PIX_FMT_NV20BE       ///< interleaved chroma YUV 4:2:2, 20bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_RGBA64BE       AVPixelFormat = C.AV_PIX_FMT_RGBA64BE     ///< packed RGBA 16:16:16:16, 64bpp, 16R, 16G, 16B, 16A, the 2-byte value for each R/G/B/A component is stored as big-endian
	AV_PIX_FMT_RGBA64LE       AVPixelFormat = C.AV_PIX_FMT_RGBA64LE     ///< packed RGBA 16:16:16:16, 64bpp, 16R, 16G, 16B, 16A, the 2-byte value for each R/G/B/A component is stored as little-endian
	AV_PIX_FMT_BGRA64BE       AVPixelFormat = C.AV_PIX_FMT_BGRA64BE     ///< packed RGBA 16:16:16:16, 64bpp, 16B, 16G, 16R, 16A, the 2-byte value for each R/G/B/A component is stored as big-endian
	AV_PIX_FMT_BGRA64LE       AVPixelFormat = C.AV_PIX_FMT_BGRA64LE     ///< packed RGBA 16:16:16:16, 64bpp, 16B, 16G, 16R, 16A, the 2-byte value for each R/G/B/A component is stored as little-endian
	AV_PIX_FMT_YVYU422        AVPixelFormat = C.AV_PIX_FMT_YVYU422      ///< packed YUV 4:2:2, 16bpp, Y0 Cr Y1 Cb
	AV_PIX_FMT_YA16BE         AVPixelFormat = C.AV_PIX_FMT_YA16BE       ///< 16 bits gray, 16 bits alpha (big-endian)
	AV_PIX_FMT_YA16LE         AVPixelFormat = C.AV_PIX_FMT_YA16LE       ///< 16 bits gray, 16 bits alpha (little-endian)
	AV_PIX_FMT_GBRAP          AVPixelFormat = C.AV_PIX_FMT_GBRAP        ///< planar GBRA 4:4:4:4 32bpp
	AV_PIX_FMT_GBRAP16BE      AVPixelFormat = C.AV_PIX_FMT_GBRAP16BE    ///< planar GBRA 4:4:4:4 64bpp, big-endian
	AV_PIX_FMT_GBRAP16LE      AVPixelFormat = C.AV_PIX_FMT_GBRAP16LE    ///< planar GBRA 4:4:4:4 64bpp, little-endian
	AV_PIX_FMT_QSV            AVPixelFormat = C.AV_PIX_FMT_QSV
	AV_PIX_FMT_MMAL           AVPixelFormat = C.AV_PIX_FMT_MMAL
	AV_PIX_FMT_D3D11VA_VLD    AVPixelFormat = C.AV_PIX_FMT_D3D11VA_VLD ///< HW decoding through Direct3D11 via old API, Picture.data[3] contains a ID3D11VideoDecoderOutputView pointer
	AV_PIX_FMT_CUDA           AVPixelFormat = C.AV_PIX_FMT_CUDA
	AV_PIX_FMT_0RGB           AVPixelFormat = C.AV_PIX_FMT_0RGB           ///< packed RGB 8:8:8, 32bpp, XRGBXRGB...   X=unused/undefined
	AV_PIX_FMT_RGB0           AVPixelFormat = C.AV_PIX_FMT_RGB0           ///< packed RGB 8:8:8, 32bpp, RGBXRGBX...   X=unused/undefined
	AV_PIX_FMT_0BGR           AVPixelFormat = C.AV_PIX_FMT_0BGR           ///< packed BGR 8:8:8, 32bpp, XBGRXBGR...   X=unused/undefined
	AV_PIX_FMT_BGR0           AVPixelFormat = C.AV_PIX_FMT_BGR0           ///< packed BGR 8:8:8, 32bpp, BGRXBGRX...   X=unused/undefined
	AV_PIX_FMT_YUV420P12BE    AVPixelFormat = C.AV_PIX_FMT_YUV420P12BE    ///< planar YUV 4:2:0,18bpp, (1 Cr & Cb sample per 2x2 Y samples), big-endian
	AV_PIX_FMT_YUV420P12LE    AVPixelFormat = C.AV_PIX_FMT_YUV420P12LE    ///< planar YUV 4:2:0,18bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV420P14BE    AVPixelFormat = C.AV_PIX_FMT_YUV420P14BE    ///< planar YUV 4:2:0,21bpp, (1 Cr & Cb sample per 2x2 Y samples), big-endian
	AV_PIX_FMT_YUV420P14LE    AVPixelFormat = C.AV_PIX_FMT_YUV420P14LE    ///< planar YUV 4:2:0,21bpp, (1 Cr & Cb sample per 2x2 Y samples), little-endian
	AV_PIX_FMT_YUV422P12BE    AVPixelFormat = C.AV_PIX_FMT_YUV422P12BE    ///< planar YUV 4:2:2,24bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV422P12LE    AVPixelFormat = C.AV_PIX_FMT_YUV422P12LE    ///< planar YUV 4:2:2,24bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_YUV422P14BE    AVPixelFormat = C.AV_PIX_FMT_YUV422P14BE    ///< planar YUV 4:2:2,28bpp, (1 Cr & Cb sample per 2x1 Y samples), big-endian
	AV_PIX_FMT_YUV422P14LE    AVPixelFormat = C.AV_PIX_FMT_YUV422P14LE    ///< planar YUV 4:2:2,28bpp, (1 Cr & Cb sample per 2x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P12BE    AVPixelFormat = C.AV_PIX_FMT_YUV444P12BE    ///< planar YUV 4:4:4,36bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P12LE    AVPixelFormat = C.AV_PIX_FMT_YUV444P12LE    ///< planar YUV 4:4:4,36bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_YUV444P14BE    AVPixelFormat = C.AV_PIX_FMT_YUV444P14BE    ///< planar YUV 4:4:4,42bpp, (1 Cr & Cb sample per 1x1 Y samples), big-endian
	AV_PIX_FMT_YUV444P14LE    AVPixelFormat = C.AV_PIX_FMT_YUV444P14LE    ///< planar YUV 4:4:4,42bpp, (1 Cr & Cb sample per 1x1 Y samples), little-endian
	AV_PIX_FMT_GBRP12BE       AVPixelFormat = C.AV_PIX_FMT_GBRP12BE       ///< planar GBR 4:4:4 36bpp, big-endian
	AV_PIX_FMT_GBRP12LE       AVPixelFormat = C.AV_PIX_FMT_GBRP12LE       ///< planar GBR 4:4:4 36bpp, little-endian
	AV_PIX_FMT_GBRP14BE       AVPixelFormat = C.AV_PIX_FMT_GBRP14BE       ///< planar GBR 4:4:4 42bpp, big-endian
	AV_PIX_FMT_GBRP14LE       AVPixelFormat = C.AV_PIX_FMT_GBRP14LE       ///< planar GBR 4:4:4 42bpp, little-endian
	AV_PIX_FMT_YUVJ411P       AVPixelFormat = C.AV_PIX_FMT_YUVJ411P       ///< planar YUV 4:1:1, 12bpp, (1 Cr & Cb sample per 4x1 Y samples) full scale (JPEG), deprecated in favor of AV_PIX_FMT_YUV411P and setting color_range
	AV_PIX_FMT_BAYER_BGGR8    AVPixelFormat = C.AV_PIX_FMT_BAYER_BGGR8    ///< bayer, BGBG..(odd line), GRGR..(even line), 8-bit samples
	AV_PIX_FMT_BAYER_RGGB8    AVPixelFormat = C.AV_PIX_FMT_BAYER_RGGB8    ///< bayer, RGRG..(odd line), GBGB..(even line), 8-bit samples
	AV_PIX_FMT_BAYER_GBRG8    AVPixelFormat = C.AV_PIX_FMT_BAYER_GBRG8    ///< bayer, GBGB..(odd line), RGRG..(even line), 8-bit samples
	AV_PIX_FMT_BAYER_GRBG8    AVPixelFormat = C.AV_PIX_FMT_BAYER_GRBG8    ///< bayer, GRGR..(odd line), BGBG..(even line), 8-bit samples
	AV_PIX_FMT_BAYER_BGGR16LE AVPixelFormat = C.AV_PIX_FMT_BAYER_BGGR16LE ///< bayer, BGBG..(odd line), GRGR..(even line), 16-bit samples, little-endian
	AV_PIX_FMT_BAYER_BGGR16BE AVPixelFormat = C.AV_PIX_FMT_BAYER_BGGR16BE ///< bayer, BGBG..(odd line), GRGR..(even line), 16-bit samples, big-endian
	AV_PIX_FMT_BAYER_RGGB16LE AVPixelFormat = C.AV_PIX_FMT_BAYER_RGGB16LE ///< bayer, RGRG..(odd line), GBGB..(even line), 16-bit samples, little-endian
	AV_PIX_FMT_BAYER_RGGB16BE AVPixelFormat = C.AV_PIX_FMT_BAYER_RGGB16BE ///< bayer, RGRG..(odd line), GBGB..(even line), 16-bit samples, big-endian
	AV_PIX_FMT_BAYER_GBRG16LE AVPixelFormat = C.AV_PIX_FMT_BAYER_GBRG16LE ///< bayer, GBGB..(odd line), RGRG..(even line), 16-bit samples, little-endian
	AV_PIX_FMT_BAYER_GBRG16BE AVPixelFormat = C.AV_PIX_FMT_BAYER_GBRG16BE ///< bayer, GBGB..(odd line), RGRG..(even line), 16-bit samples, big-endian
	AV_PIX_FMT_BAYER_GRBG16LE AVPixelFormat = C.AV_PIX_FMT_BAYER_GRBG16LE ///< bayer, GRGR..(odd line), BGBG..(even line), 16-bit samples, little-endian
	AV_PIX_FMT_BAYER_GRBG16BE AVPixelFormat = C.AV_PIX_FMT_BAYER_GRBG16BE ///< bayer, GRGR..(odd line), BGBG..(even line), 16-bit samples, big-endian
	AV_PIX_FMT_YUV440P10LE    AVPixelFormat = C.AV_PIX_FMT_YUV440P10LE    ///< planar YUV 4:4:0,20bpp, (1 Cr & Cb sample per 1x2 Y samples), little-endian
	AV_PIX_FMT_YUV440P10BE    AVPixelFormat = C.AV_PIX_FMT_YUV440P10BE    ///< planar YUV 4:4:0,20bpp, (1 Cr & Cb sample per 1x2 Y samples), big-endian
	AV_PIX_FMT_YUV440P12LE    AVPixelFormat = C.AV_PIX_FMT_YUV440P12LE    ///< planar YUV 4:4:0,24bpp, (1 Cr & Cb sample per 1x2 Y samples), little-endian
	AV_PIX_FMT_YUV440P12BE    AVPixelFormat = C.AV_PIX_FMT_YUV440P12BE    ///< planar YUV 4:4:0,24bpp, (1 Cr & Cb sample per 1x2 Y samples), big-endian
	AV_PIX_FMT_AYUV64LE       AVPixelFormat = C.AV_PIX_FMT_AYUV64LE       ///< packed AYUV 4:4:4,64bpp (1 Cr & Cb sample per 1x1 Y & A samples), little-endian
	AV_PIX_FMT_AYUV64BE       AVPixelFormat = C.AV_PIX_FMT_AYUV64BE       ///< packed AYUV 4:4:4,64bpp (1 Cr & Cb sample per 1x1 Y & A samples), big-endian
	AV_PIX_FMT_VIDEOTOOLBOX   AVPixelFormat = C.AV_PIX_FMT_VIDEOTOOLBOX   ///< hardware decoding through Videotoolbox
	AV_PIX_FMT_P010LE         AVPixelFormat = C.AV_PIX_FMT_P010LE         ///< like NV12, with 10bpp per component, data in the high bits, zeros in the low bits, little-endian
	AV_PIX_FMT_P010BE         AVPixelFormat = C.AV_PIX_FMT_P010BE         ///< like NV12, with 10bpp per component, data in the high bits, zeros in the low bits, big-endian
	AV_PIX_FMT_GBRAP12BE      AVPixelFormat = C.AV_PIX_FMT_GBRAP12BE      ///< planar GBR 4:4:4:4 48bpp, big-endian
	AV_PIX_FMT_GBRAP12LE      AVPixelFormat = C.AV_PIX_FMT_GBRAP12LE      ///< planar GBR 4:4:4:4 48bpp, little-endian
	AV_PIX_FMT_GBRAP10BE      AVPixelFormat = C.AV_PIX_FMT_GBRAP10BE      ///< planar GBR 4:4:4:4 40bpp, big-endian
	AV_PIX_FMT_GBRAP10LE      AVPixelFormat = C.AV_PIX_FMT_GBRAP10LE      ///< planar GBR 4:4:4:4 40bpp, little-endian
	AV_PIX_FMT_MEDIACODEC     AVPixelFormat = C.AV_PIX_FMT_MEDIACODEC     ///< hardware decoding through MediaCodec
	AV_PIX_FMT_GRAY12BE       AVPixelFormat = C.AV_PIX_FMT_GRAY12BE       ///<        Y        , 12bpp, big-endian
	AV_PIX_FMT_GRAY12LE       AVPixelFormat = C.AV_PIX_FMT_GRAY12LE       ///<        Y        , 12bpp, little-endian
	AV_PIX_FMT_GRAY10BE       AVPixelFormat = C.AV_PIX_FMT_GRAY10BE       ///<        Y        , 10bpp, big-endian
	AV_PIX_FMT_GRAY10LE       AVPixelFormat = C.AV_PIX_FMT_GRAY10LE       ///<        Y        , 10bpp, little-endian
	AV_PIX_FMT_P016LE         AVPixelFormat = C.AV_PIX_FMT_P016LE         ///< like NV12, with 16bpp per component, little-endian
	AV_PIX_FMT_P016BE         AVPixelFormat = C.AV_PIX_FMT_P016BE         ///< like NV12, with 16bpp per component, big-endian
	AV_PIX_FMT_D3D11          AVPixelFormat = C.AV_PIX_FMT_D3D11
	AV_PIX_FMT_GRAY9BE        AVPixelFormat = C.AV_PIX_FMT_GRAY9BE    ///<        Y        , 9bpp, big-endian
	AV_PIX_FMT_GRAY9LE        AVPixelFormat = C.AV_PIX_FMT_GRAY9LE    ///<        Y        , 9bpp, little-endian
	AV_PIX_FMT_GBRPF32BE      AVPixelFormat = C.AV_PIX_FMT_GBRPF32BE  ///< IEEE-754 single precision planar GBR 4:4:4,     96bpp, big-endian
	AV_PIX_FMT_GBRPF32LE      AVPixelFormat = C.AV_PIX_FMT_GBRPF32LE  ///< IEEE-754 single precision planar GBR 4:4:4,     96bpp, little-endian
	AV_PIX_FMT_GBRAPF32BE     AVPixelFormat = C.AV_PIX_FMT_GBRAPF32BE ///< IEEE-754 single precision planar GBRA 4:4:4:4, 128bpp, big-endian
	AV_PIX_FMT_GBRAPF32LE     AVPixelFormat = C.AV_PIX_FMT_GBRAPF32LE ///< IEEE-754 single precision planar GBRA 4:4:4:4, 128bpp, little-endian
	AV_PIX_FMT_DRM_PRIME      AVPixelFormat = C.AV_PIX_FMT_DRM_PRIME
	AV_PIX_FMT_OPENCL         AVPixelFormat = C.AV_PIX_FMT_OPENCL
	AV_PIX_FMT_GRAY14BE       AVPixelFormat = C.AV_PIX_FMT_GRAY14BE     ///<        Y        , 14bpp, big-endian
	AV_PIX_FMT_GRAY14LE       AVPixelFormat = C.AV_PIX_FMT_GRAY14LE     ///<        Y        , 14bpp, little-endian
	AV_PIX_FMT_GRAYF32BE      AVPixelFormat = C.AV_PIX_FMT_GRAYF32BE    ///< IEEE-754 single precision Y, 32bpp, big-endian
	AV_PIX_FMT_GRAYF32LE      AVPixelFormat = C.AV_PIX_FMT_GRAYF32LE    ///< IEEE-754 single precision Y, 32bpp, little-endian
	AV_PIX_FMT_YUVA422P12BE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P12BE ///< planar YUV 4:2:2,24bpp, (1 Cr & Cb sample per 2x1 Y samples), 12b alpha, big-endian
	AV_PIX_FMT_YUVA422P12LE   AVPixelFormat = C.AV_PIX_FMT_YUVA422P12LE ///< planar YUV 4:2:2,24bpp, (1 Cr & Cb sample per 2x1 Y samples), 12b alpha, little-endian
	AV_PIX_FMT_YUVA444P12BE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P12BE ///< planar YUV 4:4:4,36bpp, (1 Cr & Cb sample per 1x1 Y samples), 12b alpha, big-endian
	AV_PIX_FMT_YUVA444P12LE   AVPixelFormat = C.AV_PIX_FMT_YUVA444P12LE ///< planar YUV 4:4:4,36bpp, (1 Cr & Cb sample per 1x1 Y samples), 12b alpha, little-endian
	AV_PIX_FMT_NV24           AVPixelFormat = C.AV_PIX_FMT_NV24         ///< planar YUV 4:4:4, 24bpp, 1 plane for Y and 1 plane for the UV components, which are interleaved (first byte U and the following byte V)
	AV_PIX_FMT_NV42           AVPixelFormat = C.AV_PIX_FMT_NV42         ///< as above, but U and V bytes are swapped
	AV_PIX_FMT_VULKAN         AVPixelFormat = C.AV_PIX_FMT_VULKAN
	AV_PIX_FMT_Y210BE         AVPixelFormat = C.AV_PIX_FMT_Y210BE    ///< packed YUV 4:2:2 like YUYV422, 20bpp, data in the high bits, big-endian
	AV_PIX_FMT_Y210LE         AVPixelFormat = C.AV_PIX_FMT_Y210LE    ///< packed YUV 4:2:2 like YUYV422, 20bpp, data in the high bits, little-endian
	AV_PIX_FMT_X2RGB10LE      AVPixelFormat = C.AV_PIX_FMT_X2RGB10LE ///< packed RGB 10:10:10, 30bpp, (msb)2X 10R 10G 10B(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_X2RGB10BE      AVPixelFormat = C.AV_PIX_FMT_X2RGB10BE ///< packed RGB 10:10:10, 30bpp, (msb)2X 10R 10G 10B(lsb), big-endian, X=unused/undefined
	AV_PIX_FMT_X2BGR10LE      AVPixelFormat = C.AV_PIX_FMT_X2BGR10LE ///< packed BGR 10:10:10, 30bpp, (msb)2X 10B 10G 10R(lsb), little-endian, X=unused/undefined
	AV_PIX_FMT_X2BGR10BE      AVPixelFormat = C.AV_PIX_FMT_X2BGR10BE ///< packed BGR 10:10:10, 30bpp, (msb)2X 10B 10G 10R(lsb), big-endian, X=unused/undefined
	AV_PIX_FMT_P210BE         AVPixelFormat = C.AV_PIX_FMT_P210BE    ///< interleaved chroma YUV 4:2:2, 20bpp, data in the high bits, big-endian
	AV_PIX_FMT_P210LE         AVPixelFormat = C.AV_PIX_FMT_P210LE    ///< interleaved chroma YUV 4:2:2, 20bpp, data in the high bits, little-endian
	AV_PIX_FMT_P410BE         AVPixelFormat = C.AV_PIX_FMT_P410BE    ///< interleaved chroma YUV 4:4:4, 30bpp, data in the high bits, big-endian
	AV_PIX_FMT_P410LE         AVPixelFormat = C.AV_PIX_FMT_P410LE    ///< interleaved chroma YUV 4:4:4, 30bpp, data in the high bits, little-endian
	AV_PIX_FMT_P216BE         AVPixelFormat = C.AV_PIX_FMT_P216BE    ///< interleaved chroma YUV 4:2:2, 32bpp, big-endian
	AV_PIX_FMT_P216LE         AVPixelFormat = C.AV_PIX_FMT_P216LE    ///< interleaved chroma YUV 4:2:2, 32bpp, little-endian
	AV_PIX_FMT_P416BE         AVPixelFormat = C.AV_PIX_FMT_P416BE    ///< interleaved chroma YUV 4:4:4, 48bpp, big-endian
	AV_PIX_FMT_P416LE         AVPixelFormat = C.AV_PIX_FMT_P416LE    ///< interleaved chroma YUV 4:4:4, 48bpp, little-endian
	//AV_PIX_FMT_VUYA           AVPixelFormat = C.AV_PIX_FMT_VUYA      ///< packed VUYA 4:4:4, 32bpp, VUYAVUYA...
	//AV_PIX_FMT_RGBAF16BE      AVPixelFormat = C.AV_PIX_FMT_RGBAF16BE ///< IEEE-754 half precision packed RGBA 16:16:16:16, 64bpp, RGBARGBA..., big-endian
	//AV_PIX_FMT_RGBAF16LE      AVPixelFormat = C.AV_PIX_FMT_RGBAF16LE ///< IEEE-754 half precision packed RGBA 16:16:16:16, 64bpp, RGBARGBA..., little-endian
	//AV_PIX_FMT_VUYX AVPixelFormat = C.AV_PIX_FMT_VUYX ///< packed VUYX 4:4:4, 32bpp, Variant of VUYA where alpha channel is left undefined
	//AV_PIX_FMT_P012LE         AVPixelFormat = C.AV_PIX_FMT_P012LE    ///< like NV12, with 12bpp per component, data in the high bits, zeros in the low bits, little-endian
	//AV_PIX_FMT_P012BE         AVPixelFormat = C.AV_PIX_FMT_P012BE    ///< like NV12, with 12bpp per component, data in the high bits, zeros in the low bits, big-endian
	//AV_PIX_FMT_Y212BE AVPixelFormat = C.AV_PIX_FMT_Y212BE ///< packed YUV 4:2:2 like YUYV422, 24bpp, data in the high bits, zeros in the low bits, big-endian
	//AV_PIX_FMT_Y212LE AVPixelFormat = C.AV_PIX_FMT_Y212LE ///< packed YUV 4:2:2 like YUYV422, 24bpp, data in the high bits, zeros in the low bits, little-endian
	//AV_PIX_FMT_XV30BE AVPixelFormat = C.AV_PIX_FMT_XV30BE ///< packed XVYU 4:4:4, 32bpp, (msb)2X 10V 10Y 10U(lsb), big-endian, variant of Y410 where alpha channel is left undefined
	//AV_PIX_FMT_XV30LE AVPixelFormat = C.AV_PIX_FMT_XV30LE ///< packed XVYU 4:4:4, 32bpp, (msb)2X 10V 10Y 10U(lsb), little-endian, variant of Y410 where alpha channel is left undefined
	//AV_PIX_FMT_XV36BE AVPixelFormat = C.AV_PIX_FMT_XV36BE ///< packed XVYU 4:4:4, 48bpp, data in the high bits, zeros in the low bits, big-endian, variant of Y412 where alpha channel is left undefined
	//AV_PIX_FMT_XV36LE AVPixelFormat = C.AV_PIX_FMT_XV36LE ///< packed XVYU 4:4:4, 48bpp, data in the high bits, zeros in the low bits, little-endian, variant of Y412 where alpha channel is left undefined
	//AV_PIX_FMT_RGBF32BE  AVPixelFormat = C.AV_PIX_FMT_RGBF32BE  ///< IEEE-754 single precision packed RGB 32:32:32, 96bpp, RGBRGB..., big-endian
	//AV_PIX_FMT_RGBF32LE  AVPixelFormat = C.AV_PIX_FMT_RGBF32LE  ///< IEEE-754 single precision packed RGB 32:32:32, 96bpp, RGBRGB..., little-endian
	//AV_PIX_FMT_RGBAF32BE AVPixelFormat = C.AV_PIX_FMT_RGBAF32BE ///< IEEE-754 single precision packed RGBA 32:32:32:32, 128bpp, RGBARGBA..., big-endian
	//AV_PIX_FMT_RGBAF32LE AVPixelFormat = C.AV_PIX_FMT_RGBAF32LE ///< IEEE-754 single precision packed RGBA 32:32:32:32, 128bpp, RGBARGBA..., little-endian
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVPixelFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v AVPixelFormat) String() string {
	if f := AVUtil_get_pix_fmt_name(v); f != "" {
		return f
	}
	return fmt.Sprintf("AVPixelFormat(%d)", int(v))
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

func AVUtil_get_pix_fmt_name(pixfmt AVPixelFormat) string {
	return C.GoString(C.av_get_pix_fmt_name((C.enum_AVPixelFormat)(pixfmt)))
}

func AVUtil_get_pix_fmt_desc(pixfmt AVPixelFormat) *AVPixFmtDescriptor {
	return (*AVPixFmtDescriptor)(C.av_pix_fmt_desc_get(C.enum_AVPixelFormat(pixfmt)))
}
