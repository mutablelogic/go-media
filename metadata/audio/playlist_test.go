package audio

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	// Packages
	metadata "github.com/mutablelogic/go-media/metadata"
)

// playlistMetadata returns the metadata of a playlist, by key
func playlistMetadata(t *testing.T, contentType, data string) map[string]any {
	t.Helper()
	meta, err := metadata.GetMetadata(context.Background(), strings.NewReader(data), contentType)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]any, len(meta))
	for _, m := range meta {
		result[m.Key()] = m.Any()
	}
	return result
}

func Test_playlist_m3u(t *testing.T) {
	t.Run("Extended", func(t *testing.T) {
		meta := playlistMetadata(t, "audio/x-mpegurl", "#EXTM3U\n#PLAYLIST:Road Trip\n#EXTINF:123,Artist One - First Song\nmusic/01.mp3\n\n#EXTINF:200.5,Artist Two - Second Song\nhttp://example.com/02.mp3\n")
		assertPlaylist(t, meta, map[string]any{
			"playlist:format":   "m3u",
			"playlist:count":    2,
			"dc:title":          "Road Trip",
			"playlist:duration": 323500 * time.Millisecond,
			"playlist:titles":   []string{"Artist One - First Song", "Artist Two - Second Song"},
			"playlist:entries":  []string{"music/01.mp3", "http://example.com/02.mp3"},
		})
	})

	t.Run("Attributes", func(t *testing.T) {
		// An IPTV playlist, where an attribute has a comma, and the
		// durations are unknown, so there's no total
		meta := playlistMetadata(t, "audio/mpegurl", "#EXTM3U\n#EXTINF:-1 tvg-id=\"bbc1\" tvg-name=\"BBC One, HD\" group-title=\"UK\",BBC One\nhttp://example.com/bbc1\n")
		assertPlaylist(t, meta, map[string]any{
			"playlist:format":  "m3u",
			"playlist:count":   1,
			"playlist:titles":  []string{"BBC One"},
			"playlist:entries": []string{"http://example.com/bbc1"},
		})
		assertNo(t, meta, "playlist:duration", "dc:title")
	})

	t.Run("Simple", func(t *testing.T) {
		// Locations only, with a comment
		meta := playlistMetadata(t, "audio/x-mpegurl", "# my mix\r\nC:\\Music\\a.mp3\r\nC:\\Music\\b.mp3\r\n")
		assertPlaylist(t, meta, map[string]any{
			"playlist:count":   2,
			"playlist:entries": []string{`C:\Music\a.mp3`, `C:\Music\b.mp3`},
		})
		assertNo(t, meta, "playlist:titles", "playlist:duration")
	})

	t.Run("Windows1252", func(t *testing.T) {
		meta := playlistMetadata(t, "audio/x-mpegurl", "#EXTM3U\n#EXTINF:60,Bj\xf6rk - J\xf3ga\nj\xf3ga.mp3\n")
		assertPlaylist(t, meta, map[string]any{
			"playlist:titles":  []string{"Björk - Jóga"},
			"playlist:entries": []string{"jóga.mp3"},
		})
	})

	t.Run("Empty", func(t *testing.T) {
		meta := playlistMetadata(t, "audio/x-mpegurl", "")
		assertPlaylist(t, meta, map[string]any{"playlist:format": "m3u", "playlist:count": 0})
		assertNo(t, meta, "playlist:duration", "playlist:entries")
	})

	t.Run("Limit", func(t *testing.T) {
		// Every entry is counted, but only the first are returned
		var b strings.Builder
		for i := range maxPlaylistEntries + 5 {
			fmt.Fprintf(&b, "#EXTINF:1,Track %d\ntrack%d.mp3\n", i, i)
		}
		meta := playlistMetadata(t, "audio/x-mpegurl", b.String())
		assertPlaylist(t, meta, map[string]any{"playlist:count": maxPlaylistEntries + 5})
		if titles, _ := meta["playlist:titles"].([]string); len(titles) != maxPlaylistEntries {
			t.Errorf("titles = %d, want %d", len(titles), maxPlaylistEntries)
		}
	})
}

func Test_playlist_hls(t *testing.T) {
	t.Run("Media", func(t *testing.T) {
		// Segments, which have a duration, but aren't listed
		meta := playlistMetadata(t, "application/vnd.apple.mpegurl", "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:10\n#EXTINF:9.009,\nsegment0.ts\n#EXTINF:9.009,\nsegment1.ts\n#EXT-X-ENDLIST\n")
		assertPlaylist(t, meta, map[string]any{
			"playlist:format":   "hls",
			"playlist:count":    2,
			"playlist:duration": 18018 * time.Millisecond,
		})
		assertNo(t, meta, "playlist:entries", "playlist:titles")
	})

	t.Run("Master", func(t *testing.T) {
		// Streams, which have no duration
		meta := playlistMetadata(t, "application/x-mpegurl", "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1280000,RESOLUTION=640x360\nlow/index.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=2560000\nhigh/index.m3u8\n")
		assertPlaylist(t, meta, map[string]any{"playlist:format": "hls", "playlist:count": 2})
		assertNo(t, meta, "playlist:duration", "playlist:entries")
	})
}

func Test_playlist_pls(t *testing.T) {
	// Entries are in the order of their numbers, the keys are in any case,
	// and an entry without a file isn't one
	meta := playlistMetadata(t, "audio/x-scpls", "[playlist]\nFile2=http://example.com/second\nTitle2=Second\nLength2=-1\nfile1=first.mp3\ntitle1=First\nlength1=30\nTitle3=No file\nNumberOfEntries=2\nVersion=2\n")
	assertPlaylist(t, meta, map[string]any{
		"playlist:format":  "pls",
		"playlist:count":   2,
		"playlist:titles":  []string{"First", "Second"},
		"playlist:entries": []string{"first.mp3", "http://example.com/second"},
	})
	assertNo(t, meta, "playlist:duration")
}

// assertPlaylist checks the metadata has the values
func assertPlaylist(t *testing.T, meta map[string]any, want map[string]any) {
	t.Helper()
	for key, value := range want {
		if got := fmt.Sprintf("%#v", meta[key]); got != fmt.Sprintf("%#v", value) {
			t.Errorf("%s = %s, want %#v", key, got, value)
		}
	}
}

// assertNo checks the metadata doesn't have the keys
func assertNo(t *testing.T, meta map[string]any, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if value, exists := meta[key]; exists {
			t.Errorf("%s = %#v, want none", key, value)
		}
	}
}
