package ffmpeg

import (
	"fmt"
	"sync"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat libavutil
#include <libavformat/avformat.h>

typedef const uint8_t const_char_t;
extern int avio_read_callback(void* userInfo, uint8_t* buf, int buf_size);
extern int avio_write_callback(void* userInfo, const_char_t* buf, int buf_size);
extern int64_t avio_seek_callback(void* userInfo, int64_t offset, int whence);

static AVIOContext* avio_alloc_context_(int sz, int writeable, void* userInfo) {
	uint8_t* buf = av_malloc(sz);
	if (!buf) {
		return NULL;
	}
	return avio_alloc_context(buf, sz, writeable, userInfo, avio_read_callback, avio_write_callback, avio_seek_callback);
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Wrapper around AVIOContext with callbacks
type AVIOContextEx struct {
	*AVIOContext
}

// Callbacks for AVIOContextEx
type AVIOContextCallback interface {
	Reader(buf []byte) int
	Writer(buf []byte) int
	Seeker(offset int64, whence int) int64
}

var (
	lock      sync.RWMutex
	callbacks = make(map[uintptr]AVIOContextCallback)
)

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS

// avio_alloc_context
func AVFormat_avio_alloc_context(sz int, writeable bool, callback AVIOContextCallback) *AVIOContextEx {
	// Create a context
	ctx := new(AVIOContextEx)

	// Set the callback
	ptr := uintptr(unsafe.Pointer(ctx))
	lock.Lock()
	callbacks[ptr] = callback
	lock.Unlock()

	// Allocate the context
	ctx.AVIOContext = (*AVIOContext)(C.avio_alloc_context_(
		C.int(sz),
		boolToInt(writeable),
		unsafe.Pointer(ctx),
	))
	if ctx.AVIOContext == nil {
		return nil
	}

	return ctx
}

// Create and initialize a AVIOContext for accessing the resource indicated by url.
func AVFormat_avio_open(url string, flags AVIOFlag) (*AVIOContextEx, error) {
	ctx := new(AVIOContextEx)
	cUrl := C.CString(url)
	defer C.free(unsafe.Pointer(cUrl))
	if err := AVError(C.avio_open((**C.struct_AVIOContext)(unsafe.Pointer(&ctx.AVIOContext)), cUrl, C.int(flags))); err != 0 {
		return nil, err
	}

	// Return success
	return ctx, nil
}

// Close the resource and free it.
// This function can only be used if it was opened by avio_open().
func AVFormat_avio_close(ctx *AVIOContextEx) error {
	ctx_ := (*C.struct_AVIOContext)(ctx.AVIOContext)
	if err := AVError(C.avio_closep(&ctx_)); err != 0 {
		return err
	}

	// Return success
	return nil
}

// avio_context_free
func AVFormat_avio_context_free(ctx *AVIOContextEx) {
	C.av_free(unsafe.Pointer(ctx.buffer))
	C.avio_context_free((**C.struct_AVIOContext)(unsafe.Pointer(&ctx.AVIOContext)))

	// Remove the callback
	ptr := uintptr(unsafe.Pointer(ctx))
	lock.Lock()
	delete(callbacks, ptr)
	lock.Unlock()
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
	lock.RLock()
	defer lock.RUnlock()

	ptr := uintptr(userInfo)
	callback, ok := callbacks[ptr]
	if !ok {
		panic("avio_read_callback: callback not found")
	}
	return C.int(callback.Reader(cByteSlice(unsafe.Pointer(buf), size)))
}

//export avio_write_callback
func avio_write_callback(userInfo unsafe.Pointer, buf *C.const_char_t, size C.int) C.int {
	lock.RLock()
	defer lock.RUnlock()

	ptr := uintptr(userInfo)
	callback, ok := callbacks[ptr]
	if !ok {
		panic("avio_write_callback: callback not found " + fmt.Sprint(ptr))
	}
	return C.int(callback.Writer(cByteSlice(unsafe.Pointer(buf), size)))
}

//export avio_seek_callback
func avio_seek_callback(userInfo unsafe.Pointer, offset C.int64_t, whence C.int) C.int64_t {
	lock.RLock()
	defer lock.RUnlock()

	ptr := uintptr(userInfo)
	callback, ok := callbacks[ptr]
	if !ok {
		panic("avio_seek_callback: callback not found")
	}
	return C.int64_t(callback.Seeker(int64(offset), int(whence)))
}
