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

// tvShow is a show in tvStub, with the runtimes of its episodes, by season
// and episode, such as "1/2", which is zero when TMDB doesn't know it, and
// the typical runtime of its episodes
type tvShow struct {
	id       uint64
	name     string
	runtimes map[string]uint64
	typical  uint64
}

// tvStub answers TMDB's TV search with the shows, and the episodes and
// details of each, and records the requests' paths
type tvStub struct {
	shows    []tvShow
	fail     bool
	requests []string
}

func (s *tvStub) RoundTrip(req *http.Request) (*http.Response, error) {
	s.requests = append(s.requests, req.URL.Path)
	status, body := http.StatusNotFound, `{"status_code":34}`
	switch {
	case req.URL.Path == "/3/search/tv":
		var results []string
		for _, show := range s.shows {
			results = append(results, fmt.Sprintf(`{"id":%d,"name":%q,"original_name":%q}`, show.id, show.name, show.name))
		}
		status, body = http.StatusOK, `{"page":1,"results":[`+strings.Join(results, ",")+`]}`
	case s.fail:
		status, body = http.StatusServiceUnavailable, `{"status_message":"offline"}`
	default:
		for _, show := range s.shows {
			prefix := fmt.Sprintf("/3/tv/%d", show.id)
			if req.URL.Path == prefix {
				status, body = http.StatusOK, fmt.Sprintf(`{"id":%d,"name":%q,"episode_run_time":[%d]}`, show.id, show.name, show.typical)
			} else if rest, found := strings.CutPrefix(req.URL.Path, prefix+"/season/"); found {
				var season, episode uint64
				fmt.Sscanf(strings.Replace(rest, "/episode/", " ", 1), "%d %d", &season, &episode)
				if runtime, exists := show.runtimes[fmt.Sprintf("%d/%d", season, episode)]; exists {
					status, body = http.StatusOK, fmt.Sprintf(`{"id":%d%d%d,"name":"Episode %d","overview":"Overview","air_date":"2005-09-20","season_number":%d,"episode_number":%d,"runtime":%d,"external_ids":{"imdb_id":"tt%07d"}}`, show.id, season, episode, episode, season, episode, runtime, show.id)
				}
			}
		}
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

// tvMatch returns the TMDB metadata of an episode in the file, with the
// duration, which is zero when it isn't known
func tvMatch(t *testing.T, stub *tvStub, name string, duration time.Duration) (map[string]any, error) {
	t.Helper()
	opts := []metadata.Option{
		metadata.WithTMDB("token", client.OptTransport(func(http.RoundTripper) http.RoundTripper { return stub })),
		metadata.WithNamespace("tmdb"),
	}
	if duration > 0 {
		opts = append(opts, metadata.WithDuration(duration))
	}
	meta, err := metadata.GetMetadata(context.Background(), namedReader{Reader: strings.NewReader(""), name: name}, "video/x-matroska", opts...)
	result := make(map[string]any, len(meta))
	for _, m := range meta {
		result[m.Key()] = m.Any()
	}
	return result, err
}

// theOffice is the search for The Office, whose top results are the US
// show, then the UK one, both with the name, and another show
func theOffice() *tvStub {
	return &tvStub{shows: []tvShow{
		{id: 1, name: "The Office", runtimes: map[string]uint64{"2/1": 21}},
		{id: 2, name: "The Office", runtimes: map[string]uint64{"2/1": 30}},
		{id: 3, name: "The Office Mix-Up", runtimes: map[string]uint64{"2/1": 29}},
	}}
}

func Test_tmdb_tv(t *testing.T) {
	t.Run("Entries", func(t *testing.T) {
		// The episode's metadata, with a title of the show, the season and
		// episode, and the episode's name
		meta, err := tvMatch(t, theOffice(), "The.Office.S02E01.mkv", 21*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		for key, want := range map[string]any{
			"tmdb:showid":      "1",
			"tmdb:show":        "The Office",
			"tmdb:season":      uint64(2),
			"tmdb:episode":     uint64(1),
			"tmdb:title":       "Episode 1",
			"tmdb:releasedate": "2005-09-20",
			"tmdb:runtime":     21 * time.Minute,
			"tmdb:imdbid":      "tt0000001",
		} {
			if meta[key] != want {
				t.Errorf("%s = %#v, want %#v", key, meta[key], want)
			}
		}
	})

	t.Run("RuntimeChoosesShow", func(t *testing.T) {
		// Of two shows with the name, the one whose episode is within 5%,
		// although it's second
		meta, err := tvMatch(t, theOffice(), "The.Office.S02E01.mkv", 29*time.Minute+30*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:showid"] != "2" {
			t.Errorf("matched %v, want the second show", meta["tmdb:showid"])
		}
	})

	t.Run("NotExact", func(t *testing.T) {
		// When no show with the name is within 5%, the first within 10%
		stub := theOffice()
		stub.shows[1].runtimes["2/1"] = 33
		meta, err := tvMatch(t, stub, "The.Office.S02E01.mkv", 29*time.Minute+30*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:showid"] != "3" {
			t.Errorf("matched %v, want the third show", meta["tmdb:showid"])
		}
	})

	t.Run("NoEpisode", func(t *testing.T) {
		// A show without the episode isn't a match, which isn't an error
		stub := theOffice()
		delete(stub.shows[0].runtimes, "2/1")
		meta, err := tvMatch(t, stub, "The.Office.S02E01.mkv", 30*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:showid"] != "2" {
			t.Errorf("matched %v, want the second show", meta["tmdb:showid"])
		}
		if meta, err := tvMatch(t, theOffice(), "The.Office.S09E99.mkv", 30*time.Minute); err != nil || len(meta) != 0 {
			t.Errorf("meta = %v, err = %v, want nothing", meta, err)
		}
	})

	t.Run("TypicalRuntime", func(t *testing.T) {
		// When TMDB doesn't know an episode's runtime, it's the show's
		// typical one
		stub := &tvStub{shows: []tvShow{{id: 1, name: "The Office", runtimes: map[string]uint64{"2/1": 0}, typical: 22}}}
		meta, err := tvMatch(t, stub, "The.Office.S02E01.mkv", 22*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:runtime"] != 22*time.Minute {
			t.Errorf("tmdb:runtime = %v, want 22m", meta["tmdb:runtime"])
		}
	})

	t.Run("NoMatch", func(t *testing.T) {
		// A clip which isn't any episode has no metadata
		meta, err := tvMatch(t, theOffice(), "The.Office.S02E01.mkv", 5*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if len(meta) != 0 {
			t.Errorf("meta = %v, want nothing", meta)
		}
	})

	t.Run("NoDuration", func(t *testing.T) {
		// Without the duration, the first show with the name and the episode
		meta, err := tvMatch(t, theOffice(), "The.Office.S02E01.mkv", 0)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:showid"] != "1" {
			t.Errorf("matched %v, want the first show", meta["tmdb:showid"])
		}
	})

	t.Run("Fail", func(t *testing.T) {
		// An error getting an episode, other than it not being found, is
		// returned
		stub := theOffice()
		stub.fail = true
		if _, err := tvMatch(t, stub, "The.Office.S02E01.mkv", 21*time.Minute); err == nil {
			t.Error("expected an error")
		}
	})

	t.Run("ExactBeyondTop", func(t *testing.T) {
		// A show with the name is checked anywhere on the search's first page
		stub := &tvStub{}
		for id := uint64(1); id <= 8; id++ {
			stub.shows = append(stub.shows, tvShow{id: id, name: fmt.Sprintf("The Office Show %d", id), runtimes: map[string]uint64{"2/1": 50}})
		}
		stub.shows[7] = tvShow{id: 8, name: "The Office", runtimes: map[string]uint64{"2/1": 22}}
		meta, err := tvMatch(t, stub, "The.Office.S02E01.mkv", 22*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if meta["tmdb:showid"] != "8" {
			t.Errorf("matched %v, want 8", meta["tmdb:showid"])
		}
	})

	t.Run("EpisodesOnce", func(t *testing.T) {
		// Each show's episode is got once, over both passes
		stub := theOffice()
		if _, err := tvMatch(t, stub, "The.Office.S02E01.mkv", 5*time.Minute); err != nil {
			t.Fatal(err)
		}
		if got := strings.Join(stub.requests, ","); got != "/3/search/tv,/3/tv/1/season/2/episode/1,/3/tv/2/season/2/episode/1,/3/tv/3/season/2/episode/1" {
			t.Errorf("requests = %s", got)
		}
	})
}
