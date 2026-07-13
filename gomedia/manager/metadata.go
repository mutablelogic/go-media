package manager

import (
	"bytes"
	"context"
	"io"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	metadata "github.com/mutablelogic/go-media/metadata"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type namedReadSeeker struct {
	io.ReadSeeker
	name string
}

func (r namedReadSeeker) Name() string {
	return r.name
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return metadata for a given reader, and any warnings that may have been generated during the
// extraction process.
func (m *Media) GetMetadata(ctx context.Context, r io.Reader, filter string, warn *error) (_ schema.Meta, err error) {
	var name string
	if fr, ok := r.(metadata.NamedStream); ok {
		name = fr.Name()
	}
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "GetMetadata",
		attribute.String("name", name),
		attribute.String("filter", filter),
	)
	defer func() { endSpan(err) }()

	// Make the input replayable so we can read once for content type
	// detection and a second time for metadata extraction.
	var rr io.ReadSeeker
	if seeker, ok := r.(io.ReadSeeker); ok {
		rr = seeker
	} else {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			return schema.Meta{}, err
		}
		if name != "" {
			rr = namedReadSeeker{ReadSeeker: bytes.NewReader(buf.Bytes()), name: name}
		} else {
			rr = bytes.NewReader(buf.Bytes())
		}
	}
	if _, err := rr.Seek(0, io.SeekStart); err != nil {
		return schema.Meta{}, err
	}

	// First pass: content type and its parameters.
	contentType, _, err := metadata.ContentType(rr)
	if err != nil {
		return schema.Meta{}, err
	}

	// Create the result object with the content type and its parameters.
	var result schema.Meta
	result.Name = name
	result.ContentType = contentType
	// TODO
	//	for key := range params {
	//		result = result.Append(key, params[key])
	//	}

	// Second pass: metadata handlers for the detected content type.
	if _, err := rr.Seek(0, io.SeekStart); err != nil {
		return schema.Meta{}, err
	}
	items, err := metadata.GetMetadata(ctx, rr, contentType, filter)
	if err != nil && len(items) == 0 {
		return schema.Meta{}, err
	} else if err != nil && warn != nil {
		*warn = err
	}

	// Append additional metadata items into the result
	for _, item := range items {
		result.Meta = append(result.Meta, schema.MetaItem{Metadata: item})
	}

	// Return success
	return result, nil
}
