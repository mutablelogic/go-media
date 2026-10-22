package tmdb_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	// Packages
	client "github.com/mutablelogic/go-client"
	tmdb "github.com/mutablelogic/go-media/tmdb/httpclient"
)

// imageStub answers every request with an image, and records the requests
type imageStub struct {
	requests []*http.Request
}

func (s *imageStub) RoundTrip(req *http.Request) (*http.Response, error) {
	s.requests = append(s.requests, req)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"image/jpeg"}},
		Body:       io.NopCloser(strings.NewReader("jpeg")),
		Request:    req,
	}, nil
}

func newImageClient(t *testing.T, opts ...client.ClientOpt) (*tmdb.Client, *imageStub) {
	t.Helper()
	stub := new(imageStub)
	c, err := tmdb.New("token", append([]client.ClientOpt{client.OptTransport(func(http.RoundTripper) http.RoundTripper { return stub })}, opts...)...)
	if err != nil {
		t.Fatal(err)
	}
	return c, stub
}

func Test_image_000(t *testing.T) {
	ctx := context.Background()

	t.Run("Image", func(t *testing.T) {
		// The image is from the image server, in the size, without the token
		c, stub := newImageClient(t)
		data, err := c.Image(ctx, tmdb.SizePoster, "/abc.jpg")
		if err != nil {
			t.Fatal(err)
		} else if string(data) != "jpeg" {
			t.Errorf("unexpected image %q", data)
		}
		req := stub.requests[0]
		if got := req.URL.String(); got != "https://image.tmdb.org/t/p/w780/abc.jpg" {
			t.Errorf("unexpected URL %q", got)
		} else if req.Header.Get("Authorization") != "" {
			t.Errorf("the API token was sent to the image server")
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		c, _ := newImageClient(t)
		for _, args := range [][2]string{{"", "/abc.jpg"}, {"w780/x", "/abc.jpg"}, {"w780", ""}, {"w780", "abc.jpg"}, {"w780", "/../abc.jpg"}} {
			if _, err := c.Image(ctx, args[0], args[1]); err == nil {
				t.Errorf("expected an error for %q", args)
			}
		}
	})
}

func Test_quota_000(t *testing.T) {
	ctx := context.Background()

	// elapsed returns how long the requests took
	elapsed := func(t *testing.T, c *tmdb.Client, n int) time.Duration {
		t.Helper()
		start := time.Now()
		for range n {
			if _, err := c.Image(ctx, tmdb.SizeStill, "/abc.jpg"); err != nil {
				t.Fatal(err)
			}
		}
		return time.Since(start)
	}

	t.Run("Quota", func(t *testing.T) {
		// A quota of 600 each minute spaces the requests a tenth of a
		// second apart
		c, _ := newImageClient(t, tmdb.OptQuota(600))
		if d := elapsed(t, c, 4); d < 250*time.Millisecond {
			t.Errorf("expected the requests to be spaced, took %v", d)
		}
	})

	t.Run("Ceiling", func(t *testing.T) {
		// A quota above what TMDB allows is limited to it, so requests are
		// spaced at least a fortieth of a second apart
		c, _ := newImageClient(t, tmdb.OptQuota(1_000_000))
		if d := elapsed(t, c, 5); d < 90*time.Millisecond {
			t.Errorf("expected the requests to be limited, took %v", d)
		}
	})
}
