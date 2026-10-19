package tmdb_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	// Packages
	client "github.com/mutablelogic/go-client"
	tmdb "github.com/mutablelogic/go-media/tmdb/httpclient"
	schema "github.com/mutablelogic/go-media/tmdb/schema"
)

// tvStub answers TMDB's TV endpoints with the responses, by path, and
// records the requests
type tvStub struct {
	responses map[string]string
	requests  []*http.Request
}

func (s *tvStub) RoundTrip(req *http.Request) (*http.Response, error) {
	s.requests = append(s.requests, req)
	status, body := http.StatusOK, s.responses[req.URL.Path]
	if body == "" {
		status, body = http.StatusNotFound, `{"status_code":34,"status_message":"The resource you requested could not be found."}`
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func newTVClient(t *testing.T) (*tmdb.Client, *tvStub) {
	t.Helper()
	stub := &tvStub{responses: map[string]string{
		"/3/search/tv":                  `{"page":1,"results":[{"id":1396,"name":"Breaking Bad","original_name":"Breaking Bad","first_air_date":"2008-01-20"}],"total_pages":1,"total_results":1}`,
		"/3/tv/1396":                    `{"id":1396,"name":"Breaking Bad","episode_run_time":[47],"number_of_seasons":5,"number_of_episodes":62,"external_ids":{"imdb_id":"tt0903747","tvdb_id":81189}}`,
		"/3/tv/1396/season/1/episode/2": `{"id":62086,"name":"Cat's in the Bag...","air_date":"2008-01-27","season_number":1,"episode_number":2,"runtime":49,"external_ids":{"imdb_id":"tt1054724","tvdb_id":349233}}`,
	}}
	c, err := tmdb.New("token", client.OptTransport(func(http.RoundTripper) http.RoundTripper { return stub }))
	if err != nil {
		t.Fatal(err)
	}
	return c, stub
}

func Test_tv_000(t *testing.T) {
	ctx := context.Background()

	t.Run("SearchTVShows", func(t *testing.T) {
		c, stub := newTVClient(t)
		resp, err := c.SearchTVShows(ctx, schema.TVSearchRequest{Term: "Breaking Bad"})
		if err != nil {
			t.Fatal(err)
		}
		if len(resp.Results) != 1 || resp.Results[0].ID != 1396 {
			t.Errorf("results = %+v", resp.Results)
		}
		req := stub.requests[0]
		if req.URL.Query().Get("query") != "Breaking Bad" || req.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("request = %s %v", req.URL, req.Header)
		}
	})

	t.Run("GetTVShow", func(t *testing.T) {
		// The show's details, with its external identifiers in one request
		c, stub := newTVClient(t)
		show, err := c.GetTVShow(ctx, 1396)
		if err != nil {
			t.Fatal(err)
		}
		if show.Name != "Breaking Bad" || len(show.EpisodeRunTime) != 1 || show.ExternalIDs == nil || show.ExternalIDs.IMDbID != "tt0903747" {
			t.Errorf("show = %+v", show)
		}
		if len(stub.requests) != 1 || stub.requests[0].URL.Query().Get("append_to_response") != "external_ids" {
			t.Errorf("requests = %v", stub.requests)
		}
	})

	t.Run("GetTVEpisode", func(t *testing.T) {
		// The episode, with its external identifiers in one request
		c, stub := newTVClient(t)
		episode, err := c.GetTVEpisode(ctx, 1396, 1, 2)
		if err != nil {
			t.Fatal(err)
		}
		if episode.Name != "Cat's in the Bag..." || episode.Runtime != 49 || episode.ExternalIDs == nil || episode.ExternalIDs.IMDbID != "tt1054724" {
			t.Errorf("episode = %+v", episode)
		}
		if len(stub.requests) != 1 || stub.requests[0].URL.Query().Get("append_to_response") != "external_ids" {
			t.Errorf("requests = %v", stub.requests)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		// An episode which doesn't exist is an error
		c, _ := newTVClient(t)
		if _, err := c.GetTVEpisode(ctx, 1396, 9, 99); err == nil {
			t.Error("expected an error")
		}
	})
}
