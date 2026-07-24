package schema

import (
	"net/url"
	"strconv"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Format struct {
	Name        string   `json:"name" help:"Format name." example:"mp4"`
	Description string   `json:"description,omitempty" help:"Human-readable format description." example:"MP4 (MPEG-4 Part 14)"`
	Type        string   `json:"type,omitempty" help:"MIME content type." example:"video/mp4"`
	Ext         string   `json:"ext,omitempty" help:"File extensions associated with this format." example:"mp4,m4a,m4v"`
	Audio       []string `json:"audio,omitempty" help:"Audio codecs supported by this format; the first is the default." example:"[\"aac\"]"`
	Video       []string `json:"video,omitempty" help:"Video codecs supported by this format; the first is the default." example:"[\"h264\"]"`
	Subtitle    []string `json:"subtitle,omitempty" help:"Subtitle codecs supported by this format; the first is the default." example:"[\"mov_text\"]"`
	Opts        []Option `json:"opts,omitempty" help:"Format-specific options."`
}

type FormatListRequest struct {
	Name *string `json:"name,omitempty" help:"Filter by format name." placeholder:"mp4" example:"mp4"`
	Type *string `json:"type,omitempty" help:"Filter by MIME content type." placeholder:"video/mp4" example:"video/mp4"`
	Ext  *string `json:"ext,omitempty" help:"Filter by file extension." placeholder:"mp4" example:"mp4"`
	pg.OffsetLimit
}

type FormatList struct {
	FormatListRequest
	Count uint64    `json:"count" help:"Number of formats." example:"1"`
	Body  []*Format `json:"body" help:"List of formats."`
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r FormatListRequest) Query() url.Values {
	query := url.Values{}
	if r.Name != nil {
		query.Set("name", types.Value(r.Name))
	}
	if r.Type != nil {
		query.Set("type", types.Value(r.Type))
	}
	if r.Ext != nil {
		query.Set("ext", types.Value(r.Ext))
	}
	if r.Offset > 0 {
		query.Set("offset", strconv.FormatUint(r.Offset, 10))
	}
	if r.Limit != nil {
		query.Set("limit", strconv.FormatUint(types.Value(r.Limit), 10))
	}
	return query
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewOutputFormat(format *ff.AVOutputFormat) *Format {
	if format == nil {
		return nil
	}

	// Iterate over all registered encoders to find the audio, video, and subtitle
	// codecs supported by this format.
	var audioCodecs []string
	var videoCodecs []string
	var subtitleCodecs []string
	var opaque uintptr
	for {
		codec := ff.AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}
		if !ff.AVCodec_is_encoder(codec) {
			continue
		}
		if ff.AVFormat_query_codec(format, codec.ID(), ff.FF_COMPLIANCE_NORMAL) != 1 {
			continue
		}
		switch codec.Type() {
		case ff.AVMEDIA_TYPE_AUDIO:
			audioCodecs = append(audioCodecs, codec.Name())
		case ff.AVMEDIA_TYPE_VIDEO:
			videoCodecs = append(videoCodecs, codec.Name())
		case ff.AVMEDIA_TYPE_SUBTITLE:
			subtitleCodecs = append(subtitleCodecs, codec.Name())
		}
	}

	return &Format{
		Name:        format.Name(),
		Description: format.LongName(),
		Type:        format.MimeTypes(),
		Ext:         format.Extensions(),
		Audio:       withDefaultFirst(audioCodecs, format.AudioCodec()),
		Video:       withDefaultFirst(videoCodecs, format.VideoCodec()),
		Subtitle:    withDefaultFirst(subtitleCodecs, format.SubtitleCodec()),
		Opts:        OptionsForFormat(format),
	}
}

// withDefaultFirst reorders codecs so the format's actual default codec (as
// declared by FFmpeg itself, via AVOutputFormat's audio_codec/video_codec/
// subtitle_codec) is first - codecs is otherwise in whatever order iterating
// the codec registry happened to produce, which has no relationship to which
// one the format actually prefers.
func withDefaultFirst(codecs []string, id ff.AVCodecID) []string {
	if id == ff.AV_CODEC_ID_NONE {
		return codecs
	}
	encoder := ff.AVCodec_find_encoder(id)
	if encoder == nil {
		return codecs
	}
	name := encoder.Name()
	for i, codec := range codecs {
		if codec != name {
			continue
		}
		if i == 0 {
			return codecs
		}
		out := make([]string, 0, len(codecs))
		out = append(out, name)
		out = append(out, codecs[:i]...)
		return append(out, codecs[i+1:]...)
	}
	return codecs
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Format) String() string {
	return types.Stringify(r)
}

func (r FormatList) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func OptionsForFormat(format *ff.AVOutputFormat) []Option {
	if format == nil {
		return nil
	}

	// Prefer default bitrate from the codec private class if available.
	class := format.PrivClass()
	if class == nil {
		return nil
	}

	// Extract options
	ffopts := ff.AVUtil_opt_list_from_class(class)
	consts := make(map[string][]OptionConst, len(ffopts))
	result := make([]Option, 0, len(ffopts))
	for _, opt := range ffopts {
		if opt == nil {
			continue
		}
		if opt.Type() == ff.AV_OPT_TYPE_CONST {
			key := opt.Unit()
			consts[key] = append(consts[key], NewOptionConst(opt))
			continue
		}
		name := strings.TrimSpace(opt.Name())
		if name == "" {
			continue
		}
		result = append(result, NewOption(opt))
	}

	// Append the constants to the options
	for i, opt := range result {
		consts, exists := consts[opt.Name]
		if exists && len(consts) > 0 {
			result[i].Const = consts
		}
	}

	// Return the result
	return result
}
