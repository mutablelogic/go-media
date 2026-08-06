package metadata

import (
	"bytes"
	"errors"
	"io"
	"sync"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	profile "github.com/mutablelogic/go-media/profile/schema"
	schema "github.com/mutablelogic/go-media/task/schema"
	"github.com/mutablelogic/go-server/pkg/types"

	// Register metadata handlers
	_ "github.com/mutablelogic/go-media/metadata/application"
	_ "github.com/mutablelogic/go-media/metadata/audio"
	_ "github.com/mutablelogic/go-media/metadata/image"
	_ "github.com/mutablelogic/go-media/metadata/video"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// namedReader pairs an in-memory reader over the input's buffered bytes
// with its name, so gomedia.NamedReader-aware handlers (e.g. the TMDB
// lookup, which needs the filename) still work once the input has been
// read into memory up front.
type namedReader struct {
	*bytes.Reader
	name string
}

func (r namedReader) Name() string { return r.name }

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (req *MetadataRequest) Run(ctx schema.Context) error {
	if err := req.Validate(); err != nil {
		return err
	}

	// Create a new ReadSeeker from the Reader.
	// This will read the entire reader into memory if it is not already a ReadSeeker.
	seeker, err := schema.NewReadSeeker(req.Reader)
	if err != nil {
		return err
	}

	// First pass: content type and its parameters.
	var result MetadataResponse
	result.Name = seeker.Name()
	contentType, params, err := metadata.ContentType(seeker)
	if err != nil {
		return err
	}

	// Set the result type and some additional metadata parameters
	result.Type = contentType
	for key, value := range params {
		result.Metadata = append(result.Metadata, profile.NewMetadata(key, value))
	}

	// Always set result, regardless of the error
	defer func() {
		ctx.Result(types.Ptr(result))
	}()

	// Buffer the whole input once, up front, so the metadata and artwork
	// passes below can each get their own independent reader over it and
	// run concurrently - an io.ReadSeeker can't safely be Seek'd and Read
	// from two goroutines at once, and re-reading the (possibly large,
	// possibly network-mounted) input from scratch for each pass, as two
	// sequential GetMetadata calls would otherwise do internally, is pure
	// waste when it's the same bytes both times.
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return err
	}
	data, err := io.ReadAll(seeker)
	if err != nil {
		return err
	}
	name := seeker.Name()
	newReader := func() io.Reader {
		return namedReader{bytes.NewReader(data), name}
	}

	// Artwork extraction is opt-in per handler and unrelated to general
	// metadata extraction, so run both passes concurrently rather than one
	// after the other - the general pass alone can take hundreds of ms
	// when a handler makes a network call (e.g. the TMDB lookup).
	artworkOpts := append(append([]metadata.Option{}, ctx.MetaOpts()...), metadata.WithNamespace("artwork"))

	var (
		wg                  sync.WaitGroup
		metaItems, artItems []gomedia.Metadata
		metaErr, artErr     error
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		metaItems, metaErr = metadata.GetMetadata(ctx, newReader(), contentType, ctx.MetaOpts()...)
	}()
	go func() {
		defer wg.Done()
		artItems, artErr = metadata.GetMetadata(ctx, newReader(), contentType, artworkOpts...)
	}()
	wg.Wait()

	// A pass that returned an error alongside zero items has hard-failed;
	// otherwise the error is a warning alongside the items it did manage
	// to return (see GetMetadata's own doc comment). Both passes still get
	// to contribute whatever they found, even if the other hard-failed.
	var runErr error
	if metaErr != nil && len(metaItems) == 0 {
		runErr = errors.Join(runErr, metaErr)
	} else if metaErr != nil {
		// TODO: Log the warning or handle it as needed
	}
	for _, item := range metaItems {
		result.Metadata = append(result.Metadata, profile.Metadata{Key: item.Key(), Value: item.Value(), Any: item.Any()})
	}

	if artErr != nil && len(artItems) == 0 {
		runErr = errors.Join(runErr, artErr)
	} else if artErr != nil {
		// TODO: Log the warning or handle it as needed
	}
	result.Artwork = append(result.Artwork, profile.NewArtworkList(artItems)...)

	return runErr
}
