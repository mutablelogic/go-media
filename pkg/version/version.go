package version

import (
	"encoding/binary"
	"fmt"
	"runtime"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

var (
	GitSource   string
	GitTag      string
	GitBranch   string
	GoBuildTime string
)

type Metadata struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Return version information as a set of metadata
func Version() []Metadata {
	metadata := []Metadata{
		{"libavcodec_version", ffVersionAsString(ff.AVCodec_version())},
		{"libavformat_version", ffVersionAsString(ff.AVFormat_version())},
		{"libavutil_version", ffVersionAsString(ff.AVUtil_version())},
		{"libavdevice_version", ffVersionAsString(ff.AVDevice_version())},
		//		Version{"libavfilter_version", ff.AVFilter_version()},
		{"libswscale_version", ffVersionAsString(ff.SWScale_version())},
		{"libswresample_version", ffVersionAsString(ff.SWResample_version())},
	}
	if GitSource != "" {
		metadata = append(metadata, Metadata{"git_source", GitSource})
	}
	if GitBranch != "" {
		metadata = append(metadata, Metadata{"git_branch", GitBranch})
	}
	if GitTag != "" {
		metadata = append(metadata, Metadata{"git_tag", GitTag})
	}
	if GoBuildTime != "" {
		metadata = append(metadata, Metadata{"go_build_time", GoBuildTime})
	}
	if runtime.Version() != "" {
		metadata = append(metadata, Metadata{"go_version", runtime.Version()})
		metadata = append(metadata, Metadata{"os_arch", runtime.GOOS + "/" + runtime.GOARCH})
	}
	switch NativeEndian {
	case binary.LittleEndian:
		metadata = append(metadata, Metadata{"os_endian", "little"})
	case binary.BigEndian:
		metadata = append(metadata, Metadata{"os_endian", "big"})
	}
	return metadata
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func ffVersionAsString(version uint) string {
	return fmt.Sprintf("%d.%d.%d", version&0xFF0000>>16, version&0xFF00>>8, version&0xFF)
}
