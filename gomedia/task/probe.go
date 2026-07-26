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
	Name     string                   `json:"name,omitempty"`
	Format   *profile.FormatMeta      `json:"format,omitempty"`
	Duration profile.Duration         `json:"duration,omitempty"`
	Streams  []*profile.StreamProfile `json:"streams,omitempty"`
	Metadata []gomedia.Metadata       `json:"metadata,omitempty"`
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

	// Metadata-level information.
	//metadata := reader.Metadata()

	// Artwork information.

	if ctx.Result != nil {
		ctx.Result(&result)
	}

	// Return success
	return nil
}
