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
	MimeTypes   string `json:"mimetypes,omitempty" writer:",wrap,width:40"`
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
	v := &inputformat{ctx: ctx}
	v.formatmeta.Name = strings.Join(v.Name(), " ")
	v.formatmeta.Description = v.Description()
	v.formatmeta.Extensions = strings.Join(v.Extensions(), " ")
	v.formatmeta.MimeTypes = strings.Join(v.MimeTypes(), " ")
	return v
}

func newOutputFormat(ctx *ff.AVOutputFormat) *outputformat {
	v := &outputformat{ctx: ctx}
	v.formatmeta.Name = strings.Join(v.Name(), " ")
	v.formatmeta.Description = v.Description()
	v.formatmeta.Extensions = strings.Join(v.Extensions(), " ")
	v.formatmeta.MimeTypes = strings.Join(v.MimeTypes(), " ")
	return v
}

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v *inputformat) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ctx)
}

func (v *outputformat) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ctx)
}

func (v *inputformat) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

func (v *outputformat) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the list of matching input formats, optionally filtering by name,
// extension or mimetype File extensions should be prefixed with a dot,
// e.g. ".mp4"
func (manager *manager) InputFormats(filter ...string) []Format {
	var result []Format

	// Iterate over all input formats
	var opaque uintptr
	for {
		demuxer := ff.AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}
		if matchesInput(demuxer, filter...) {
			result = append(result, newInputFormat(demuxer))
		}
	}

	// Return success
	return result
}

// Return the list of matching output formats, optionally filtering by name,
// extension or mimetype File extensions should be prefixed with a dot,
// e.g. ".mp4"
func (manager *manager) OutputFormats(filter ...string) []Format {
	var result []Format

	// Iterate over all output formats
	var opaque uintptr
	for {
		muxer := ff.AVFormat_muxer_iterate(&opaque)
		if muxer == nil {
			break
		}
		if matchesOutput(muxer, filter...) {
			result = append(result, newOutputFormat(muxer))
		}
	}

	// Return success
	return result
}

func (v *inputformat) Name() []string {
	return strings.Split(v.ctx.Name(), ",")
}

func (v *inputformat) Description() string {
	return v.ctx.LongName()
}

func (v *inputformat) Extensions() []string {
	result := []string{}
	for _, ext := range strings.Split(v.ctx.Extensions(), ",") {
		ext = strings.TrimSpace(ext)
		if ext != "" {
			result = append(result, "."+ext)
		}
	}
	return result
}

func (v *inputformat) MimeTypes() []string {
	result := []string{}
	for _, mimetype := range strings.Split(v.ctx.MimeTypes(), ",") {
		if mimetype != "" {
			result = append(result, mimetype)
		}
	}
	return result
}

func (v *inputformat) Type() MediaType {
	return INPUT
}

func (v *outputformat) Name() []string {
	return strings.Split(v.ctx.Name(), ",")
}

func (v *outputformat) Description() string {
	return v.ctx.LongName()
}

func (v *outputformat) Extensions() []string {
	result := []string{}
	for _, ext := range strings.Split(v.ctx.Extensions(), ",") {
		ext = strings.TrimSpace(ext)
		if ext != "" {
			result = append(result, "."+ext)
		}
	}
	return result
}

func (v *outputformat) MimeTypes() []string {
	result := []string{}
	for _, mimetype := range strings.Split(v.ctx.MimeTypes(), ",") {
		if mimetype != "" {
			result = append(result, mimetype)
		}
	}
	return result
}

func (v *outputformat) Type() MediaType {
	return OUTPUT
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func matchesInput(demuxer *ff.AVInputFormat, mimetype ...string) bool {
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

func matchesOutput(muxer *ff.AVOutputFormat, filter ...string) bool {
	// Match any
	if len(filter) == 0 {
		return true
	}
	// Match mimetype
	for _, filter := range filter {
		if filter == "" {
			continue
		}
		filter = strings.ToLower(strings.TrimSpace(filter))
		if slices.Contains(strings.Split(muxer.Name(), ","), filter) {
			return true
		}
		if strings.HasPrefix(filter, ".") {
			if slices.Contains(strings.Split(muxer.Extensions(), ","), filter[1:]) {
				return true
			}
		}
		mt := strings.Split(muxer.MimeTypes(), ",")
		if slices.Contains(mt, filter) {
			return true
		}
	}
	// No match
	return false
}
