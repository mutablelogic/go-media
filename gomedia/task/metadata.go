package task

import (
	"bytes"
	"io"
	"net/url"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	profile "github.com/mutablelogic/go-media/profile/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"

	// Register metadata handlers
	_ "github.com/mutablelogic/go-media/metadata/application"
	_ "github.com/mutablelogic/go-media/metadata/audio"
	_ "github.com/mutablelogic/go-media/metadata/image"
	_ "github.com/mutablelogic/go-media/metadata/video"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type seeker struct {
	io.ReadSeeker
	name string
}

type MetadataRequest struct {
	Reader io.Reader `json:"-"`
}

type MetadataResponse struct {
	Name     string             `json:"name,omitempty" help:"Name of the probed input, if known (e.g. the uploaded filename)." example:"sample.mp3"`
	Type     string             `json:"type,omitempty" help:"Detected content type (MIME type) of the input, e.g. \"video/mp4\"." example:"video/mp4"`
	Metadata []profile.Metadata `json:"metadata,omitempty" help:"Container-level metadata tags, e.g. \"title\", \"artist\"; excludes artwork and chapters."`
	Artwork  []profile.Artwork  `json:"artwork,omitempty" help:"Artwork images embedded in the input, e.g. album cover art."`
}

var _ Task = (*MetadataRequest)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r MetadataRequest) Query() url.Values {
	return url.Values{}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (task *MetadataRequest) Run(ctx Context) (err error) {
	var result MetadataResponse
	if named, ok := task.Reader.(gomedia.NamedReader); ok && named != nil {
		result.Name = named.Name()
	}

	_, endSpan := otel.StartSpan(ctx.Tracer, ctx, "Metadata",
		attribute.String("req", types.Stringify(task)),
		attribute.String("name", result.Name),
	)
	defer func() { endSpan(err) }()

	// Create a seeker
	seeker, err := seekerFromReader(task.Reader)
	if err != nil {
		return err
	}

	// First pass: content type and its parameters.
	contentType, params, err := metadata.ContentType(seeker)
	if err != nil {
		return err
	} else {
		result.Type = contentType
		for key, value := range params {
			result.Metadata = append(result.Metadata, profile.Metadata{Key: key, Value: value})
		}
	}

	// Seek to start
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Second pass: extract metadata
	items, err := metadata.GetMetadata(ctx, seeker, contentType)
	if err != nil && len(items) == 0 {
		return err
	} else if err != nil {
		// TODO: Log the warning or handle it as needed
	}

	// Append additional metadata items into the result
	for _, item := range items {
		result.Metadata = append(result.Metadata, profile.Metadata{Key: item.Key(), Value: item.Value(), Any: item.Any()})
	}

	// Seek to start
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Third pass: extract artwork
	items, err = metadata.GetMetadata(ctx, seeker, contentType, metadata.WithNamespace("artwork"))
	if err != nil && len(items) == 0 {
		return err
	} else if err != nil {
		// TODO: Log the warning or handle it as needed
	}

	// Append artwork items into the result
	result.Artwork = append(result.Artwork, profile.NewArtworkList(items)...)

	// Set result
	ctx.Result(types.Ptr(result))

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - SEEKER

func seekerFromReader(r io.Reader) (io.ReadSeeker, error) {
	var self seeker

	// Check arguments and set name
	if r == nil {
		return nil, gomedia.ErrBadParameter.With("nil reader")
	} else if named, ok := r.(gomedia.NamedReader); ok && named != nil {
		self.name = named.Name()
	}

	// Make the input replayable so we can read once for content type
	// detection and a second time for metadata extraction.
	if seeker, ok := r.(io.ReadSeeker); ok {
		self.ReadSeeker = seeker
	} else {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			return nil, err
		} else {
			self.ReadSeeker = bytes.NewReader(buf.Bytes())
		}
	}

	// Seek to start
	if _, err := self.ReadSeeker.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	// Return success
	return self, nil
}

func (r seeker) Name() string {
	return r.name
}
