package libexif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libexif
#include <stdlib.h>
#include <libexif/exif-utils.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Rational struct {
		Numerator   uint32
		Denominator uint32
	}
	SRational struct {
		Numerator   int32
		Denominator int32
	}
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - GET

func Exif_get_short(data []byte, order ByteOrder) uint16 {
	return uint16(C.exif_get_short((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order)))
}

func Exif_get_sshort(data []byte, order ByteOrder) int16 {
	return int16(C.exif_get_sshort((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order)))
}

func Exif_get_long(data []byte, order ByteOrder) uint32 {
	return uint32(C.exif_get_long((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order)))
}

func Exif_get_slong(data []byte, order ByteOrder) int32 {
	return int32(C.exif_get_slong((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order)))
}

func Exif_get_rational(data []byte, order ByteOrder) Rational {
	r := C.exif_get_rational((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order))
	return Rational{
		Numerator:   uint32(r.numerator),
		Denominator: uint32(r.denominator),
	}
}

func Exif_get_srational(data []byte, order ByteOrder) SRational {
	r := C.exif_get_srational((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order))
	return SRational{
		Numerator:   int32(r.numerator),
		Denominator: int32(r.denominator),
	}
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - SET

func Exif_set_short(data []byte, order ByteOrder, value uint16) {
	C.exif_set_short((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order), C.ExifShort(value))
}

func Exif_set_sshort(data []byte, order ByteOrder, value int16) {
	C.exif_set_sshort((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order), C.ExifSShort(value))
}

func Exif_set_long(data []byte, order ByteOrder, value uint32) {
	C.exif_set_long((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order), C.ExifLong(value))
}

func Exif_set_slong(data []byte, order ByteOrder, value int32) {
	C.exif_set_slong((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order), C.ExifSLong(value))
}

func Exif_set_rational(data []byte, order ByteOrder, value Rational) {
	r := C.ExifRational{
		numerator:   C.ExifLong(value.Numerator),
		denominator: C.ExifLong(value.Denominator),
	}
	C.exif_set_rational((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order), r)
}

func Exif_set_srational(data []byte, order ByteOrder, value SRational) {
	r := C.ExifSRational{
		numerator:   C.ExifSLong(value.Numerator),
		denominator: C.ExifSLong(value.Denominator),
	}
	C.exif_set_srational((*C.uchar)(unsafe.Pointer(&data[0])), C.ExifByteOrder(order), r)
}
