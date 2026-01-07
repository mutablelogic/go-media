package ffmpeg

import (
	"encoding/json"
	"math"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/opt.h>
#include <stdlib.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVOption       C.struct_AVOption
	AVOptionType   C.enum_AVOptionType
	AVOptionRanges C.struct_AVOptionRanges
	AVOptionRange  C.struct_AVOptionRange
)

////////////////////////////////////////////////////////////////////////////////
// METHODS - AVClass

// Name returns the class name.
func (c *AVClass) Name() string {
	return C.GoString((*C.struct_AVClass)(unsafe.Pointer(c)).class_name)
}

////////////////////////////////////////////////////////////////////////////////
// METHODS - AVOption

// Name returns the option name.
func (o *AVOption) Name() string {
	return C.GoString((*C.struct_AVOption)(unsafe.Pointer(o)).name)
}

// Help returns the short English help text for the option.
func (o *AVOption) Help() string {
	help := (*C.struct_AVOption)(unsafe.Pointer(o)).help
	if help == nil {
		return ""
	}
	return C.GoString(help)
}

// Type returns the option type.
func (o *AVOption) Type() AVOptionType {
	return AVOptionType((*C.struct_AVOption)(unsafe.Pointer(o))._type)
}

// Offset returns the offset of the option in the context structure.
func (o *AVOption) Offset() int {
	return int((*C.struct_AVOption)(unsafe.Pointer(o)).offset)
}

// Min returns the minimum valid value for numeric options.
func (o *AVOption) Min() float64 {
	return float64((*C.struct_AVOption)(unsafe.Pointer(o)).min)
}

// Max returns the maximum valid value for numeric options.
func (o *AVOption) Max() float64 {
	return float64((*C.struct_AVOption)(unsafe.Pointer(o)).max)
}

// DefaultVal returns the default value as an interface{}, type depends on the option type.
// For const options, returns their constant value; returns nil if no default is set or the type is unsupported.
func (o *AVOption) DefaultVal() interface{} {
	opt := (*C.struct_AVOption)(unsafe.Pointer(o))
	switch o.Type() {
	case AV_OPT_TYPE_INT, AV_OPT_TYPE_INT64, AV_OPT_TYPE_UINT64:
		// Cast the union bytes to int64
		return int64(*(*C.int64_t)(unsafe.Pointer(&opt.default_val)))
	case AV_OPT_TYPE_UINT:
		// Cast the union bytes to uint
		return uint(*(*C.int64_t)(unsafe.Pointer(&opt.default_val)))
	case AV_OPT_TYPE_DOUBLE, AV_OPT_TYPE_DURATION:
		// Cast the union bytes to double
		return float64(*(*C.double)(unsafe.Pointer(&opt.default_val)))
	case AV_OPT_TYPE_FLOAT:
		// Cast the union bytes to double (FFmpeg stores float as double in union)
		return float32(*(*C.double)(unsafe.Pointer(&opt.default_val)))
	case AV_OPT_TYPE_STRING:
		// Cast the union bytes to string pointer
		strPtr := *(**C.char)(unsafe.Pointer(&opt.default_val))
		if strPtr != nil {
			return C.GoString(strPtr)
		}
		return nil
	case AV_OPT_TYPE_RATIONAL, AV_OPT_TYPE_VIDEO_RATE:
		// Cast the union bytes to AVRational
		q := *(*C.AVRational)(unsafe.Pointer(&opt.default_val))
		return AVRational(q)
	case AV_OPT_TYPE_FLAGS:
		// Cast the union bytes to uint
		return uint(*(*C.int64_t)(unsafe.Pointer(&opt.default_val)))
	case AV_OPT_TYPE_BOOL:
		// Cast the union bytes to int64 and check if non-zero
		return *(*C.int64_t)(unsafe.Pointer(&opt.default_val)) != 0
	case AV_OPT_TYPE_CONST:
		// For const, the default_val contains the value of the constant
		return int64(*(*C.int64_t)(unsafe.Pointer(&opt.default_val)))
	default:
		return nil
	}
}

// Flags returns the option flags (AV_OPT_FLAG_*).
func (o *AVOption) Flags() int {
	return int((*C.struct_AVOption)(unsafe.Pointer(o)).flags)
}

// Unit returns the logical unit to which the option belongs.
// For const options, this is the name of the option they provide values for.
func (o *AVOption) Unit() string {
	unit := (*C.struct_AVOption)(unsafe.Pointer(o)).unit
	if unit == nil {
		return ""
	}
	return C.GoString(unit)
}

// MarshalJSON implements json.Marshaler for AVOption.
func (o *AVOption) MarshalJSON() ([]byte, error) {
	type optJSON struct {
		Name    string       `json:"name"`
		Help    string       `json:"help,omitempty"`
		Type    AVOptionType `json:"type"`
		Default interface{}  `json:"default,omitempty"`
		Min     *float64     `json:"min,omitempty"`
		Max     *float64     `json:"max,omitempty"`
		Unit    string       `json:"unit,omitempty"`
	}

	result := optJSON{
		Name: o.Name(),
		Help: o.Help(),
		Type: o.Type(),
		Unit: o.Unit(),
	}

	// Handle default value, checking for NaN in float types
	defVal := o.DefaultVal()
	if defVal != nil {
		switch v := defVal.(type) {
		case float64:
			if !math.IsNaN(v) && !math.IsInf(v, 0) {
				result.Default = v
			}
		default:
			result.Default = v
		}
	}

	// Only include min/max for numeric types, and skip NaN values
	switch o.Type() {
	case AV_OPT_TYPE_INT, AV_OPT_TYPE_INT64, AV_OPT_TYPE_UINT, AV_OPT_TYPE_UINT64,
		AV_OPT_TYPE_DOUBLE, AV_OPT_TYPE_FLOAT, AV_OPT_TYPE_DURATION:
		min := o.Min()
		max := o.Max()
		if !math.IsNaN(min) && !math.IsInf(min, 0) {
			result.Min = &min
		}
		if !math.IsNaN(max) && !math.IsInf(max, 0) {
			result.Max = &max
		}
	}

	return json.Marshal(result)
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS - AVOptionType

const (
	AV_OPT_TYPE_FLAGS      AVOptionType = C.AV_OPT_TYPE_FLAGS // Bitmask flags option (unsigned int)
	AV_OPT_TYPE_INT        AVOptionType = C.AV_OPT_TYPE_INT
	AV_OPT_TYPE_INT64      AVOptionType = C.AV_OPT_TYPE_INT64
	AV_OPT_TYPE_UINT       AVOptionType = C.AV_OPT_TYPE_UINT
	AV_OPT_TYPE_UINT64     AVOptionType = C.AV_OPT_TYPE_UINT64
	AV_OPT_TYPE_DOUBLE     AVOptionType = C.AV_OPT_TYPE_DOUBLE
	AV_OPT_TYPE_FLOAT      AVOptionType = C.AV_OPT_TYPE_FLOAT
	AV_OPT_TYPE_STRING     AVOptionType = C.AV_OPT_TYPE_STRING
	AV_OPT_TYPE_RATIONAL   AVOptionType = C.AV_OPT_TYPE_RATIONAL
	AV_OPT_TYPE_BINARY     AVOptionType = C.AV_OPT_TYPE_BINARY
	AV_OPT_TYPE_DICT       AVOptionType = C.AV_OPT_TYPE_DICT
	AV_OPT_TYPE_CONST      AVOptionType = C.AV_OPT_TYPE_CONST // Named constant used to define possible values for FLAGS and other options
	AV_OPT_TYPE_IMAGE_SIZE AVOptionType = C.AV_OPT_TYPE_IMAGE_SIZE
	AV_OPT_TYPE_PIXEL_FMT  AVOptionType = C.AV_OPT_TYPE_PIXEL_FMT
	AV_OPT_TYPE_SAMPLE_FMT AVOptionType = C.AV_OPT_TYPE_SAMPLE_FMT
	AV_OPT_TYPE_VIDEO_RATE AVOptionType = C.AV_OPT_TYPE_VIDEO_RATE
	AV_OPT_TYPE_DURATION   AVOptionType = C.AV_OPT_TYPE_DURATION
	AV_OPT_TYPE_COLOR      AVOptionType = C.AV_OPT_TYPE_COLOR
	AV_OPT_TYPE_BOOL       AVOptionType = C.AV_OPT_TYPE_BOOL
	AV_OPT_TYPE_CHLAYOUT   AVOptionType = C.AV_OPT_TYPE_CHLAYOUT
	AV_OPT_TYPE_FLAG_ARRAY AVOptionType = C.AV_OPT_TYPE_FLAG_ARRAY
)

// MarshalJSON implements json.Marshaler for AVOptionType.
func (t AVOptionType) MarshalJSON() ([]byte, error) {
	var s string
	switch t {
	case AV_OPT_TYPE_FLAGS:
		s = "flags"
	case AV_OPT_TYPE_INT:
		s = "int"
	case AV_OPT_TYPE_INT64:
		s = "int64"
	case AV_OPT_TYPE_UINT:
		s = "uint"
	case AV_OPT_TYPE_UINT64:
		s = "uint64"
	case AV_OPT_TYPE_DOUBLE:
		s = "double"
	case AV_OPT_TYPE_FLOAT:
		s = "float"
	case AV_OPT_TYPE_STRING:
		s = "string"
	case AV_OPT_TYPE_RATIONAL:
		s = "rational"
	case AV_OPT_TYPE_BINARY:
		s = "binary"
	case AV_OPT_TYPE_DICT:
		s = "dict"
	case AV_OPT_TYPE_CONST:
		s = "const"
	case AV_OPT_TYPE_IMAGE_SIZE:
		s = "image_size"
	case AV_OPT_TYPE_PIXEL_FMT:
		s = "pixel_fmt"
	case AV_OPT_TYPE_SAMPLE_FMT:
		s = "sample_fmt"
	case AV_OPT_TYPE_VIDEO_RATE:
		s = "video_rate"
	case AV_OPT_TYPE_DURATION:
		s = "duration"
	case AV_OPT_TYPE_COLOR:
		s = "color"
	case AV_OPT_TYPE_BOOL:
		s = "bool"
	case AV_OPT_TYPE_CHLAYOUT:
		s = "chlayout"
	case AV_OPT_TYPE_FLAG_ARRAY:
		s = "flag_array"
	default:
		s = "unknown"
	}
	return json.Marshal(s)
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS - Search flags

const (
	// Search in possible children of the given object first.
	AV_OPT_SEARCH_CHILDREN = C.AV_OPT_SEARCH_CHILDREN

	// Search fake options (option names used in documented examples).
	AV_OPT_SEARCH_FAKE_OBJ = C.AV_OPT_SEARCH_FAKE_OBJ

	// Allow to pass options as flags instead of values.
	AV_OPT_FLAG_IMPLICIT_KEY = C.AV_OPT_FLAG_IMPLICIT_KEY
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS - Serialization flags

const (
	// Serialize options that are not set to default values.
	AV_OPT_SERIALIZE_SKIP_DEFAULTS = C.AV_OPT_SERIALIZE_SKIP_DEFAULTS

	// Serialize options that are exactly equal to default values.
	AV_OPT_SERIALIZE_OPT_FLAGS_EXACT = C.AV_OPT_SERIALIZE_OPT_FLAGS_EXACT
)

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_show2

// AVUtil_opt_show2 shows the obj options. Returns 0 on success, a negative value on error.
func AVUtil_opt_show2(obj unsafe.Pointer, av_log_obj unsafe.Pointer, req_flags, rej_flags int) error {
	if ret := AVError(C.av_opt_show2(obj, av_log_obj, C.int(req_flags), C.int(rej_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_defaults

// AVUtil_opt_set_defaults sets the values of all AVOption fields to their default values.
func AVUtil_opt_set_defaults(obj unsafe.Pointer) {
	C.av_opt_set_defaults(obj)
}

// AVUtil_opt_set_defaults2 sets the values of all AVOption fields to their default values.
// Only options which are not set to default values will be set.
func AVUtil_opt_set_defaults2(obj unsafe.Pointer, mask, flags int) {
	C.av_opt_set_defaults2(obj, C.int(mask), C.int(flags))
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set (string)

// AVUtil_opt_set sets a string option value.
func AVUtil_opt_set(obj unsafe.Pointer, name, value string, search_flags int) error {
	cName := C.CString(name)
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cValue))

	if ret := AVError(C.av_opt_set(obj, cName, cValue, C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_int

// AVUtil_opt_set_int sets an integer option value.
func AVUtil_opt_set_int(obj unsafe.Pointer, name string, value int64, search_flags int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	if ret := AVError(C.av_opt_set_int(obj, cName, C.int64_t(value), C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_double

// AVUtil_opt_set_double sets a double option value.
func AVUtil_opt_set_double(obj unsafe.Pointer, name string, value float64, search_flags int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	if ret := AVError(C.av_opt_set_double(obj, cName, C.double(value), C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_q

// AVUtil_opt_set_q sets a rational option value.
func AVUtil_opt_set_q(obj unsafe.Pointer, name string, value AVRational, search_flags int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	if ret := AVError(C.av_opt_set_q(obj, cName, (C.AVRational)(value), C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_bin

// AVUtil_opt_set_bin sets a binary option value.
func AVUtil_opt_set_bin(obj unsafe.Pointer, name string, value []byte, search_flags int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var ptr unsafe.Pointer
	if len(value) > 0 {
		ptr = unsafe.Pointer(&value[0])
	}

	if ret := AVError(C.av_opt_set_bin(obj, cName, (*C.uint8_t)(ptr), C.int(len(value)), C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_image_size

// AVUtil_opt_set_image_size sets an image size option value.
func AVUtil_opt_set_image_size(obj unsafe.Pointer, name string, width, height, search_flags int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	if ret := AVError(C.av_opt_set_image_size(obj, cName, C.int(width), C.int(height), C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_pixel_fmt

// AVUtil_opt_set_pixel_fmt sets a pixel format option value.
func AVUtil_opt_set_pixel_fmt(obj unsafe.Pointer, name string, fmt AVPixelFormat, search_flags int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	if ret := AVError(C.av_opt_set_pixel_fmt(obj, cName, (C.enum_AVPixelFormat)(fmt), C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_sample_fmt

// AVUtil_opt_set_sample_fmt sets a sample format option value.
func AVUtil_opt_set_sample_fmt(obj unsafe.Pointer, name string, fmt AVSampleFormat, search_flags int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	if ret := AVError(C.av_opt_set_sample_fmt(obj, cName, (C.enum_AVSampleFormat)(fmt), C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_video_rate

// AVUtil_opt_set_video_rate sets a video rate option value.
func AVUtil_opt_set_video_rate(obj unsafe.Pointer, name string, rate AVRational, search_flags int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	if ret := AVError(C.av_opt_set_video_rate(obj, cName, (C.AVRational)(rate), C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_set_channel_layout

// AVUtil_opt_set_channel_layout sets a channel layout option value.
func AVUtil_opt_set_channel_layout(obj unsafe.Pointer, name string, layout *AVChannelLayout, search_flags int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	if ret := AVError(C.av_opt_set_chlayout(obj, cName, (*C.AVChannelLayout)(unsafe.Pointer(layout)), C.int(search_flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_get (string)

// AVUtil_opt_get gets a string option value.
func AVUtil_opt_get(obj unsafe.Pointer, name string, search_flags int) (string, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var cValue *C.uint8_t
	if ret := AVError(C.av_opt_get(obj, cName, C.int(search_flags), &cValue)); ret != 0 {
		return "", ret
	}
	defer C.av_free(unsafe.Pointer(cValue))

	return C.GoString((*C.char)(unsafe.Pointer(cValue))), nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_get_int

// AVUtil_opt_get_int gets an integer option value.
func AVUtil_opt_get_int(obj unsafe.Pointer, name string, search_flags int) (int64, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var value C.int64_t
	if ret := AVError(C.av_opt_get_int(obj, cName, C.int(search_flags), &value)); ret != 0 {
		return 0, ret
	}
	return int64(value), nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_get_double

// AVUtil_opt_get_double gets a double option value.
func AVUtil_opt_get_double(obj unsafe.Pointer, name string, search_flags int) (float64, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var value C.double
	if ret := AVError(C.av_opt_get_double(obj, cName, C.int(search_flags), &value)); ret != 0 {
		return 0, ret
	}
	return float64(value), nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_get_q

// AVUtil_opt_get_q gets a rational option value.
func AVUtil_opt_get_q(obj unsafe.Pointer, name string, search_flags int) (AVRational, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var value C.AVRational
	if ret := AVError(C.av_opt_get_q(obj, cName, C.int(search_flags), &value)); ret != 0 {
		return AVRational{}, ret
	}
	return AVRational(value), nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_get_image_size

// AVUtil_opt_get_image_size gets an image size option value.
func AVUtil_opt_get_image_size(obj unsafe.Pointer, name string, search_flags int) (int, int, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var width, height C.int
	if ret := AVError(C.av_opt_get_image_size(obj, cName, C.int(search_flags), &width, &height)); ret != 0 {
		return 0, 0, ret
	}
	return int(width), int(height), nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_get_pixel_fmt

// AVUtil_opt_get_pixel_fmt gets a pixel format option value.
func AVUtil_opt_get_pixel_fmt(obj unsafe.Pointer, name string, search_flags int) (AVPixelFormat, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var fmt C.enum_AVPixelFormat
	if ret := AVError(C.av_opt_get_pixel_fmt(obj, cName, C.int(search_flags), &fmt)); ret != 0 {
		return AV_PIX_FMT_NONE, ret
	}
	return AVPixelFormat(fmt), nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_get_sample_fmt

// AVUtil_opt_get_sample_fmt gets a sample format option value.
func AVUtil_opt_get_sample_fmt(obj unsafe.Pointer, name string, search_flags int) (AVSampleFormat, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var fmt C.enum_AVSampleFormat
	if ret := AVError(C.av_opt_get_sample_fmt(obj, cName, C.int(search_flags), &fmt)); ret != 0 {
		return AV_SAMPLE_FMT_NONE, ret
	}
	return AVSampleFormat(fmt), nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_get_video_rate

// AVUtil_opt_get_video_rate gets a video rate option value.
func AVUtil_opt_get_video_rate(obj unsafe.Pointer, name string, search_flags int) (AVRational, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var rate C.AVRational
	if ret := AVError(C.av_opt_get_video_rate(obj, cName, C.int(search_flags), &rate)); ret != 0 {
		return AVRational{}, ret
	}
	return AVRational(rate), nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_get_channel_layout

// AVUtil_opt_get_channel_layout gets a channel layout option value.
func AVUtil_opt_get_channel_layout(obj unsafe.Pointer, name string, search_flags int) (*AVChannelLayout, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	layout := &AVChannelLayout{}
	if ret := AVError(C.av_opt_get_chlayout(obj, cName, C.int(search_flags), (*C.AVChannelLayout)(unsafe.Pointer(layout)))); ret != 0 {
		return nil, ret
	}
	return layout, nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_find

// AVUtil_opt_find looks for an option in an object.
// Returns a pointer to the option found, or nil if no option was found.
func AVUtil_opt_find(obj unsafe.Pointer, name string, unit string, opt_flags, search_flags int) *AVOption {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var cUnit *C.char
	if unit != "" {
		cUnit = C.CString(unit)
		defer C.free(unsafe.Pointer(cUnit))
	}

	return (*AVOption)(C.av_opt_find(obj, cName, cUnit, C.int(opt_flags), C.int(search_flags)))
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_find2

// AVUtil_opt_find2 looks for an option in an object.
// Returns a pointer to the option found, or nil if no option was found.
// target_obj is set to the object where the option was found.
func AVUtil_opt_find2(obj unsafe.Pointer, name string, unit string, opt_flags, search_flags int) (*AVOption, unsafe.Pointer) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var cUnit *C.char
	if unit != "" {
		cUnit = C.CString(unit)
		defer C.free(unsafe.Pointer(cUnit))
	}

	var targetObj unsafe.Pointer
	opt := (*AVOption)(C.av_opt_find2(obj, cName, cUnit, C.int(opt_flags), C.int(search_flags), &targetObj))
	return opt, targetObj
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_copy

// AVUtil_opt_copy copies options from src to dst.
func AVUtil_opt_copy(dst, src unsafe.Pointer) error {
	if ret := AVError(C.av_opt_copy(dst, src)); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_query_ranges

// AVUtil_opt_query_ranges gets a list of allowed ranges for the given option.
// The returned ranges must be freed with AVUtil_opt_freep_ranges.
func AVUtil_opt_query_ranges(ranges **AVOptionRanges, obj unsafe.Pointer, key string, flags int) error {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	if ret := AVError(C.av_opt_query_ranges((**C.struct_AVOptionRanges)(unsafe.Pointer(ranges)), obj, cKey, C.int(flags))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_freep_ranges

// AVUtil_opt_freep_ranges frees an AVOptionRanges struct and set it to nil.
func AVUtil_opt_freep_ranges(ranges **AVOptionRanges) {
	C.av_opt_freep_ranges((**C.struct_AVOptionRanges)(unsafe.Pointer(ranges)))
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_is_set_to_default

// AVUtil_opt_is_set_to_default checks if given option is set to its default value.
// Returns 1 if set to default, 0 if not default, negative on error.
func AVUtil_opt_is_set_to_default(obj unsafe.Pointer, opt *AVOption) int {
	return int(C.av_opt_is_set_to_default(obj, (*C.struct_AVOption)(unsafe.Pointer(opt))))
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_is_set_to_default_by_name

// AVUtil_opt_is_set_to_default_by_name checks if given option is set to its default value.
// Returns 1 if set to default, 0 if not default, negative on error.
func AVUtil_opt_is_set_to_default_by_name(obj unsafe.Pointer, name string, search_flags int) int {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	return int(C.av_opt_is_set_to_default_by_name(obj, cName, C.int(search_flags)))
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_next

// AVUtil_opt_next iterates over AVOptions-enabled objects.
// Pass nil for prev to get the first option, then pass the returned option to get the next one.
// Returns nil when there are no more options.
func AVUtil_opt_next(obj unsafe.Pointer, prev *AVOption) *AVOption {
	return (*AVOption)(C.av_opt_next(obj, (*C.struct_AVOption)(unsafe.Pointer(prev))))
}

// AVUtil_opt_list returns all options for an object as a slice.
// This is a convenience wrapper around AVUtil_opt_next.
func AVUtil_opt_list(obj unsafe.Pointer) []*AVOption {
	var options []*AVOption
	var prev *AVOption
	for {
		opt := AVUtil_opt_next(obj, prev)
		if opt == nil {
			break
		}
		options = append(options, opt)
		prev = opt
	}
	return options
}

// AVUtil_opt_list_from_class returns all options for an AVClass as a slice.
// This uses the FAKE_OBJ trick - treating the AVClass* as if it's an object
// whose first field points to the class. This is how ffmpeg's cmdutils does it.
func AVUtil_opt_list_from_class(class *AVClass) []*AVOption {
	if class == nil {
		return nil
	}
	// The FAKE_OBJ trick: cast AVClass* to void* and use it as an object
	// This works because the first field of every AVClass-based object is AVClass*
	return AVUtil_opt_list(unsafe.Pointer(&class))
}

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS - av_opt_serialize

// AVUtil_opt_serialize serializes object's options to a string.
// Returns the allocated string (must be freed with av_free), or empty string on error.
func AVUtil_opt_serialize(obj unsafe.Pointer, opt_flags, flags int, key_val_sep, pairs_sep byte) string {
	var cBuf *C.char
	if ret := C.av_opt_serialize(obj, C.int(opt_flags), C.int(flags), &cBuf, C.char(key_val_sep), C.char(pairs_sep)); ret < 0 {
		return ""
	}
	defer C.av_free(unsafe.Pointer(cBuf))

	return C.GoString(cBuf)
}
