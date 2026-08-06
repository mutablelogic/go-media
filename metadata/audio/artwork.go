package audio

import (
	"context"
	"fmt"
	"io"
	"regexp"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	imageutil "github.com/mutablelogic/go-media/metadata/image"
	reader "github.com/mutablelogic/go-media/reader"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Add metadata handler for embedded cover art in audio files
	metadata.AddHandler(regexp.MustCompile(`^audio/.*$`), "artwork", func(_ context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		// Reject unless the "artwork" namespace was requested
		if !o.HasNamespace("artwork") {
			return nil, nil
		}

		rd, err := reader.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer rd.Close()

		// Attached-picture streams (e.g. ID3 APIC frames), if any. A file
		// can carry more than one (front cover, back cover, artist photo,
		// etc.); the first is keyed "artwork:cover", subsequent ones
		// "artwork:cover-2", "artwork:cover-3", and so on.
		artwork := rd.Metadata(gomedia.MetaArtwork)
		if len(artwork) == 0 {
			return nil, nil
		}

		entries := make(map[string]gomedia.Metadata, len(artwork))
		for i, pic := range artwork {
			key := "artwork:cover"
			if i > 0 {
				key = fmt.Sprintf("artwork:cover-%d", i+1)
			}
			m, err := imageutil.ExtractArtwork(pic.Bytes(), key)
			if err != nil {
				continue
			}
			entries[key] = m
		}

		return metadata.FilterMetadata(entries, o), nil
	}, "artwork")
}
