package task

import (
	"io"
	"maps"
	"net/url"
	"slices"
	"strings"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	profile "github.com/mutablelogic/go-media/profile/schema"
	reader "github.com/mutablelogic/go-media/reader"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

///////////////////////////////////////////////////////////////////////////////
// PROBE TASK

type ProbeMediaRequest struct {
	Reader io.Reader `json:"-"`
	ProbeRequestOpts
}

// Url is a plain string, not *url.URL - go-server's httprequest.Query has
// no case for a *url.URL field (only strings, numbers, bools, slices, and
// time.Time), so a string keeps ProbeSourceRequest decodable as a whole via
// the generic decoder rather than needing a field-by-field workaround. It's
// parsed to *url.URL in ProbeSourceTask.Run.
type ProbeSourceRequest struct {
	Url string `json:"url" name:"url" help:"URL of the media to probe, e.g. \"https://example.com/sample.mp3\"." example:"https://example.com/sample.mp3"`
	ProbeRequestOpts
}

type ProbeRequestOpts struct {
	Format string   `json:"format,omitempty" name:"format" help:"Input format name (e.g. mpegts)"`
	Opts   []string `json:"opts,omitempty" name:"opts" help:"Input format options"`
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

type ProbeMediaTask struct {
	req ProbeMediaRequest
}

type ProbeSourceTask struct {
	req ProbeSourceRequest
}

var _ Task = (*ProbeMediaTask)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewProbeMediaTask(req ProbeMediaRequest) (*ProbeMediaTask, error) {
	return &ProbeMediaTask{req: req}, nil
}

func NewProbeSourceTask(req ProbeSourceRequest) (*ProbeSourceTask, error) {
	return &ProbeSourceTask{req: req}, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

// Query returns the Format/Opts fields as URL query parameters, for a client
// to attach to the probe request (Reader is carried as the request body, not
// a query parameter).
func (r ProbeRequestOpts) Query() url.Values {
	query := url.Values{}
	if r.Format != "" {
		query.Set("format", r.Format)
	}
	for _, opt := range r.Opts {
		query.Add("opts", opt)
	}
	return query
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (task *ProbeMediaTask) Run(ctx Context) (err error) {
	var result ProbeResponse
	if named, ok := task.req.Reader.(gomedia.NamedReader); ok && named != nil {
		result.Name = named.Name()
	}

	_, endSpan := otel.StartSpan(ctx.Tracer, ctx, "ProbeMedia",
		attribute.String("req", types.Stringify(task.req)),
		attribute.String("name", result.Name),
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

	// Set result
	ctx.Result(types.Ptr(result))

	// Return success
	return nil
}

func (task *ProbeSourceTask) Run(ctx Context) (err error) {
	if task.req.Url == "" {
		return gomedia.ErrBadParameter.With("missing URL")
	}
	u, err := url.Parse(task.req.Url)
	if err != nil {
		return gomedia.ErrBadParameter.Withf("invalid URL %q: %w", task.req.Url, err)
	}

	var result ProbeResponse
	result.Name = u.String()

	_, endSpan := otel.StartSpan(ctx.Tracer, ctx, "ProbeSource",
		attribute.String("url", result.Name),
	)
	defer func() { endSpan(err) }()

	// "device://<format>/<address>" (e.g. "device://avfoundation/0:0",
	// "device://v4l2//dev/video0") names an input device rather than a URL
	// FFmpeg can open directly: the demuxer name comes from the host and
	// the device's own address (whatever ListDevices/WithInput expects for
	// that format) from the path. Anything else must be a scheme FFmpeg
	// actually has a protocol registered for (http, https, rtmp, ...) -
	// checked up front via reader.Protocols so an unsupported scheme fails
	// fast with a clear error rather than a confusing one from
	// AVFormat_open_url. Note this only rules out schemes FFmpeg has no
	// protocol for at all; it doesn't guarantee the URL is otherwise valid
	// or reachable ("rtsp" itself, for instance, is a container format, not
	// a protocol, so it never appears in this list even though it's a
	// perfectly valid scheme to open).
	inputFormat, address := task.req.Format, u.String()
	if u.Scheme == "device" {
		inputFormat = u.Host
		address = strings.TrimPrefix(u.Path, "/")
		if inputFormat == "" || address == "" {
			return gomedia.ErrBadParameter.Withf("invalid device URL %q, expected \"device://<format>/<address>\"", u.String())
		}
	} else if !slices.Contains(reader.Protocols(), u.Scheme) {
		return gomedia.ErrBadParameter.Withf("unsupported URL scheme %q", u.Scheme)
	}

	rdr, err := reader.Open(address, reader.WithInput(inputFormat, task.req.Opts...))
	if err != nil {
		return err
	}
	defer rdr.Close()

	if format := rdr.Format(); format != nil {
		result.Format = &format.FormatMeta
	}
	result.Duration = profile.Duration(rdr.Duration())

	streams := rdr.Streams()
	result.Streams = make([]*profile.StreamProfile, 0, len(streams))
	for _, id := range slices.Sorted(maps.Keys(streams)) {
		if sp, ok := streams[id].(*profile.StreamProfile); ok {
			result.Streams = append(result.Streams, sp)
		}
	}

	result.Metadata = profile.NewMetadataList(rdr.Metadata())
	result.Artwork = profile.NewArtworkList(rdr.Metadata(gomedia.MetaArtwork))
	result.Chapters = profile.NewChapterList(rdr.Metadata(gomedia.MetaChapter))

	// Set result
	ctx.Result(types.Ptr(result))

	// Return success
	return nil
}
