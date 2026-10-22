package video

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	metadata "github.com/mutablelogic/go-media/metadata"
)

// artworkStub answers TMDB's movie search and details, and its image server
// with a poster, but not a backdrop, and records the requests
type artworkStub struct {
	sync.Mutex
	poster   []byte
	requests []*http.Request
}

func (s *artworkStub) RoundTrip(req *http.Request) (*http.Response, error) {
	s.Lock()
	s.requests = append(s.requests, req)
	s.Unlock()
	status, contentType, body := http.StatusOK, "application/json", []byte(nil)
	switch {
	case req.URL.Host == "api.themoviedb.org" && req.URL.Path == "/3/search/movie":
		body = []byte(`{"page":1,"results":[{"id":603,"title":"The Matrix","original_title":"The Matrix","release_date":"1999-03-31","poster_path":"/poster.png","backdrop_path":"/backdrop.png"}],"total_pages":1,"total_results":1}`)
	case req.URL.Host == "api.themoviedb.org" && req.URL.Path == "/3/movie/603":
		body = []byte(`{"id":603,"runtime":136,"imdb_id":"tt0133093"}`)
	case req.URL.Host == "image.tmdb.org" && req.URL.Path == "/t/p/w780/poster.png":
		contentType, body = "image/png", s.poster
	default:
		status, body = http.StatusNotFound, []byte(`{"status_code":34,"status_message":"Not found"}`)
	}
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    req,
	}, nil
}

// images returns the requests for images
func (s *artworkStub) images() []*http.Request {
	s.Lock()
	defer s.Unlock()
	var result []*http.Request
	for _, req := range s.requests {
		if req.URL.Host == "image.tmdb.org" {
			result = append(result, req)
		}
	}
	return result
}

func Test_tmdb_artwork(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 10, 15))); err != nil {
		t.Fatal(err)
	}

	// match returns the metadata of the film, by key, in the namespaces.
	// The other handlers in the namespaces, such as the one for embedded
	// artwork, fail, as there's no video, which is a warning.
	match := func(t *testing.T, stub *artworkStub, namespaces ...string) map[string]any {
		t.Helper()
		meta, _ := metadata.GetMetadata(context.Background(), namedReader{Reader: strings.NewReader(""), name: "The.Matrix.1999.mkv"}, "video/x-matroska",
			metadata.WithTMDB("token", client.OptTransport(func(http.RoundTripper) http.RoundTripper { return stub })),
			metadata.WithNamespace(namespaces...),
			metadata.WithDuration(136*time.Minute),
		)
		result := make(map[string]any, len(meta))
		for _, m := range meta {
			result[m.Key()] = m.Any()
		}
		return result
	}

	t.Run("Artwork", func(t *testing.T) {
		// The poster is downloaded, without the API token, and a backdrop
		// which can't be downloaded is left out
		stub := &artworkStub{poster: buf.Bytes()}
		meta := match(t, stub, "tmdb", "artwork")
		if meta["tmdb:id"] != "603" {
			t.Fatalf("expected the film, got %v", meta)
		} else if _, ok := meta["artwork:poster"]; !ok {
			t.Errorf("expected the poster, got %v", meta)
		} else if _, ok := meta["artwork:backdrop"]; ok {
			t.Errorf("expected no backdrop, got %v", meta)
		}
		for _, req := range stub.images() {
			if req.Header.Get("Authorization") != "" {
				t.Errorf("the API token was sent to the image server: %v", req.URL)
			}
		}
		if len(stub.images()) != 2 {
			t.Errorf("expected the poster and backdrop to be requested, got %d requests", len(stub.images()))
		}
	})

	t.Run("NoArtwork", func(t *testing.T) {
		// Without the artwork namespace, no images are downloaded
		stub := &artworkStub{poster: buf.Bytes()}
		meta := match(t, stub, "tmdb")
		if meta["tmdb:id"] != "603" {
			t.Fatalf("expected the film, got %v", meta)
		} else if len(stub.images()) != 0 {
			t.Errorf("expected no images to be requested, got %d", len(stub.images()))
		}
	})
}
