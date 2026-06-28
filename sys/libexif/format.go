package libexif

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libexif
#include <stdlib.h>
#include <libexif/exif-format.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Format C.ExifFormat
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	EXIF_FORMAT_BYTE      Format = C.EXIF_FORMAT_BYTE
	EXIF_FORMAT_ASCII     Format = C.EXIF_FORMAT_ASCII
	EXIF_FORMAT_SHORT     Format = C.EXIF_FORMAT_SHORT
	EXIF_FORMAT_LONG      Format = C.EXIF_FORMAT_LONG
	EXIF_FORMAT_RATIONAL  Format = C.EXIF_FORMAT_RATIONAL
	EXIF_FORMAT_SBYTE     Format = C.EXIF_FORMAT_SBYTE
	EXIF_FORMAT_UNDEFINED Format = C.EXIF_FORMAT_UNDEFINED
	EXIF_FORMAT_SSHORT    Format = C.EXIF_FORMAT_SSHORT
	EXIF_FORMAT_SLONG     Format = C.EXIF_FORMAT_SLONG
	EXIF_FORMAT_SRATIONAL Format = C.EXIF_FORMAT_SRATIONAL
	EXIF_FORMAT_FLOAT     Format = C.EXIF_FORMAT_FLOAT
	EXIF_FORMAT_DOUBLE    Format = C.EXIF_FORMAT_DOUBLE
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

func Exif_format_get_name(format Format) string {
	return C.GoString(C.exif_format_get_name(C.ExifFormat(format)))
}

func Exif_format_get_size(format Format) uint {
	return uint(C.exif_format_get_size(C.ExifFormat(format)))
}
