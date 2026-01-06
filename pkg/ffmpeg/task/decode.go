package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Decode decodes media from a file or reader. If the writer implements FrameWriter,
// frames are passed directly to WriteFrame(). Otherwise, frame metadata is written
// as JSON to the writer.
//
// The context can be used to cancel decoding.
func (m *Manager) Decode(ctx context.Context, w io.Writer, req *schema.DecodeRequest) error {
	if req == nil {
		return errors.New("nil decode request")
	}
	if w == nil {
		return errors.New("nil writer")
	}
	if req.Input == "" && req.Reader == nil {
		return errors.New("either Input or Reader must be provided")
	}

	// Open the input reader
	var reader *ffmpeg.Reader
	var err error
	opt := ffmpeg.WithInput(req.InputFormat, req.InputOpts...)
	if req.Reader != nil {
		reader, err = ffmpeg.NewReader(req.Reader, opt)
	} else {
		reader, err = ffmpeg.Open(req.Input, opt)
	}
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}
	defer reader.Close()

	// Check if writer supports frame writing
	frameWriter, hasFrameWriter := w.(schema.FrameWriter)

	// Create the map function to decode all streams
	mapfn := func(streamIndex int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		// Decode all streams - return the parameters unchanged (no resampling/resizing)
		return par, nil
	}

	// Decode frames
	err = reader.Demux(ctx, mapfn, func(streamIndex int, frame *ffmpeg.Frame) error {
		// Write frame as JSON
		data, err := json.Marshal(frame)
		if err != nil {
			return fmt.Errorf("marshal frame: %w", err)
		} else if _, err := w.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("write frame: %w", err)
		}

		// Also call WriteFrame if the writer supports it
		if hasFrameWriter {
			if err := frameWriter.WriteFrame(streamIndex, frame); err != nil {
				return err
			}
		}

		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("demux: %w", err)
	}

	return nil
}
