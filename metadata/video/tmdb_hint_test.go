package video

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
)

// tmdbHint returns the TMDB metadata of a video named name, with the
// options, from the stub
func tmdbHint(t *testing.T, stub http.RoundTripper, name string, opts ...metadata.Option) (map[string]any, error) {
	t.Helper()
	opts = append([]metadata.Option{
		metadata.WithTMDB("token", client.OptTransport(func(http.RoundTripper) http.RoundTripper { return stub })),
		metadata.WithNamespace("tmdb"),
	}, opts...)
	meta, err := metadata.GetMetadata(context.Background(), namedReader{Reader: strings.NewReader(""), name: name}, "video/mp4", opts...)
	result := make(map[string]any, len(meta))
	for _, m := range meta {
		result[m.Key()] = m.Any()
	}
	return result, err
}

func Test_tmdb_hint_movie(t *testing.T) {
	t.Run("ByID", func(t *testing.T) {
		// The movie is got by its identifier, without a search, whatever the
		// name, duration and year
		stub := &tmdbStub{runtimes: map[uint64]uint64{42: 100}}
		meta, err := tmdbHint(t, stub, "Something Else (1970).mp4", metadata.WithTMDBMovie(42), metadata.WithDuration(30*time.Minute), metadata.WithYear(1970))
		if err != nil {
			t.Fatal(err)
		}
		for key, want := range map[string]any{
			"tmdb:id":      "42",
			"tmdb:title":   "Movie 42",
			"tmdb:runtime": 100 * time.Minute,
			"tmdb:imdbid":  "tt0000042",
		} {
			if meta[key] != want {
				t.Errorf("%s = %#v, want %#v", key, meta[key], want)
			}
		}
		if len(stub.queries) != 0 || len(stub.details) != 1 {
			t.Errorf("searches %v, details %v", stub.queries, stub.details)
		}
	})

	t.Run("WithoutName", func(t *testing.T) {
		// A reader without a name has the movie, by its identifier
		stub := &tmdbStub{runtimes: map[uint64]uint64{42: 100}}
		meta, err := metadata.GetMetadata(context.Background(), strings.NewReader(""), "video/mp4",
			metadata.WithTMDB("token", client.OptTransport(func(http.RoundTripper) http.RoundTripper { return stub })),
			metadata.WithNamespace("tmdb"), metadata.WithTMDBMovie(42))
		if err != nil || len(meta) == 0 {
			t.Errorf("got %v, %v", meta, err)
		}
	})

	t.Run("Error", func(t *testing.T) {
		// An error getting the movie is returned
		stub := &tmdbStub{fail: true}
		if _, err := tmdbHint(t, stub, "Alien.mp4", metadata.WithTMDBMovie(42)); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("Replaces", func(t *testing.T) {
		// A movie replaces an episode
		stub := &tmdbStub{runtimes: map[uint64]uint64{42: 100}}
		meta, err := tmdbHint(t, stub, "Alien.mp4", metadata.WithTMDBEpisode(1, 2, 1), metadata.WithTMDBMovie(42))
		if err != nil || meta["tmdb:id"] != "42" || meta["tmdb:showid"] != nil {
			t.Errorf("got %v, %v", meta, err)
		}
	})
}

func Test_tmdb_hint_episode(t *testing.T) {
	t.Run("ByID", func(t *testing.T) {
		// The episode is got by its show's identifier, without a search, and
		// isn't matched by its runtime
		stub := theOffice()
		meta, err := tmdbHint(t, stub, "office-2x1.mkv", metadata.WithTMDBEpisode(2, 2, 1), metadata.WithDuration(90*time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		for key, want := range map[string]any{
			"tmdb:showid":  "2",
			"tmdb:show":    "The Office",
			"tmdb:season":  uint64(2),
			"tmdb:episode": uint64(1),
			"tmdb:runtime": 30 * time.Minute,
		} {
			if meta[key] != want {
				t.Errorf("%s = %#v, want %#v", key, meta[key], want)
			}
		}
		for _, path := range stub.requests {
			if strings.HasPrefix(path, "/3/search/") {
				t.Errorf("searched: %v", stub.requests)
			}
		}
	})

	t.Run("TypicalRuntime", func(t *testing.T) {
		// An episode whose runtime isn't known has the show's typical one
		stub := &tvStub{shows: []tvShow{{id: 7, name: "Show", runtimes: map[string]uint64{"1/3": 0}, typical: 25}}}
		meta, err := tmdbHint(t, stub, "show.mkv", metadata.WithTMDBEpisode(7, 1, 3))
		if err != nil || meta["tmdb:runtime"] != 25*time.Minute {
			t.Errorf("got %v, %v", meta, err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		// A show or episode which TMDB doesn't have has no metadata
		for _, opt := range []metadata.Option{metadata.WithTMDBEpisode(9, 1, 1), metadata.WithTMDBEpisode(1, 9, 9)} {
			meta, err := tmdbHint(t, theOffice(), "The.Office.S02E01.mkv", opt)
			if err != nil || len(meta) != 0 {
				t.Errorf("got %v, %v", meta, err)
			}
		}
	})
}

func Test_tmdb_hint_invalid(t *testing.T) {
	// A hint must have an identifier, and an episode a number
	for _, opt := range []metadata.Option{metadata.WithTMDBMovie(0), metadata.WithTMDBEpisode(0, 1, 1), metadata.WithTMDBEpisode(1, 1, 0)} {
		_, err := metadata.GetMetadata(context.Background(), strings.NewReader(""), "video/mp4", opt)
		if !errors.Is(err, gomedia.ErrBadParameter) {
			t.Errorf("got %v", err)
		}
	}
}
