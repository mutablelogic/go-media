package version

import (
	"fmt"
	"runtime"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

var (
	GitSource   string
	GitTag      string
	GitBranch   string
	GoBuildTime string
)

// Return version information as a set of metadata
func Version() []*ffmpeg.Metadata {
	metadata := []*ffmpeg.Metadata{
		ffmpeg.NewMetadata("libavcodec_version", ffVersionAsString(ff.AVCodec_version())),
		ffmpeg.NewMetadata("libavformat_version", ffVersionAsString(ff.AVFormat_version())),
		ffmpeg.NewMetadata("libavutil_version", ffVersionAsString(ff.AVUtil_version())),
		ffmpeg.NewMetadata("libavdevice_version", ffVersionAsString(ff.AVDevice_version())),
		//		newMetadata("libavfilter_version", ff.AVFilter_version()),
		ffmpeg.NewMetadata("libswscale_version", ffVersionAsString(ff.SWScale_version())),
		ffmpeg.NewMetadata("libswresample_version", ffVersionAsString(ff.SWResample_version())),
	}
	if GitSource != "" {
		metadata = append(metadata, ffmpeg.NewMetadata("git_source", GitSource))
	}
	if GitBranch != "" {
		metadata = append(metadata, ffmpeg.NewMetadata("git_branch", GitBranch))
	}
	if GitTag != "" {
		metadata = append(metadata, ffmpeg.NewMetadata("git_tag", GitTag))
	}
	if GoBuildTime != "" {
		metadata = append(metadata, ffmpeg.NewMetadata("go_build_time", GoBuildTime))
	}
	if runtime.Version() != "" {
		metadata = append(metadata, ffmpeg.NewMetadata("go_version", runtime.Version()))
		metadata = append(metadata, ffmpeg.NewMetadata("go_arch", runtime.GOOS+"/"+runtime.GOARCH))
	}
	return metadata
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func ffVersionAsString(version uint) string {
	return fmt.Sprintf("%d.%d.%d", version&0xFF0000>>16, version&0xFF00>>8, version&0xFF)
}
