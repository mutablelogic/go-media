package media

import (
	"encoding/json"
	"slices"
	"strings"

	// Package imports
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

type manager struct {
}

type formatmeta struct {
	Name        string `json:"name" writer:",width:25"`
	Description string `json:"description" writer:",wrap,width:40"`
	Extensions  string `json:"extensions,omitempty"`
	MimeTypes   string `json:"mimetypes,omitempty"`
}

type inputformat struct {
	formatmeta
	ctx *ff.AVInputFormat
}

type outputformat struct {
	formatmeta
	ctx *ff.AVOutputFormat
}

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewManager() *manager {
	return new(manager)
}

func newInputFormat(ctx *ff.AVInputFormat) *inputformat {
	return &inputformat{
		ctx: ctx,
		formatmeta: formatmeta{
			Name:        ctx.Name(),
			Description: ctx.LongName(),
			Extensions:  ctx.Extensions(),
			MimeTypes:   ctx.MimeTypes(),
		},
	}
}

func newOutputFormat(ctx *ff.AVOutputFormat) *outputformat {
	return &outputformat{
		ctx: ctx,
		formatmeta: formatmeta{
			Name:        ctx.Name(),
			Description: ctx.LongName(),
			Extensions:  ctx.Extensions(),
			MimeTypes:   ctx.MimeTypes(),
		},
	}
}

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v inputformat) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ctx)
}

func (v outputformat) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ctx)
}

func (v inputformat) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

func (v outputformat) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the list of matching input formats, optionally filtering by name,
// extension or mimetype File extensions should be prefixed with a dot,
// e.g. ".mp4"
func (manager *manager) InputFormats(filter ...string) []InputFormat {
	var result []InputFormat

	// Iterate over all input formats
	var opaque uintptr
	for {
		demuxer := ff.AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}
		if len(filter) == 0 {
			result = append(result, newInputFormat(demuxer))
		} else if manager.matchesInput(demuxer, filter...) {
			result = append(result, newInputFormat(demuxer))
		}
	}

	// Return success
	return result
}

// Return the list of matching output formats, optionally filtering by name,
// extension or mimetype File extensions should be prefixed with a dot,
// e.g. ".mp4"
func (manager *manager) OutputFormats(filter ...string) []OutputFormat {
	var result []OutputFormat

	// Iterate over all output formats
	var opaque uintptr
	for {
		muxer := ff.AVFormat_muxer_iterate(&opaque)
		if muxer == nil {
			break
		}
		if len(filter) == 0 {
			result = append(result, newOutputFormat(muxer))
		} else if manager.matchesOutput(muxer, filter...) {
			result = append(result, newOutputFormat(muxer))
		}
	}

	// Return success
	return result
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *manager) matchesInput(demuxer *ff.AVInputFormat, mimetype ...string) bool {
	// Match any
	if len(mimetype) == 0 {
		return true
	}
	// Match mimetype
	for _, mimetype := range mimetype {
		mimetype = strings.ToLower(strings.TrimSpace(mimetype))
		if slices.Contains(strings.Split(demuxer.Name(), ","), mimetype) {
			return true
		}
		if strings.HasPrefix(mimetype, ".") {
			ext := strings.TrimPrefix(mimetype, ".")
			if slices.Contains(strings.Split(demuxer.Extensions(), ","), ext) {
				return true
			}
		}
		if slices.Contains(strings.Split(demuxer.MimeTypes(), ","), mimetype) {
			return true
		}
	}
	// No match
	return false
}

func (this *manager) matchesOutput(muxer *ff.AVOutputFormat, mimetype ...string) bool {
	// Match any
	if len(mimetype) == 0 {
		return true
	}
	// Match mimetype
	for _, mimetype := range mimetype {
		mimetype = strings.ToLower(strings.TrimSpace(mimetype))
		if slices.Contains(strings.Split(muxer.Name(), ","), mimetype) {
			return true
		}
		if strings.HasPrefix(mimetype, ".") {
			ext := strings.TrimPrefix(mimetype, ".")
			if slices.Contains(strings.Split(muxer.Extensions(), ","), ext) {
				return true
			}
		}
		if slices.Contains(strings.Split(muxer.MimeTypes(), ","), mimetype) {
			return true
		}
	}
	// No match
	return false
}
