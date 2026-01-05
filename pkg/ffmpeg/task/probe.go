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
	if req.Reader != nil {
		if req.Path != "" {
			reader, err = ffmpeg.NewReader(req.Reader, ffmpeg.WithInput(req.Path))
		} else {
			reader, err = ffmpeg.NewReader(req.Reader)
		}
	} else {
		reader, err = ffmpeg.Open(req.Path)
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

	// Get metadata if requested
	if req.Metadata {
		metaEntries := reader.Metadata()
		if len(metaEntries) > 0 {
			resp.Metadata = make(map[string]string, len(metaEntries))
			for _, entry := range metaEntries {
				if entry.Key() != ffmpeg.MetaArtwork {
					resp.Metadata[entry.Key()] = entry.Value()
				}
			}
		}
	}

	// Get artwork if requested
	if req.Artwork {
		artworkEntries := reader.Metadata(ffmpeg.MetaArtwork)
		if len(artworkEntries) > 0 {
			resp.Artwork = make([]schema.Artwork, 0, len(artworkEntries))
			for _, entry := range artworkEntries {
				if data := entry.Bytes(); len(data) > 0 {
					resp.Artwork = append(resp.Artwork, schema.Artwork(data))
				}
			}
		}
	}

	return resp, nil
}
