package image

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"regexp"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	types "github.com/mutablelogic/go-server/pkg/types"
	xdraw "golang.org/x/image/draw"

	// Image decoders
	_ "image/gif"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// artworkMetadata wraps an extracted artwork image as gomedia.Metadata.
// Value() returns the encoded MIME type, Bytes() returns the encoded image
// data, and Image()/Any() return the decoded (and possibly resized) image.
type artworkMetadata struct {
	key      string
	mimeType string
	data     []byte
	img      image.Image
}

func (m *artworkMetadata) Key() string        { return m.key }
func (m *artworkMetadata) Value() string      { return m.mimeType }
func (m *artworkMetadata) Bytes() []byte      { return m.data }
func (m *artworkMetadata) Image() image.Image { return m.img }
func (m *artworkMetadata) Any() any           { return m.img }

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// The maximum width of the artwork to extract. If the image is larger than this, it will be resized,
	// but the aspect ratio will be preserved. If the image is smaller than this, it will be returned as is,
	// but potentially converted to a jpeg.
	MaxWidth = 640

	// jpegQuality is the quality used when re-encoding artwork as JPEG.
	jpegQuality = 90

	// Formats that should be encoded as PNG when resized or re-encoded. All other formats are encoded as JPEG.
	PNGFormats = "png,gif,bmp,tiff"
)

// pngFormats is PNGFormats split into a set, for exact (not substring) matching.
var pngFormats = func() map[string]bool {
	set := make(map[string]bool)
	for _, format := range strings.Split(PNGFormats, ",") {
		set[format] = true
	}
	return set
}()

////////////////////////////////////////////////////////////////////////////////
// SHARED HELPERS

// ExtractArtwork decodes image data and returns it as a gomedia.Metadata
// entry under key, resized to at most MaxWidth (preserving aspect ratio)
// and re-encoded as PNG or JPEG as appropriate. Already-jpeg/png images
// near the target width are returned unchanged rather than re-encoded. It
// is shared by the generic image/* artwork handler, for RAW files' embedded
// thumbnail, and by metadata/audio for embedded cover art.
func ExtractArtwork(data []byte, key string) (gomedia.Metadata, error) {
	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Fast-path: if the image is already a jpeg or png and approximately the
	// max width, return the original bytes unchanged rather than re-encoding
	if (format == "png" || format == "jpeg") && img.Bounds().Dx() <= int(float64(MaxWidth)*1.1) {
		return types.Ptr(artworkMetadata{key: key, mimeType: "image/" + format, data: data, img: img}), nil
	}

	// Preserve the aspect ratio and resize the image down to the max
	// width, if it's larger than that
	if img.Bounds().Dx() > MaxWidth {
		width := MaxWidth
		height := img.Bounds().Dy() * width / img.Bounds().Dx()
		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		xdraw.BiLinear.Scale(dst, dst.Rect, img, img.Bounds(), xdraw.Over, nil)
		img = dst
	}

	// Encode the (possibly resized) image: lossless PNG for formats that
	// are themselves lossless (or palette-based), lossy JPEG otherwise
	var buf bytes.Buffer
	var mimeType string
	if pngFormats[format] {
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
		mimeType = "image/png"
	} else {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return nil, err
		}
		mimeType = "image/jpeg"
	}

	return types.Ptr(artworkMetadata{key: key, mimeType: mimeType, data: buf.Bytes(), img: img}), nil
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Add metadata handler for image files in general
	metadata.AddHandler(regexp.MustCompile("^image/.*$"), func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		// Reject when filter is not "artwork:" or "artwork:thumbnail"
		if filter != "artwork:" && filter != "artwork:thumbnail" {
			return nil, nil
		}

		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		m, err := ExtractArtwork(data, "artwork:thumbnail")
		if err != nil {
			return nil, err
		}
		return []gomedia.Metadata{m}, nil
	}, "artwork")
}
