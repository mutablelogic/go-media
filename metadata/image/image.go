package image

import (
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"regexp"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	types "github.com/mutablelogic/go-server/pkg/types"

	// Image decoders
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/mutablelogic/go-media/pkg/heif"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type imageMetadata struct {
	key   string
	value any
}

func (m *imageMetadata) Key() string {
	return m.key
}

func (m *imageMetadata) Value() string {
	switch v := m.value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprint(v)
	case []string:
		return strings.Join(v, ", ")
	default:
		return ""
	}
}

func (m *imageMetadata) Bytes() []byte {
	return nil
}

func (m *imageMetadata) Image() image.Image {
	switch v := m.value.(type) {
	case image.Image:
		return v
	default:
		return nil
	}
}

func (m *imageMetadata) Any() any {
	return m.value
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Add metadata handler for image files in general
	metadata.AddHandler(regexp.MustCompile("^image/.*$"), "image", func(_ context.Context, r io.Reader, _ *metadata.Opts) ([]gomedia.Metadata, error) {
		// Decode the image's format and size, from its header, rather than
		// decoding the image, which can be large
		config, format, err := image.DecodeConfig(r)
		if errors.Is(err, image.ErrFormat) {
			// There's no decoder for the format, such as an SVG image,
			// which isn't an error, as other handlers may read it
			return nil, nil
		} else if err != nil {
			return nil, err
		}

		// Return metadata
		return []gomedia.Metadata{
			types.Ptr(imageMetadata{"image:format", format}),
			types.Ptr(imageMetadata{"image:width", config.Width}),
			types.Ptr(imageMetadata{"image:height", config.Height}),
		}, nil
	}, "image")
}
