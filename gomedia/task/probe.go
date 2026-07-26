package task

import (
	"io"
	"maps"
	"net/url"
	"slices"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
	attribute "go.opentelemetry.io/otel/attribute"
)

///////////////////////////////////////////////////////////////////////////////
// PROBE TASK

type ProbeRequest struct {
	Reader io.Reader `json:"-"` // supplied from the request body, not a query/JSON field
	Format string    `json:"format,omitempty" name:"format" help:"Input format name (e.g. mpegts)"`
	Opts   []string  `json:"opts,omitempty" name:"opts" help:"Input format options"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

// Query returns the Format/Opts fields as URL query parameters, for a client
// to attach to the probe request (Reader is carried as the request body, not
// a query parameter).
func (r ProbeRequest) Query() url.Values {
	query := url.Values{}
	if r.Format != "" {
		query.Set("format", r.Format)
	}
	for _, opt := range r.Opts {
		query.Add("opts", opt)
	}
	return query
}

type ProbeResponse struct {
	Name     string                   `json:"name,omitempty" help:"Name of the probed input, if known (e.g. the uploaded filename)." example:"sample.mp3"`
	Format   *profile.FormatMeta      `json:"format,omitempty" help:"Detected container format."`
	Duration profile.Duration         `json:"duration,omitempty" help:"Media duration, as a duration string (e.g. \"1h2m3.5s\"); absent if unknown - which is common for streamed, non-seekable input, since duration often can't be estimated without seeking." example:"1h2m3.5s"`
	Streams  []*profile.StreamProfile `json:"streams,omitempty" help:"Audio, video, subtitle, data, and attachment streams found in the input."`
	Metadata []profile.Metadata       `json:"metadata,omitempty" help:"Container-level metadata tags, e.g. \"title\", \"artist\"; excludes artwork and chapters."`
	Artwork  []profile.Artwork        `json:"artwork,omitempty" help:"Embedded artwork (cover art, thumbnail, etc.) found in the input; absent if none was found."`
	Chapters []profile.Chapter        `json:"chapters,omitempty" help:"Chapter markers found in the input, if any."`
}

type ProbeTask struct {
	req ProbeRequest
}

var _ Task = (*ProbeTask)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProbeTask(req ProbeRequest) (*ProbeTask, error) {
	self := &ProbeTask{req: req}
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (task *ProbeTask) Run(ctx Context) (err error) {
	var result ProbeResponse
	if named, ok := task.req.Reader.(gomedia.NamedReader); ok && named != nil {
		result.Name = named.Name()
	}

	_, endSpan := otel.StartSpan(ctx.Tracer, ctx, "Probe",
		attribute.String("input", result.Name),
		attribute.String("input_format", task.req.Format),
	)
	defer func() { endSpan(err) }()

	// Check for a nil reader
	if task.req.Reader == nil {
		return gomedia.ErrBadParameter.With("nil reader")
	}

	// Create a new reader for the input stream
	reader, err := reader.NewReader(task.req.Reader, reader.WithInput(task.req.Format, task.req.Opts...))
	if err != nil {
		return err
	}
	defer reader.Close()

	// Containers and streams information. Only the format's identifying
	// metadata is reported here, not its full input/output/codec-support
	// details.
	if format := reader.Format(); format != nil {
		result.Format = &format.FormatMeta
	}
	result.Duration = profile.Duration(reader.Duration())

	streams := reader.Streams()
	result.Streams = make([]*profile.StreamProfile, 0, len(streams))
	for _, id := range slices.Sorted(maps.Keys(streams)) {
		if sp, ok := streams[id].(*profile.StreamProfile); ok {
			result.Streams = append(result.Streams, sp)
		}
	}

	// Metadata-level information
	result.Metadata = profile.NewMetadataList(reader.Metadata())
	result.Artwork = profile.NewArtworkList(reader.Metadata(gomedia.MetaArtwork))
	result.Chapters = profile.NewChapterList(reader.Metadata(gomedia.MetaChapter))

	if ctx.Result != nil {
		ctx.Result(&result)
	}

	// Return success
	return nil
}
