package ffmpeg

import (
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"syscall"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>

extern int avio_read_callback(void* userInfo, uint8_t* buf, int buf_size);
extern int avio_write_callback(void* userInfo, uint8_t* buf, int buf_size);
extern int64_t avio_seek_callback(void* userInfo, int64_t offset, int whence);
static AVIOContext* avio_alloc_context_(unsigned char* buf, int sz, int writeable, void* userInfo, int r,int w,int s) {
	return avio_alloc_context(buf, sz, writeable, userInfo, r ? avio_read_callback : NULL,w ? avio_write_callback : NULL,s ? avio_seek_callback : NULL);
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVIOContext       C.struct_AVIOContext
	AVIOReadWriteFunc func(buf []byte) (int, error)
	AVIOSeekFunc      func(offset int64, whence int) (int64, error)
)

type AVIOContextEx struct {
	*AVIOContext
	buf    unsafe.Pointer
	reader AVIOReadWriteFunc
	writer AVIOReadWriteFunc
	seeker AVIOSeekFunc
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAVIOOpen(url *url.URL, flags AVIOFlag) (*AVIOContext, error) {
	ctx := (*C.AVIOContext)(nil)
	url_ := C.CString(url.String())
	defer C.free(unsafe.Pointer(url_))

	if err := AVError(C.avio_open(&ctx, url_, C.int(flags))); err != 0 {
		return nil, err
	} else {
		return (*AVIOContext)(ctx), nil
	}
}

func NewAVIOContext(size int, writeable bool, reader, writer AVIOReadWriteFunc, seeker AVIOSeekFunc) *AVIOContextEx {
	// Set up call
	ctxex := new(AVIOContextEx)
	userInfo := unsafe.Pointer(ctxex)
	r, w, s := boolToInt(reader != nil), boolToInt(writer != nil), boolToInt(seeker != nil)

	// Make buffer
	ptr := AVMalloc(int64(size))
	if ptr == nil {
		return nil
	}

	// Make call and wrap in a AVIOContextEx with callbacks
	ctx := (*C.AVIOContext)(C.avio_alloc_context_((*C.uchar)(ptr), C.int(size), (C.int)(boolToInt(writeable)), userInfo, C.int(r), C.int(w), C.int(s)))
	if ctx == nil {
		AVFree(ptr)
		return nil
	}

	// Set object
	ctxex.AVIOContext = (*AVIOContext)(ctx)
	ctxex.reader = reader
	ctxex.writer = writer
	ctxex.seeker = seeker
	ctxex.buf = ptr

	// Return success
	return ctxex
}

func (this *AVIOContext) Close() error {
	ctx := (*C.AVIOContext)(this)
	if err := AVError(C.avio_close(ctx)); err != 0 {
		return err
	} else {
		return nil
	}
}

func (this *AVIOContextEx) Free() {
	ctx := (*C.AVIOContext)(this.AVIOContext)
	AVFree(unsafe.Pointer(ctx.buffer))
	this.AVIOContext.Free()
}

func (this *AVIOContext) Free() {
	ctx := (*C.AVIOContext)(this)
	C.avio_context_free(&ctx)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *AVIOContext) String() string {
	str := "<aviocontext"
	if this.EOF() {
		str += " eof"
	}
	if this.Writeable() {
		str += " writable"
	}
	if this.Direct() {
		str += " direct"
	}
	if pos := this.Pos(); pos >= 0 {
		str += fmt.Sprint(" pos=", pos)
	}
	if size := this.Size(); size > 0 {
		str += fmt.Sprint(" buffer_size=", size)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *AVIOContext) Size() int {
	ctx := (*C.AVIOContext)(this)
	return int(ctx.buffer_size)
}

func (this *AVIOContext) Bytes() []byte {
	ctx := (*C.AVIOContext)(this)
	return cByteSlice(unsafe.Pointer(ctx.buffer), ctx.buffer_size)
}

func (this *AVIOContext) Pos() int64 {
	ctx := (*C.AVIOContext)(this)
	return int64(ctx.pos)
}

func (this *AVIOContext) EOF() bool {
	ctx := (*C.AVIOContext)(this)
	return intToBool(int(ctx.eof_reached))
}

func (this *AVIOContext) Writeable() bool {
	ctx := (*C.AVIOContext)(this)
	return intToBool(int(ctx.write_flag))
}

func (this *AVIOContext) Seekable() bool {
	ctx := (*C.AVIOContext)(this)
	return intToBool(int(ctx.seekable))
}

func (this *AVIOContext) Direct() bool {
	ctx := (*C.AVIOContext)(this)
	return intToBool(int(ctx.direct))
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (this *AVIOContext) Flush() {
	ctx := (*C.AVIOContext)(this)
	C.avio_flush(ctx)
}

func (this *AVIOContext) Read(buf []byte) (int, error) {
	ctx := (*C.AVIOContext)(this)
	data := unsafe.Pointer(nil)
	if buf != nil {
		data = unsafe.Pointer(&buf[0])
	}
	size := len(buf)
	if ret := C.avio_read(ctx, (*C.uint8_t)(data), C.int(size)); ret >= 0 {
		return int(ret), nil
	} else {
		return -1, AVError(ret)
	}
}

func (this *AVIOContext) Write(buf []byte) {
	ctx := (*C.AVIOContext)(this)
	data := unsafe.Pointer(nil)
	if buf != nil {
		data = unsafe.Pointer(&buf[0])
	}
	size := len(buf)
	C.avio_write(ctx, (*C.uint8_t)(data), C.int(size))
}

////////////////////////////////////////////////////////////////////////////////
// CALLBACKS

//export avio_read_callback
func avio_read_callback(userInfo unsafe.Pointer, buf *C.uint8_t, size C.int) C.int {
	ctx := (*AVIOContextEx)(userInfo)
	if ctx == nil || ctx.reader == nil {
		return C.int(AVERROR_EOF)
	}
	n, err := ctx.reader(cByteSlice(unsafe.Pointer(buf), size))
	if err != nil {
		// TODO: Check for ErrNO
		return C.int(AVERROR_EOF)
	} else {
		return C.int(n)
	}
}

//export avio_write_callback
func avio_write_callback(userInfo unsafe.Pointer, buf *C.uint8_t, size C.int) C.int {
	ctx := (*AVIOContextEx)(userInfo)
	if ctx == nil || ctx.writer == nil {
		return C.int(AVERROR_EOF)
	}
	n, err := ctx.writer(cByteSlice(unsafe.Pointer(buf), size))
	if err != nil {
		// TODO: Check for ErrNO
		return C.int(AVERROR_EOF)
	} else {
		return C.int(n)
	}
}

//export avio_seek_callback
func avio_seek_callback(userInfo unsafe.Pointer, offset C.int64_t, whence C.int) C.int64_t {
	ctx := (*AVIOContextEx)(userInfo)
	if ctx == nil || ctx.seeker == nil {
		return C.int64_t(syscall.EINVAL)
	}
	n, err := ctx.seeker(int64(offset), int(whence))
	if err != nil {
		if err, ok := err.(*fs.PathError); ok {
			if err, ok := err.Err.(syscall.Errno); ok {
				return C.int64_t(err)
			}
		}
		if syserr, ok := err.(syscall.Errno); ok {
			return C.int64_t(syserr)
		} else if err == io.EOF {
			return C.int64_t(AVERROR_EOF)
		} else {
			return C.int64_t(syscall.EINVAL)
		}
	} else {
		return C.int64_t(n)
	}
}
