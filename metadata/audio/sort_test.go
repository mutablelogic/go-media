package audio

import (
	"testing"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
)

func Test_fillFromSortTags(t *testing.T) {
	// entries returns the metadata with the values, by key
	entries := func(values map[string]string) map[string]gomedia.Metadata {
		result := make(map[string]gomedia.Metadata, len(values))
		for key, value := range values {
			result[key] = meta{key: key, value: value}
		}
		return result
	}

	t.Run("OnlySortTags", func(t *testing.T) {
		// As Apple Music downloads have
		got := entries(map[string]string{
			"audio:sort-name":   "Jungle Love",
			"audio:sort-artist": "Time",
			"audio:sort-album":  "Ice Cream Castle",
		})
		fillFromSortTags(got)
		for key, want := range map[string]string{"dc:title": "Jungle Love", "dc:creator": "Time", "audio:Album": "Ice Cream Castle", "audio:sort-name": "Jungle Love"} {
			if got[key] == nil || got[key].Value() != want {
				t.Errorf("%s = %v, want %q", key, got[key], want)
			}
		}
	})

	t.Run("TagsWin", func(t *testing.T) {
		// A file's own title, artist and album are kept
		got := entries(map[string]string{
			"dc:title":          "The Jungle Love",
			"dc:creator":        "The Time",
			"audio:Album":       "Ice Cream Castle (Remastered)",
			"audio:sort-name":   "Jungle Love",
			"audio:sort-artist": "Time",
			"audio:sort-album":  "Ice Cream Castle",
		})
		fillFromSortTags(got)
		for key, want := range map[string]string{"dc:title": "The Jungle Love", "dc:creator": "The Time", "audio:Album": "Ice Cream Castle (Remastered)"} {
			if got[key].Value() != want {
				t.Errorf("%s = %q, want %q", key, got[key].Value(), want)
			}
		}
	})

	t.Run("AlbumArtist", func(t *testing.T) {
		// The sort artist is preferred to the sort album artist, which is
		// used without it, and an empty sort tag isn't
		got := entries(map[string]string{"audio:sort-album-artist": "Various Artists", "audio:sort-name": " "})
		fillFromSortTags(got)
		if got["dc:creator"] == nil || got["dc:creator"].Value() != "Various Artists" {
			t.Errorf("dc:creator = %v", got["dc:creator"])
		}
		if _, exists := got["dc:title"]; exists {
			t.Errorf("unexpected dc:title %v", got["dc:title"])
		}
	})
}
