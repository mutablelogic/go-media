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
	titles   map[uint64]string // the title of a movie, or Movie and its id
	original map[uint64]string // the original title of a movie, or its title
	details  []string
	fail     bool
}

// title returns the title of a movie
func (s *tmdbStub) title(id uint64) string {
	if title, exists := s.titles[id]; exists {
		return title
	}
	return fmt.Sprintf("Movie %d", id)
}

func (s *tmdbStub) RoundTrip(req *http.Request) (*http.Response, error) {
	status, body := http.StatusOK, ""
	switch {
	case req.URL.Path == "/3/search/movie":
		var results []string
		for _, id := range s.order {
			original, exists := s.original[id]
			if !exists {
				original = s.title(id)
			}
			results = append(results, fmt.Sprintf(`{"id":%d,"title":%q,"original_title":%q,"release_date":"1990-03-02","overview":"Overview %d"}`, id, s.title(id), original, id))
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
	return tmdbMatchFile(t, stub, "The Hunt for Red October (1990).mp4", duration)
}

// tmdbMatchFile returns the TMDB metadata of a video with the file name and
// duration
func tmdbMatchFile(t *testing.T, stub *tmdbStub, name string, duration time.Duration) (map[string]any, error) {
	t.Helper()
	opts := []metadata.Option{
		metadata.WithTMDB("token", client.OptTransport(func(http.RoundTripper) http.RoundTripper { return stub })),
		metadata.WithNamespace("tmdb"),
	}
	if duration > 0 {
		opts = append(opts, metadata.WithDuration(duration))
	}
	r := namedReader{Reader: strings.NewReader(""), name: name}
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
		// A runtime within 10% of the duration matches, either side, and
		// further doesn't
		for runtime, want := range map[uint64]bool{
			90:  true,
			110: true,
			89:  false,
			111: false,
			100: true,
		} {
			stub := &tmdbStub{order: []uint64{1}, runtimes: map[uint64]uint64{1: runtime}}
			meta, err := tmdbMatch(t, stub, 100*time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			if got := meta["tmdb:id"] == "1"; got != want {
				t.Errorf("runtime %d minutes: matched = %v, want %v", runtime, got, want)
			}
		}

		// The duration is the baseline: a 90 minute stream doesn't match a
		// 100 minute movie, which is 11% longer
		stub := &tmdbStub{order: []uint64{1}, runtimes: map[uint64]uint64{1: 100}}
		if meta, err := tmdbMatch(t, stub, 90*time.Minute); err != nil {
			t.Fatal(err)
		} else if meta["tmdb:id"] == "1" {
			t.Error("a 90 minute stream matched a 100 minute movie")
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

func Test_tmdb_title(t *testing.T) {
	// The search for "Life" without a year, whose top results are two recent
	// films whose titles contain it, before Life (2017), and Life (1999)
	life := func() *tmdbStub {
		return &tmdbStub{
			order:    []uint64{1, 2, 3, 4, 5},
			titles:   map[uint64]string{1: "Love of Your Life", 2: "Doing Life", 3: "Life", 4: "The Life List", 5: "Life"},
			runtimes: map[uint64]uint64{1: 109, 2: 112, 3: 104, 4: 123, 5: 108},
		}
	}

	t.Run("ExactWithin5", func(t *testing.T) {
		// A film with the title, within 5%, before films within 10%
		meta, err := tmdbMatchFile(t, life(), "Life.mp4", 104*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:id"] != "3" {
			t.Errorf("matched %v, want Life (2017), which is 3", meta["tmdb:id"])
		}
	})

	t.Run("ExactInOrder", func(t *testing.T) {
		// Of two films with the title within 5%, the first
		meta, err := tmdbMatchFile(t, life(), "Life.mp4", 106*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:id"] != "3" {
			t.Errorf("matched %v, want 3", meta["tmdb:id"])
		}
	})

	t.Run("ExactOutside5", func(t *testing.T) {
		// When no film with the title is within 5%, the first within 10%
		stub := life()
		stub.runtimes = map[uint64]uint64{1: 109, 2: 112, 3: 92, 4: 123, 5: 93}
		meta, err := tmdbMatchFile(t, stub, "Life.mp4", 100*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:id"] != "1" {
			t.Errorf("matched %v, want 1", meta["tmdb:id"])
		}
	})

	t.Run("Punctuation", func(t *testing.T) {
		// A title matches in any case, without punctuation
		stub := &tmdbStub{
			order:    []uint64{1, 2},
			titles:   map[uint64]string{1: "Mission: Impossible - Fallout", 2: "Mission: Impossible"},
			runtimes: map[uint64]uint64{1: 110, 2: 110},
		}
		meta, err := tmdbMatchFile(t, stub, "mission impossible (1996).mp4", 110*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:id"] != "2" {
			t.Errorf("matched %v, want 2", meta["tmdb:id"])
		}
	})

	t.Run("Apostrophes", func(t *testing.T) {
		// A title with an apostrophe, straight or curly, is the name without
		// it, before an earlier film within 10% whose title only contains it
		for _, title := range []string{"Schindler's List", "Schindler\u2019s List"} {
			stub := &tmdbStub{
				order:    []uint64{1, 2},
				titles:   map[uint64]string{1: "Schindlers List: The Making Of", 2: title},
				runtimes: map[uint64]uint64{1: 190, 2: 195},
			}
			meta, err := tmdbMatchFile(t, stub, "Schindlers List.mp4", 195*time.Minute)
			if err != nil {
				t.Fatal(err)
			}
			if meta["tmdb:id"] != "2" {
				t.Errorf("%q: matched %v, want 2", title, meta["tmdb:id"])
			}
		}
	})

	t.Run("OriginalTitle", func(t *testing.T) {
		// A film's original title is its title too
		stub := &tmdbStub{
			order:    []uint64{1, 2},
			titles:   map[uint64]string{1: "Amelie From Montmartre", 2: "Amélie"},
			original: map[uint64]string{2: "Le Fabuleux Destin d'Amélie Poulain"},
			runtimes: map[uint64]uint64{1: 120, 2: 122},
		}
		meta, err := tmdbMatchFile(t, stub, "Le Fabuleux Destin d'Amélie Poulain.mkv", 122*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:id"] != "2" {
			t.Errorf("matched %v, want 2", meta["tmdb:id"])
		}
	})

	t.Run("ExactBeyondTop", func(t *testing.T) {
		// A film with the title is checked anywhere on the search's first
		// page, as many films share a common title, such as Ambition (1992),
		// which is 13th in its search, but only the top films are checked
		// when their titles aren't the name
		stub := &tmdbStub{titles: map[uint64]string{}, runtimes: map[uint64]uint64{}}
		for id := uint64(1); id <= 20; id++ {
			stub.order = append(stub.order, id)
			stub.titles[id] = fmt.Sprintf("Blonde Ambition %d", id)
			stub.runtimes[id] = 100
		}
		for id, runtime := range map[uint64]uint64{3: 23, 5: 87, 6: 3, 7: 73, 8: 6, 13: 9} {
			stub.titles[id], stub.runtimes[id] = "Ambition", runtime
		}
		stub.runtimes[10] = 9 // a title which isn't the name, beyond the top
		meta, err := tmdbMatchFile(t, stub, "/Movies/Ambition/Ambition.m4v", 535*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:id"] != "13" {
			t.Errorf("matched %v, want 13", meta["tmdb:id"])
		}

		// Without the film with the title, a film beyond the top isn't
		// checked for 10%
		stub.titles[13] = "Ambition Rising"
		meta, err = tmdbMatchFile(t, stub, "/Movies/Ambition/Ambition.m4v", 535*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if len(meta) != 0 {
			t.Errorf("matched %v, want nothing", meta["tmdb:id"])
		}
	})

	t.Run("DetailsOnce", func(t *testing.T) {
		// Each film's details are got once, over both passes
		stub := life()
		if _, err := tmdbMatchFile(t, stub, "Life.mp4", 45*time.Minute); err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(stub.details, ","); got != "3,5,1,2,4" {
			t.Errorf("details of %s, want 3,5,1,2,4", got)
		}
	})
}
