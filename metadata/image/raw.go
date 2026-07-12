package image

import (
	"bytes"
	"context"
	"io"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	exif "github.com/mutablelogic/go-media/pkg/exif"
	raw "github.com/mutablelogic/go-media/pkg/raw"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Add metadata handler for RAW camera files
	metadata.AddHandler(raw.ContentTypes, func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		data, err := raw.Read(r)
		if err != nil {
			return nil, err
		}
		defer data.Close()

		// Start with libraw's own curated metadata (Make, Model, ISO, etc.),
		// which is reliable even if the embedded thumbnail can't be read
		entries := make(map[string]gomedia.Metadata)
		for _, m := range data.Metadata() {
			entries[m.Key()] = m
		}

		// Add image dimensions (processed size, after crop/rotation) and a
		// format marker derived from the camera make, since libraw doesn't
		// expose a per-vendor RAW format name (e.g. "CR2"/"ORF") and this
		// handler isn't told which content type was matched
		entries["image:width"] = types.Ptr(imageMetadata{"image:width", data.Width()})
		entries["image:height"] = types.Ptr(imageMetadata{"image:height", data.Height()})
		if cameraMake := strings.ToLower(strings.TrimSpace(data.Make())); cameraMake != "" {
			entries["image:format"] = types.Ptr(imageMetadata{"image:format", cameraMake})
		}

		// Enrich with the full EXIF from the embedded preview JPEG, if one
		// is present and readable; camera firmware typically writes the
		// same (or a superset of) EXIF into the thumbnail as the RAW file
		// itself, so this picks up GPS, MakerNote and other tags that
		// libraw's own curated fields don't cover. On overlapping keys,
		// the thumbnail's EXIF wins, since it uses the same parsing as a
		// standalone JPEG file.
		if thumb, err := data.ThumbnailBytes(); err == nil {
			if e, err := exif.Read(bytes.NewReader(thumb)); err == nil {
				defer e.Close()
				for key, m := range exifTagsToMetadata(e.Tags()) {
					entries[key] = m
				}
			}

			// If artwork was requested, extract/resize/encode the embedded
			// thumbnail the same way a standalone image file would be
			if filter == "artwork:" || filter == "artwork:thumbnail" {
				if m, err := ExtractArtwork(thumb, "artwork:thumbnail"); err == nil {
					entries["artwork:thumbnail"] = m
				}
			}
		}

		return metadata.FilterMetadata(entries, filter), nil
	}, "tiff", "exif", "image", "dc", "artwork")
}
