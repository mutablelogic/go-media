package task

import (
	"errors"
	"io"
	"net/url"
	"os"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
	writer "github.com/mutablelogic/go-media/writer"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

///////////////////////////////////////////////////////////////////////////////
// ENCODE TASK

type AudioEncodeMediaRequest struct {
	Reader io.Reader `json:"-"`
	ProbeRequestOpts
	Stream  *uint64              `json:"stream,omitempty" name:"stream" help:"Index of the stream to encode; 0-based." example:"0"`
	Profile profile.AudioProfile `json:"profile,omitempty" name:"profile" help:"Audio encoding profile to use."`
}

// AudioEncodeMediaResponse is the result of a successful AudioEncodeMediaRequest.
type AudioEncodeMediaResponse struct {
	// Path is the location of the encoded output. For now this is always a
	// temporary file, until a real upload/storage destination exists.
	Path string `json:"path" help:"Path to the encoded output file."`
}

var _ Task = (*AudioEncodeMediaRequest)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (task *AudioEncodeMediaRequest) Run(ctx Context) (err error) {
	_, endSpan := otel.StartSpan(ctx.Tracer, ctx, "AudioEncodeMedia",
		attribute.String("req", types.Stringify(task)),
	)
	defer func() { endSpan(err) }()

	// Check for a nil reader
	if task.Reader == nil {
		return gomedia.ErrBadParameter.With("nil reader")
	}

	// Stream to encode - defaults to the first stream
	streamID := 0
	if task.Stream != nil {
		streamID = int(*task.Stream)
	}

	// Open the input
	rd, err := reader.NewReader(task.Reader, reader.WithInput(task.Format, task.Opts...))
	if err != nil {
		return err
	}
	defer rd.Close()

	// Create the output - for now, a temporary file, until a real
	// upload/storage destination exists
	output := profile.OutputWithName("mp4")
	if output == nil {
		return gomedia.ErrInternalError.With("mp4 output format not available")
	}
	tmp, err := os.CreateTemp("", "gomedia-encode-*.mp4")
	if err != nil {
		return err
	}
	path := tmp.Name()
	if err := tmp.Close(); err != nil {
		return err
	}

	w, err := writer.Create(&url.URL{Path: path}, output, writer.WithProfile(streamID, &task.Profile))
	if err != nil {
		return err
	}

	// Decode the requested stream, buffering samples through an audioFIFO so
	// every frame handed to the encoder (bar the last) carries exactly the
	// number of samples it expects - the decoder's own frame size otherwise
	// rarely matches (e.g. mp3's 1152-sample frames feeding an AAC encoder
	// that requires 1024).
	fifo := newAudioFIFO(streamID, w.FrameSize(streamID), w.Encode)
	dec, err := reader.NewDecoder(func(f frame.Frame) error {
		af, ok := f.(*frame.AudioFrame)
		if !ok {
			return gomedia.ErrBadParameter.Withf("stream %d: expected an audio frame, got %T", streamID, f)
		}
		return fifo.write(af)
	})
	if err != nil {
		return errors.Join(err, w.Close())
	}
	if err := dec.Add(streamID); err != nil {
		return errors.Join(err, w.Close())
	}
	if err := rd.Decode(ctx, nil, dec); err != nil {
		return errors.Join(err, w.Close())
	}
	if err := fifo.flush(); err != nil {
		return errors.Join(err, w.Close())
	}
	if err := w.Flush(streamID); err != nil {
		return errors.Join(err, w.Close())
	}
	if err := w.Close(); err != nil {
		return err
	}

	// Set result
	ctx.Result(&AudioEncodeMediaResponse{Path: path})

	// Return success
	return nil
}
