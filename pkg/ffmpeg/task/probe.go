package task

import (
	"context"
	"strings"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Probe a media file or stream and return information about its format and streams
func (m *Manager) Probe(_ context.Context, req *schema.ProbeRequest) (*schema.ProbeResponse, error) {
	// The context is unused in this implementation
	// Open the file
	var reader *ffmpeg.Reader
	var err error
	opt := ffmpeg.WithInput(req.InputFormat, req.InputOpts...)
	if req.Reader != nil {
		reader, err = ffmpeg.NewReader(req.Reader, opt)
	} else {
		// Parse URL to support device:// scheme
		reader, err = OpenReaderFromURL(req.Input, opt)
	}
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Get format info
	var formatName, formatDesc string
	var mimeTypes []string
	if inputFormat := reader.InputFormat(); inputFormat != nil {
		formatName = inputFormat.Name()
		formatDesc = inputFormat.LongName()
		if mt := inputFormat.MimeTypes(); mt != "" {
			mimeTypes = strings.Split(mt, ",")
		}
	}

	// Get streams
	avStreams := reader.AVStreams()
	streams := make([]*schema.Stream, 0, len(avStreams))
	for _, avStream := range avStreams {
		if s := schema.NewStream(avStream); s != nil {
			streams = append(streams, s)
		}
	}

	// Build response
	resp := &schema.ProbeResponse{
		Format:      formatName,
		Description: formatDesc,
		MimeTypes:   mimeTypes,
		Duration:    reader.Duration().Seconds(),
		Streams:     streams,
	}

	return resp, nil
}
