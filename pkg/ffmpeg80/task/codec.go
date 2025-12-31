package task

import (
	"context"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg80/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all supported codecs
func (manager *Manager) ListCodec(_ context.Context, req *schema.ListCodecRequest) (schema.ListCodecResponse, error) {
	var opaque uintptr
	result := make(schema.ListCodecResponse, 0, 512)

	// Filter function
	matches := func(req *schema.ListCodecRequest, c *schema.Codec) bool {
		if req == nil {
			return true
		}
		if req.Name != "" && !strings.Contains(c.Name, req.Name) {
			return false
		}
		if req.Type != "" && c.Type != req.Type {
			return false
		}
		if req.IsEncoder != nil {
			if *req.IsEncoder && !c.IsEncoder {
				return false
			}
			if !*req.IsEncoder && !c.IsDecoder {
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
		if c := schema.NewCodec(codec); c != nil {
			if matches(req, c) {
				result = append(result, *c)
			}
		}
	}
	return result, nil
}
