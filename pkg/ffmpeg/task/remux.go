package task

import (
	"context"
	"fmt"
	"io"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Remux remuxes media streams from input to output without re-encoding.
// The writer must implement io.Writer and can optionally implement
// schema.Writer for enhanced feedback (progress and logging)
func (m *Manager) Remux(ctx context.Context, w io.Writer, req *schema.RemuxRequest) (*schema.RemuxResponse, error) {
	// Open the input reader
	var reader *ffmpeg.Reader
	var err error
	if req.Reader != nil {
		if req.Input != "" {
			reader, err = ffmpeg.NewReader(req.Reader, ffmpeg.WithInput(req.Input))
		} else {
			reader, err = ffmpeg.NewReader(req.Reader)
		}
	} else {
		reader, err = ffmpeg.Open(req.Input)
	}
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Decode and print packets
	var packetCount int
	err = reader.Decode(ctx, func(stream int, pkt *ffmpeg.Packet) error {
		packetCount++
		fmt.Fprintf(w, "Packet %d: stream=%d size=%d pts=%d dts=%d\n",
			packetCount, stream, pkt.Size(), pkt.Pts(), pkt.Dts())

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Build minimal response
	resp := &schema.RemuxResponse{
		Duration: reader.Duration().Seconds(),
		Size:     int64(packetCount),
	}

	if inputFormat := reader.InputFormat(); inputFormat != nil {
		resp.Format = inputFormat.Name()
	}

	return resp, nil
}
