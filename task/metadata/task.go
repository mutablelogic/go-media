package metadata

import (
	"io"

	// Packages
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
// PUBLIC METHODS

func (req *MetadataRequest) Run(ctx schema.Context) error {
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

	// Seek to start
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return err
	}

	// Second pass: extract metadata
	items, err := metadata.GetMetadata(ctx, seeker, contentType, "")
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
	items, err = metadata.GetMetadata(ctx, seeker, contentType, "artwork:")
	if err != nil && len(items) == 0 {
		return err
	} else if err != nil {
		// TODO: Log the warning or handle it as needed
	}

	// Append artwork items into the result
	result.Artwork = append(result.Artwork, profile.NewArtworkList(items)...)

	// Return success
	return nil
}
