package task

import (
	"context"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all supported codecs
func (manager *Manager) ListCodecs(_ context.Context, req *schema.ListCodecRequest) (schema.ListCodecResponse, error) {
	var opaque uintptr
	result := make(schema.ListCodecResponse, 0, 512)

	// Filter function
	matches := func(req *schema.ListCodecRequest, c *schema.Codec) bool {
		if req == nil {
			return true
		}
		if req.Name != "" && !strings.Contains(c.AVCodec.Name(), req.Name) {
			return false
		}
		if req.Type != "" {
			var typeStr string
			switch c.AVCodec.Type() {
			case ff.AVMEDIA_TYPE_VIDEO:
				typeStr = "video"
			case ff.AVMEDIA_TYPE_AUDIO:
				typeStr = "audio"
			case ff.AVMEDIA_TYPE_SUBTITLE:
				typeStr = "subtitle"
			case ff.AVMEDIA_TYPE_DATA:
				typeStr = "data"
			case ff.AVMEDIA_TYPE_ATTACHMENT:
				typeStr = "attachment"
			default:
				typeStr = "unknown"
			}
			if typeStr != req.Type {
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
		if c := schema.NewCodec(codec); c != nil {
			if matches(req, c) {
				result = append(result, *c)
			}
		}
	}
	return result, nil
}
