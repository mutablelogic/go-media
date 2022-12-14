package media

import (
	"fmt"
	"io"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
)

func PrintVersion(w io.Writer) {
	printVersionLib(w, "libavutil", ffmpeg.AVUtil_version())
	printVersionLib(w, "libavformat", ffmpeg.AVFormat_version())
	printVersionLib(w, "libavcodec", ffmpeg.AVCodec_version())
	printVersionLib(w, "libavdevice", ffmpeg.AVDevice_version())
	printVersionLib(w, "libswresample", ffmpeg.SWR_version())
	printVersionLib(w, "libswscale", ffmpeg.SWS_version())
}

func printVersionLib(w io.Writer, name string, version uint) {
	fmt.Printf("  %-10s %d.%d.%d\n", name+":", version>>16, version>>8&0xFF, version&0xFF)
}
