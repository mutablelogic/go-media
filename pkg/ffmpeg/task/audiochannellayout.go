package task

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Probe a media file or stream and return information about its format and streams
func (m *Manager) ListAudioChannelLayout(_ context.Context, req *schema.ListAudioChannelLayoutRequest) (schema.ListAudioChannelLayoutResponse, error) {
	var iter uintptr
	response := make(schema.ListAudioChannelLayoutResponse, 0, 32)

	// Filter function
	matches := func(req *schema.ListAudioChannelLayoutRequest, ch *ff.AVChannelLayout) bool {
		if req == nil {
			return true
		}
		if req.Name != "" {
			name, _ := ff.AVUtil_channel_layout_describe(ch)
			if name != req.Name {
				return false
			}
		}
		if req.NumChannels != 0 && ch.NumChannels() != req.NumChannels {
			return false
		}
		return true
	}

	// Iterate through standard channel layouts
	for {
		ch := ff.AVUtil_channel_layout_standard(&iter)
		if ch == nil {
			break
		}
		if !matches(req, ch) {
			continue
		}
		if layout := schema.NewAudioChannelLayout(ch); layout != nil {
			response = append(response, *layout)
		}
	}
	// Return the response
	return response, nil
}
