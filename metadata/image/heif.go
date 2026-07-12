package image

import (
	"context"
	"io"
	"regexp"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	heif "github.com/mutablelogic/go-media/pkg/heif"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	metadata.AddHandler(regexp.MustCompile(`^image/(?:heic|heics|heif|heifs|avif|avis)$`), func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		data, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		h, err := heif.Parse(data)
		if err != nil {
			return nil, err
		}
		defer h.Close()

		entries := make(map[string]gomedia.Metadata)
		for _, m := range h.Metadata() {
			entries[m.Key()] = m
		}

		return metadata.FilterMetadata(entries, filter), nil
	}, "tiff", "exif", "dc", "xmp")

	metadata.AddHandler(regexp.MustCompile(`^image/(?:heic|heics|heif|heifs|avif|avis)$`), func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		if filter != "artwork:" && filter != "artwork:thumbnail" {
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
