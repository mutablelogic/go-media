package video

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	metadata "github.com/mutablelogic/go-media/metadata"
)

// tmdbStub answers TMDB's movie search with the movies, and the details of
// each movie with its runtime in minutes, and records the movies whose
// details are requested
type tmdbStub struct {
	runtimes map[uint64]uint64
	order    []uint64
	details  []string
	fail     bool
}

func (s *tmdbStub) RoundTrip(req *http.Request) (*http.Response, error) {
	status, body := http.StatusOK, ""
	switch {
	case req.URL.Path == "/3/search/movie":
		var results []string
		for _, id := range s.order {
			results = append(results, fmt.Sprintf(`{"id":%d,"title":"Movie %d","release_date":"1990-03-02","overview":"Overview %d"}`, id, id, id))
		}
		body = `{"page":1,"results":[` + strings.Join(results, ",") + `],"total_results":1,"total_pages":1}`
	case strings.HasPrefix(req.URL.Path, "/3/movie/"):
		id := strings.TrimPrefix(req.URL.Path, "/3/movie/")
		s.details = append(s.details, id)
		var n uint64
		fmt.Sscan(id, &n)
		if s.fail {
			status, body = http.StatusServiceUnavailable, `{"status_message":"offline"}`
		} else {
			body = fmt.Sprintf(`{"id":%d,"title":"Movie %d","runtime":%d,"imdb_id":"tt%07d"}`, n, n, s.runtimes[n], n)
		}
	default:
		status, body = http.StatusNotFound, `{}`
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

// tmdbMatch returns the TMDB metadata of a video with the duration, which is
// zero when it isn't known
func tmdbMatch(t *testing.T, stub *tmdbStub, duration time.Duration) (map[string]any, error) {
	t.Helper()
	opts := []metadata.Option{
		metadata.WithTMDB("token", client.OptTransport(func(http.RoundTripper) http.RoundTripper { return stub })),
		metadata.WithNamespace("tmdb"),
	}
	if duration > 0 {
		opts = append(opts, metadata.WithDuration(duration))
	}
	r := namedReader{Reader: strings.NewReader(""), name: "The Hunt for Red October (1990).mp4"}
	meta, err := metadata.GetMetadata(context.Background(), r, "video/mp4", opts...)
	result := make(map[string]any, len(meta))
	for _, m := range meta {
		result[m.Key()] = m.Any()
	}
	return result, err
}

func Test_tmdb_runtime(t *testing.T) {
	t.Run("Matches", func(t *testing.T) {
		// The first result whose runtime matches is taken, skipping one which
		// doesn't, and one whose runtime isn't known
		stub := &tmdbStub{order: []uint64{1, 2, 3, 4}, runtimes: map[uint64]uint64{1: 30, 2: 0, 3: 135, 4: 135}}
		meta, err := tmdbMatch(t, stub, 140*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:id"] != "3" || meta["tmdb:title"] != "Movie 3" {
			t.Errorf("matched %v, want movie 3", meta)
		}
		if meta["tmdb:runtime"] != 135*time.Minute {
			t.Errorf("tmdb:runtime = %v, want 2h15m", meta["tmdb:runtime"])
		}
		if meta["tmdb:imdbid"] != "tt0000003" {
			t.Errorf("tmdb:imdbid = %v, want tt0000003", meta["tmdb:imdbid"])
		}
		if got := strings.Join(stub.details, ","); got != "1,2,3" {
			t.Errorf("details of %s, want 1,2,3", got)
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		// A recording which isn't any of the movies has no metadata
		stub := &tmdbStub{order: []uint64{1, 2}, runtimes: map[uint64]uint64{1: 135, 2: 90}}
		meta, err := tmdbMatch(t, stub, 45*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if len(meta) != 0 {
			t.Errorf("expected no metadata, got %v", meta)
		}
	})

	t.Run("Boundary", func(t *testing.T) {
		// Within 10% of the runtime matches, either side, and more doesn't
		for duration, want := range map[time.Duration]bool{
			110 * time.Minute:                true,
			90 * time.Minute:                 true,
			110*time.Minute + time.Second:    false,
			90*time.Minute - time.Second:     false,
			100*time.Minute + 30*time.Second: true,
		} {
			stub := &tmdbStub{order: []uint64{1}, runtimes: map[uint64]uint64{1: 100}}
			meta, err := tmdbMatch(t, stub, duration)
			if err != nil {
				t.Fatal(err)
			}
			if got := meta["tmdb:id"] == "1"; got != want {
				t.Errorf("duration %v: matched = %v, want %v", duration, got, want)
			}
		}
	})

	t.Run("Candidates", func(t *testing.T) {
		// Only the top results are checked
		stub := &tmdbStub{order: []uint64{1, 2, 3, 4, 5, 6, 7}, runtimes: map[uint64]uint64{6: 100, 7: 100}}
		meta, err := tmdbMatch(t, stub, 100*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if len(meta) != 0 {
			t.Errorf("expected no metadata, got %v", meta)
		}
		if len(stub.details) != tmdbCandidates {
			t.Errorf("details of %d movies, want %d", len(stub.details), tmdbCandidates)
		}
	})

	t.Run("NoDuration", func(t *testing.T) {
		// Without the duration, the top result is taken, as before, without
		// its details
		stub := &tmdbStub{order: []uint64{1, 2}, runtimes: map[uint64]uint64{1: 30}}
		meta, err := tmdbMatch(t, stub, 0)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:id"] != "1" {
			t.Errorf("matched %v, want movie 1", meta)
		}
		for _, key := range []string{"tmdb:runtime", "tmdb:imdbid"} {
			if _, exists := meta[key]; exists {
				t.Errorf("unexpected %s", key)
			}
		}
		if len(stub.details) != 0 {
			t.Errorf("unexpected details of %v", stub.details)
		}
	})

	t.Run("DetailsFail", func(t *testing.T) {
		// An error getting the details is returned
		stub := &tmdbStub{order: []uint64{1}, fail: true}
		if _, err := tmdbMatch(t, stub, 100*time.Minute); err == nil {
			t.Error("expected an error")
		}
	})
}

func Test_tmdb_with_duration(t *testing.T) {
	if _, err := metadata.GetMetadata(context.Background(), strings.NewReader(""), "video/mp4", metadata.WithDuration(-time.Second)); err == nil {
		t.Error("expected an error for a negative duration")
	}
}
