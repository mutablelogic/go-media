package video

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	imageutil "github.com/mutablelogic/go-media/metadata/image"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Add metadata handler for embedded cover art in video files
	metadata.AddHandler(regexp.MustCompile(`^video/.*$`), func(_ context.Context, r io.Reader, filter string) ([]gomedia.Metadata, error) {
		// Reject unless an "artwork:" namespace filter was requested
		namespace, _, hasNamespace := strings.Cut(strings.ToLower(filter), ":")
		if !hasNamespace || namespace != "artwork" {
			return nil, nil
		}

		reader, err := ffmpeg.NewReader(r)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		// Attached-picture streams (e.g. an MP4 "covr" atom or ID3 APIC
		// frame), if any. A file can carry more than one; the first is
		// keyed "artwork:cover", subsequent ones "artwork:cover-2", and
		// so on.
		artwork := reader.Metadata(ffmpeg.MetaArtwork)
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

		return metadata.FilterMetadata(entries, filter), nil
	}, "artwork")
}
