package libexif

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libexif
#include <stdlib.h>
#include <libexif/exif-data-type.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	DataType C.ExifDataType
)

////////////////////////////////////////////////////////////////////////////////
// GLBOALS

const (
	EXIF_DATA_TYPE_UNCOMPRESSED_CHUNKY DataType = C.EXIF_DATA_TYPE_UNCOMPRESSED_CHUNKY
	EXIF_DATA_TYPE_UNCOMPRESSED_PLANAR DataType = C.EXIF_DATA_TYPE_UNCOMPRESSED_PLANAR
	EXIF_DATA_TYPE_UNCOMPRESSED_YCC    DataType = C.EXIF_DATA_TYPE_UNCOMPRESSED_YCC
	EXIF_DATA_TYPE_COMPRESSED          DataType = C.EXIF_DATA_TYPE_COMPRESSED
	EXIF_DATA_TYPE_UNKNOWN             DataType = C.EXIF_DATA_TYPE_UNKNOWN
)
