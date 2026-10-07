package manager

import (
	"context"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListCodecs returns all supported codecs, with optional filters.
func (m *Media) ListCodecs(_ context.Context, req schema.ListCodecRequest) (schema.ListCodecResponse, error) {
	var opaque uintptr
	result := make(schema.ListCodecResponse, 0, 512)

	matches := func(c *schema.Codec) bool {
		if req.Name != "" && !strings.Contains(c.AVCodec.Name(), req.Name) {
			return false
		}
		if req.Type != "" {
			mt := schema.MediaType(c.AVCodec.Type())
			if mt.String() != req.Type {
				return false
			}
		}
		if req.IsEncoder != nil {
			if *req.IsEncoder && !ff.AVCodec_is_encoder(c.AVCodec) {
				return false
			}
			if !*req.IsEncoder && !ff.AVCodec_is_decoder(c.AVCodec) {
				return false
			}
		}
		return true
	}

	for {
		codec := ff.AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}
		if c := schema.NewCodec(codec); c != nil && matches(c) {
			result = append(result, *c)
		}
	}

	return result, nil
}
