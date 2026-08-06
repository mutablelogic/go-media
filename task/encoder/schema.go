package encoder

import (
	"io"

	// Packages
	profile "github.com/mutablelogic/go-media/profile/schema"
	task "github.com/mutablelogic/go-media/task/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type EncodeRequest struct {
	Reader io.Reader `json:"-"`

	// Output is the output container format and any format-level options
	// (e.g. muxer flags), e.g. profile.OutputWithName("mp4").
	Output *profile.Output `json:"output" help:"Output container format." required:"true"`

	// Audio, Video and Subtitle are the encoding targets for this request.
	// Each is optional, but at least one must be set; when set, the first
	// stream of that type found in the input is transcoded into it.
	Audio    *profile.AudioProfile    `json:"audio,omitempty" help:"Audio encoding profile; if set, the first audio stream found in the input is transcoded into it."`
	Video    *profile.VideoProfile    `json:"video,omitempty" help:"Video encoding profile; if set, the first video stream found in the input is transcoded into it."`
	Subtitle *profile.SubtitleProfile `json:"subtitle,omitempty" help:"Subtitle encoding profile; if set, the first subtitle stream found in the input is transcoded into it."`
}

type EncodeResponse struct {
	// Path is the location of the encoded output.
	Path string `json:"path" help:"Path to the encoded output file." example:"/tmp/gomedia-encode-123456.mp4"`
}

var _ task.Task = (*EncodeRequest)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - TASK

// Task identifies this task's kind, for Status.Task.
func (r *EncodeRequest) Task() string {
	return "encoder"
}
