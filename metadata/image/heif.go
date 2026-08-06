package image

import (
	"context"
	"io"
	"regexp"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	exif "github.com/mutablelogic/go-media/pkg/exif"
	heif "github.com/mutablelogic/go-media/pkg/heif"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	metadata.AddHandler(regexp.MustCompile(`^image/(?:heic|heics|heif|heifs|avif|avis)$`), "heif", func(_ context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		h, err := heif.Parse(data)
		if err != nil {
			return nil, err
		}
		defer h.Close()

		// Collect raw metadata items, separating out the EXIF tags so they
		// can be enriched the same way as JPEG/RAW (parsed dates, decimal
		// GPS coordinates, float rationals) via the shared helper, rather
		// than leaking libexif's raw Rational types through Any().
		entries := make(map[string]gomedia.Metadata)
		var tags []*exif.Tag
		for _, m := range h.Metadata() {
			if tag, ok := m.(*exif.Tag); ok {
				tags = append(tags, tag)
				continue
			}
			entries[m.Key()] = m
		}
		for key, m := range exifTagsToMetadata(tags) {
			entries[key] = m
		}
		mirrorDCDate(entries)

		return metadata.FilterMetadata(entries, o), nil
	}, "tiff", "exif", "dc", "xmp")

	metadata.AddHandler(regexp.MustCompile(`^image/(?:heic|heics|heif|heifs|avif|avis)$`), "artwork", func(_ context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		// Reject unless the "artwork" namespace was requested
		if !o.HasNamespace("artwork") {
			return nil, nil
		}

		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		h, err := heif.Parse(data)
		if err != nil {
			return nil, err
		}
		defer h.Close()

		thumbs := h.Thumbnails()
		if len(thumbs) == 0 {
			return nil, nil
		}

		entries := make([]gomedia.Metadata, 0, len(thumbs))
		for _, thumb := range thumbs {
			m, err := thumbnailArtwork(thumb)
			if err != nil {
				return nil, err
			}
			if m != nil {
				entries = append(entries, m)
			}
		}

		return entries, nil
	}, "artwork")
}

func isHEIFContainer(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	if string(data[4:8]) != "ftyp" {
		return false
	}
	switch string(data[8:12]) {
	case "heic", "heix", "hevc", "hevm", "hevs", "heim", "heis", "mif1", "mif2", "mif3", "msf1", "avif", "avis", "vvic", "vvis":
		return true
	default:
		return false
	}
}
