package reader

import (
	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Protocols returns the name of every I/O protocol FFmpeg has registered for
// reading (e.g. "file", "http", "https", "rtmp", "udp", "tcp", ...) - the
// URL scheme a Reader's url actually travels over, which is a different
// axis from a container format: e.g. "rtsp://host/stream" is demuxed by the
// "rtsp" input format, but reaches the network via the "tcp"/"udp"
// protocol - "rtsp" itself is typically not in this list. Useful for
// validating a URL's scheme before passing it to Open/NewReader.
func Protocols() []string {
	var opaque uintptr
	var result []string
	for {
		name := ff.AVFormat_avio_enum_protocols(&opaque, false)
		if name == "" {
			break
		}
		result = append(result, name)
	}
	return result
}
