package manager

import (
	"context"
	"errors"
	"strings"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	goschema "github.com/mutablelogic/go-media/gomedia/schema"
	metadata "github.com/mutablelogic/go-media/metadata"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	ffschema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Probe a media stream from any reader and return information about its
// container format and streams.
func (m *Media) Probe(ctx context.Context, req goschema.ProbeRequest) (_ *goschema.ProbeResponse, err error) {
	name := "reader"
	if named, ok := req.Reader.(metadata.NamedStream); ok {
		name = named.Name()
	}

	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "Probe",
		attribute.String("input", name),
		attribute.String("input_format", req.InputFormat),
	)
	defer func() { endSpan(err) }()

	if req.Reader == nil {
		return nil, errors.New("nil reader")
	}

	reader, err := ffmpeg.NewReader(req.Reader, ffmpeg.WithInput(req.InputFormat, req.InputOpts...))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Format-level information.
	var formatName, formatDesc string
	var mimeTypes []string
	if inputFormat := reader.InputFormat(); inputFormat != nil {
		formatName = inputFormat.Name()
		formatDesc = inputFormat.LongName()
		if mt := inputFormat.MimeTypes(); mt != "" {
			mimeTypes = strings.Split(mt, ",")
		}
	}

	// Stream information.
	avStreams := reader.AVStreams()
	streams := make([]*goschema.Stream, 0, len(avStreams))
	for _, avStream := range avStreams {
		if s := ffschema.NewStream(avStream); s != nil {
			streams = append(streams, goschema.WrapStream(s))
		}
	}

	// Response
	resp := &goschema.ProbeResponse{
		Format:      formatName,
		Description: formatDesc,
		MimeTypes:   mimeTypes,
		Duration:    reader.Duration().Seconds(),
		Streams:     streams,
	}

	return resp, nil
}
