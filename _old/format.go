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

type formatmeta struct {
	Name        string    `json:"name" writer:",width:25"`
	Description string    `json:"description" writer:",wrap,width:40"`
	Extensions  string    `json:"extensions,omitempty"`
	MimeTypes   string    `json:"mimetypes,omitempty" writer:",wrap,width:40"`
	MediaType   MediaType `json:"type,omitempty" writer:",wrap,width:21"`
}

type inputformat struct {
	formatmeta
	ctx *ff.AVInputFormat
}

type outputformat struct {
	formatmeta
	ctx *ff.AVOutputFormat
}

var _ Format = (*inputformat)(nil)
var _ Format = (*outputformat)(nil)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newInputFormat(ctx *ff.AVInputFormat, t MediaType) *inputformat {
	v := &inputformat{ctx: ctx}
	v.formatmeta.Name = strings.Join(v.Name(), " ")
	v.formatmeta.Description = v.Description()
	v.formatmeta.Extensions = strings.Join(v.Extensions(), " ")
	v.formatmeta.MimeTypes = strings.Join(v.MimeTypes(), " ")
	v.formatmeta.MediaType = INPUT | t
	return v
}

func newOutputFormat(ctx *ff.AVOutputFormat, t MediaType) *outputformat {
	v := &outputformat{ctx: ctx}
	v.formatmeta.Name = strings.Join(v.Name(), " ")
	v.formatmeta.Description = v.Description()
	v.formatmeta.Extensions = strings.Join(v.Extensions(), " ")
	v.formatmeta.MimeTypes = strings.Join(v.MimeTypes(), " ")
	v.formatmeta.MediaType = OUTPUT | t
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
	return v.MediaType
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
	return v.MediaType
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func matchesInput(demuxer *ff.AVInputFormat, mimetype ...string) bool {
	// TODO: media_type

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
	// TODO: media_type

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
