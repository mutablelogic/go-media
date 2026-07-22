package schema

import (
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Format struct {
	Name        string   `json:"name"`                  // Codec name, e.g. "aac", "libmp3lame", "copy", ...
	Description string   `json:"description,omitempty"` // Codec description
	Type        string   `json:"type,omitempty"`        // Content type
	Ext         string   `json:"ext,omitempty"`         // File extensions
	Audio       []string `json:"audio,omitempty"`       // Audio codecs supported by this format, first one is the default
	Video       []string `json:"video,omitempty"`       // Video codecs supported by this format, first one is the default
	Subtitle    []string `json:"subtitle,omitempty"`    // Subtitle codecs supported by this format, first one is the default
	Opts        []Option `json:"opts,omitempty"`        // Codec options
}

type FormatListRequest struct {
	pg.OffsetLimit
}

type FormatList struct {
	FormatListRequest
	Count uint64    `json:"count"` // Number of formats
	Body  []*Format `json:"body"`  // List of formats
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
		Audio:       audioCodecs,
		Video:       videoCodecs,
		Subtitle:    subtitleCodecs,
		Opts:        OptionsForFormat(format),
	}
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
	consts := make(map[string][]Option, len(ffopts))
	result := make([]Option, 0, len(ffopts))
	for _, opt := range ffopts {
		if opt == nil {
			continue
		}
		if opt.Type() == ff.AV_OPT_TYPE_CONST {
			key := opt.Unit()
			consts[key] = append(consts[key], NewOption(opt))
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
