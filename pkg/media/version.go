package media

import (
	"fmt"
	"io"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
)

func PrintVersion(w io.Writer) {
	PrintVersionLib(w, "libavutil", ffmpeg.AVUtil_version())
	PrintVersionLib(w, "libavformat", ffmpeg.AVFormat_version())
	PrintVersionLib(w, "libavcodec", ffmpeg.AVCodec_version())
	PrintVersionLib(w, "libavdevice", ffmpeg.AVDevice_version())
}

func PrintVersionLib(w io.Writer, name string, version uint) {
	fmt.Printf("  %-10s %d.%d.%d\n", name+":", version>>16, version>>8&0xFF, version&0xFF)
}
