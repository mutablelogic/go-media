package ffmpeg

import (
	"runtime"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat libavutil
#include <libavformat/avformat.h>

extern int avio_read_callback(void* userInfo, uint8_t* buf, int buf_size);
extern int avio_write_callback(void* userInfo, uint8_t* buf, int buf_size);
extern int64_t avio_seek_callback(void* userInfo, int64_t offset, int whence);

static AVIOContext* avio_alloc_context_(int sz, int writeable, void* userInfo) {
	uint8_t* buf = av_malloc(sz);
	if (!buf) {
		return NULL;
	}
	return avio_alloc_context(buf, sz, writeable, userInfo,avio_read_callback,avio_write_callback,avio_seek_callback);
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVIOContext C.struct_AVIOContext
)

// Wrapper around AVIOContext with callbacks
type AVIOContextEx struct {
	*AVIOContext
	cb  AVIOContextCallback
	pin *runtime.Pinner
}

type AVIOContextCallback interface {
	Reader(buf []byte) int
	Writer(buf []byte) int
	Seeker(offset int64, whence int) int64
}

// Callbacks for avio_alloc_context
type AVFormat_avio_read_func func(buf []byte) int
type AVFormat_avio_write_func func(buf []byte) int
type AVFormat_avio_seek_func func(offset int64, whence int) int64

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS

// avio_alloc_context
func AVFormat_avio_alloc_context(sz int, writeable bool, callback AVIOContextCallback) *AVIOContextEx {
	// Create a context
	ctx := new(AVIOContextEx)
	ctx.cb = callback
	ctx.pin = new(runtime.Pinner)
	ctx.pin.Pin(ctx.cb)
	ctx.pin.Pin(ctx.pin)

	// Allocate the context
	userInfo := unsafe.Pointer(ctx)
	ctx.AVIOContext = (*AVIOContext)(C.avio_alloc_context_(
		C.int(sz),
		boolToInt(writeable),
		userInfo,
	))
	if ctx.AVIOContext == nil {
		return nil
	}

	return ctx
}

// avio_context_free
func AVFormat_avio_context_free(ctx *AVIOContextEx) {
	C.av_free(unsafe.Pointer(ctx.buffer))
	C.avio_context_free((**C.struct_AVIOContext)(unsafe.Pointer(&ctx.AVIOContext)))
	ctx.pin.Unpin()
}

// avio_w8
func AVFormat_avio_w8(ctx *AVIOContextEx, b int) {
	C.avio_w8((*C.struct_AVIOContext)(ctx.AVIOContext), C.int(b))
}

// avio_write
func AVFormat_avio_write(ctx *AVIOContextEx, buf []byte) {
	C.avio_write((*C.struct_AVIOContext)(ctx.AVIOContext), (*C.uint8_t)(&buf[0]), C.int(len(buf)))
}

// avio_wl64
func AVFormat_avio_wl64(ctx *AVIOContextEx, b uint64) {
	C.avio_wl64((*C.struct_AVIOContext)(ctx.AVIOContext), C.uint64_t(b))
}

// avio_put_str
func AVFormat_avio_put_str(ctx *AVIOContextEx, str string) int {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	return int(C.avio_put_str((*C.struct_AVIOContext)(ctx.AVIOContext), cStr))
}

// avio_seek
// whence: SEEK_SET, SEEK_CUR, SEEK_END (like fseek) and AVSEEK_SIZE
func AVFormat_avio_seek(ctx *AVIOContextEx, offset int64, whence int) int64 {
	return int64(C.avio_seek((*C.struct_AVIOContext)(ctx.AVIOContext), C.int64_t(offset), C.int(whence)))
}

// avio_flush
func AVFormat_avio_flush(ctx *AVIOContextEx) {
	C.avio_flush((*C.struct_AVIOContext)(ctx.AVIOContext))
}

// avio_read
func AVFormat_avio_read(ctx *AVIOContextEx, buf []byte) int {
	return int(C.avio_read((*C.struct_AVIOContext)(ctx.AVIOContext), (*C.uint8_t)(&buf[0]), C.int(len(buf))))
}

////////////////////////////////////////////////////////////////////////////////
// CALLBACKS

//export avio_read_callback
func avio_read_callback(userInfo unsafe.Pointer, buf *C.uint8_t, size C.int) C.int {
	ctx := (*AVIOContextEx)(userInfo)
	return C.int(ctx.cb.Reader(cByteSlice(unsafe.Pointer(buf), size)))
}

//export avio_write_callback
func avio_write_callback(userInfo unsafe.Pointer, buf *C.uint8_t, size C.int) C.int {
	ctx := (*AVIOContextEx)(userInfo)
	return C.int(ctx.cb.Writer(cByteSlice(unsafe.Pointer(buf), size)))
}

//export avio_seek_callback
func avio_seek_callback(userInfo unsafe.Pointer, offset C.int64_t, whence C.int) C.int64_t {
	ctx := (*AVIOContextEx)(userInfo)
	return C.int64_t(ctx.cb.Seeker(int64(offset), int(whence)))
}
