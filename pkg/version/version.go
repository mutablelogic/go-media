package version

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"runtime"

	// Packages
	chromaprint "github.com/mutablelogic/go-media/sys/chromaprint"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
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
func Map() []Metadata {
	metadata := []Metadata{
		{"libavcodec_version", ffVersionAsString(ff.AVCodec_version())},
		{"libavcodec_configuration", ff.AVCodec_configuration()},
		{"libavdevice_version", ffVersionAsString(ff.AVDevice_version())},
		{"libavdevice_configuration", ff.AVDevice_configuration()},
		{"libavfilter_version", ffVersionAsString(ff.AVFilter_version())},
		{"libavfilter_configuration", ff.AVFilter_configuration()},
		{"libavformat_version", ffVersionAsString(ff.AVFormat_version())},
		{"libavformat_configuration", ff.AVFormat_configuration()},
		{"libavutil_version", ffVersionAsString(ff.AVUtil_version())},
		{"libavutil_configuration", ff.AVUtil_configuration()},
		{"libswscale_version", ffVersionAsString(ff.SWScale_version())},
		{"libswscale_configuration", ff.SWScale_configuration()},
		{"libswresample_version", ffVersionAsString(ff.SWResample_version())},
		{"libswresample_configuration", ff.SWResample_configuration()},
		{"chromaprint_version", chromaprint.Version()},
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
	if exec, err := os.Executable(); err == nil {
		metadata = append(metadata, Metadata{"executable_path", exec})
		metadata = append(metadata, Metadata{"executable_name", path.Base(exec)})
	}
	return metadata
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func ffVersionAsString(version uint) string {
	return fmt.Sprintf("%d.%d.%d", version&0xFF0000>>16, version&0xFF00>>8, version&0xFF)
}
