package libraw

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libraw
#include <libraw/libraw.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Data            C.libraw_data_t
	ProcessedImage  C.libraw_processed_image_t
	IParams         C.libraw_iparams_t
	ImageSizes      C.libraw_image_sizes_t
	ImgOther        C.libraw_imgother_t
	LensInfo        C.libraw_lensinfo_t
	Thumbnail       C.libraw_thumbnail_t
	ImageFormat     C.enum_LibRaw_image_formats
	ThumbnailFormat C.enum_LibRaw_thumbnail_formats
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	IMAGE_JPEG   ImageFormat = C.LIBRAW_IMAGE_JPEG
	IMAGE_BITMAP ImageFormat = C.LIBRAW_IMAGE_BITMAP
	IMAGE_JPEGXL ImageFormat = C.LIBRAW_IMAGE_JPEGXL

	THUMBNAIL_UNKNOWN  ThumbnailFormat = C.LIBRAW_THUMBNAIL_UNKNOWN
	THUMBNAIL_JPEG     ThumbnailFormat = C.LIBRAW_THUMBNAIL_JPEG
	THUMBNAIL_BITMAP   ThumbnailFormat = C.LIBRAW_THUMBNAIL_BITMAP
	THUMBNAIL_BITMAP16 ThumbnailFormat = C.LIBRAW_THUMBNAIL_BITMAP16
	THUMBNAIL_H265     ThumbnailFormat = C.LIBRAW_THUMBNAIL_H265
	THUMBNAIL_JPEGXL   ThumbnailFormat = C.LIBRAW_THUMBNAIL_JPEGXL
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - LIFECYCLE

func Libraw_init(flags uint) *Data {
	return (*Data)(C.libraw_init(C.uint(flags)))
}

func Libraw_close(data *Data) {
	C.libraw_close((*C.libraw_data_t)(data))
}

func Libraw_recycle(data *Data) {
	C.libraw_recycle((*C.libraw_data_t)(data))
}

func Libraw_recycle_datastream(data *Data) {
	C.libraw_recycle_datastream((*C.libraw_data_t)(data))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - INFO ACCESSORS

func Libraw_get_iparams(data *Data) *IParams {
	return (*IParams)(C.libraw_get_iparams((*C.libraw_data_t)(data)))
}

func Libraw_get_lensinfo(data *Data) *LensInfo {
	return (*LensInfo)(C.libraw_get_lensinfo((*C.libraw_data_t)(data)))
}

func Libraw_get_imgother(data *Data) *ImgOther {
	return (*ImgOther)(C.libraw_get_imgother((*C.libraw_data_t)(data)))
}

func Libraw_get_thumbnail(data *Data) *Thumbnail {
	return (*Thumbnail)(&(*C.libraw_data_t)(data).thumbnail)
}

func Libraw_get_sizes(data *Data) *ImageSizes {
	return (*ImageSizes)(&(*C.libraw_data_t)(data).sizes)
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - DIMENSION GETTERS

func Libraw_get_raw_height(data *Data) int {
	return int(C.libraw_get_raw_height((*C.libraw_data_t)(data)))
}

func Libraw_get_raw_width(data *Data) int {
	return int(C.libraw_get_raw_width((*C.libraw_data_t)(data)))
}

func Libraw_get_iheight(data *Data) int {
	return int(C.libraw_get_iheight((*C.libraw_data_t)(data)))
}

func Libraw_get_iwidth(data *Data) int {
	return int(C.libraw_get_iwidth((*C.libraw_data_t)(data)))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - IPARAMS FIELD ACCESSORS

func IParams_make(p *IParams) string {
	return C.GoString(&(*C.libraw_iparams_t)(p).make[0])
}

func IParams_model(p *IParams) string {
	return C.GoString(&(*C.libraw_iparams_t)(p).model[0])
}

func IParams_software(p *IParams) string {
	return C.GoString(&(*C.libraw_iparams_t)(p).software[0])
}

func IParams_normalized_make(p *IParams) string {
	return C.GoString(&(*C.libraw_iparams_t)(p).normalized_make[0])
}

func IParams_normalized_model(p *IParams) string {
	return C.GoString(&(*C.libraw_iparams_t)(p).normalized_model[0])
}

func IParams_raw_count(p *IParams) uint {
	return uint((*C.libraw_iparams_t)(p).raw_count)
}

func IParams_colors(p *IParams) int {
	return int((*C.libraw_iparams_t)(p).colors)
}

func IParams_filters(p *IParams) uint {
	return uint((*C.libraw_iparams_t)(p).filters)
}

func IParams_xmpdata(p *IParams) []byte {
	raw := (*C.libraw_iparams_t)(p)
	if raw.xmpdata == nil || raw.xmplen == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(raw.xmpdata), C.int(raw.xmplen))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - IMGOTHER FIELD ACCESSORS

func ImgOther_iso_speed(p *ImgOther) float32 {
	return float32((*C.libraw_imgother_t)(p).iso_speed)
}

func ImgOther_shutter(p *ImgOther) float32 {
	return float32((*C.libraw_imgother_t)(p).shutter)
}

func ImgOther_aperture(p *ImgOther) float32 {
	return float32((*C.libraw_imgother_t)(p).aperture)
}

func ImgOther_focal_len(p *ImgOther) float32 {
	return float32((*C.libraw_imgother_t)(p).focal_len)
}

func ImgOther_timestamp(p *ImgOther) int64 {
	return int64((*C.libraw_imgother_t)(p).timestamp)
}

func ImgOther_shot_order(p *ImgOther) uint {
	return uint((*C.libraw_imgother_t)(p).shot_order)
}

func ImgOther_desc(p *ImgOther) string {
	return C.GoString(&(*C.libraw_imgother_t)(p).desc[0])
}

func ImgOther_artist(p *ImgOther) string {
	return C.GoString(&(*C.libraw_imgother_t)(p).artist[0])
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - THUMBNAIL FIELD ACCESSORS

func Thumbnail_format(t *Thumbnail) ThumbnailFormat {
	return ThumbnailFormat((*C.libraw_thumbnail_t)(t).tformat)
}

func Thumbnail_width(t *Thumbnail) uint16 {
	return uint16((*C.libraw_thumbnail_t)(t).twidth)
}

func Thumbnail_height(t *Thumbnail) uint16 {
	return uint16((*C.libraw_thumbnail_t)(t).theight)
}

func Thumbnail_length(t *Thumbnail) uint {
	return uint((*C.libraw_thumbnail_t)(t).tlength)
}

func Thumbnail_data(t *Thumbnail) []byte {
	raw := (*C.libraw_thumbnail_t)(t)
	if raw.thumb == nil || raw.tlength == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(raw.thumb), C.int(raw.tlength))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - PROCESSED IMAGE FIELD ACCESSORS

func ProcessedImage_type(img *ProcessedImage) ImageFormat {
	return ImageFormat((*C.libraw_processed_image_t)(img)._type)
}

func ProcessedImage_height(img *ProcessedImage) uint16 {
	return uint16((*C.libraw_processed_image_t)(img).height)
}

func ProcessedImage_width(img *ProcessedImage) uint16 {
	return uint16((*C.libraw_processed_image_t)(img).width)
}

func ProcessedImage_colors(img *ProcessedImage) uint16 {
	return uint16((*C.libraw_processed_image_t)(img).colors)
}

func ProcessedImage_bits(img *ProcessedImage) uint16 {
	return uint16((*C.libraw_processed_image_t)(img).bits)
}

func ProcessedImage_data(img *ProcessedImage) []byte {
	raw := (*C.libraw_processed_image_t)(img)
	if raw.data_size == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(&raw.data[0]), C.int(raw.data_size))
}
