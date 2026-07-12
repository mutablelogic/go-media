package metadata_test

import (
	"context"
	"errors"
	"image"
	"io"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	. "github.com/mutablelogic/go-media/metadata"
)

////////////////////////////////////////////////////////////////////////////////
// FAKE METADATA

type fakeMetadata string

func (f fakeMetadata) Key() string        { return string(f) }
func (f fakeMetadata) Value() string      { return string(f) }
func (f fakeMetadata) Bytes() []byte      { return nil }
func (f fakeMetadata) Image() image.Image { return nil }
func (f fakeMetadata) Any() any           { return string(f) }

////////////////////////////////////////////////////////////////////////////////
// HELPERS

// markerHandler returns a HandlerFunc which records that it was called
// by setting *called to true, so a test can identify which of several
// registered handlers was actually returned/invoked.
func markerHandler(called *bool) HandlerFunc {
	return func(context.Context, io.Reader, string) ([]gomedia.Metadata, error) {
		*called = true
		return nil, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_metadata_000(t *testing.T) {
	// Registering a nil regular expression should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil regexp")
		}
	}()
	AddHandler(nil, markerHandler(new(bool)))
}

func Test_metadata_001(t *testing.T) {
	// Registering a nil handler should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil handler")
		}
	}()
	AddHandler(regexp.MustCompile("^x-test/001$"), nil)
}

func Test_metadata_002(t *testing.T) {
	// No handler registered for this content type should return no handlers
	if handlers := GetHandlers("x-test/002-no-such-type"); len(handlers) != 0 {
		t.Fatalf("expected no handlers for unregistered content type, got %d", len(handlers))
	}
}

func Test_metadata_003(t *testing.T) {
	// A registered handler should be returned for a matching content type
	var called bool
	AddHandler(regexp.MustCompile("^x-test/003$"), markerHandler(&called))

	handlers := GetHandlers("x-test/003")
	if len(handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(handlers))
	}
	if _, err := handlers[0](context.Background(), nil, ""); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected the registered handler to be invoked")
	}
}

func Test_metadata_004(t *testing.T) {
	// When multiple handlers match, all of them should be returned
	var a, b bool
	AddHandler(regexp.MustCompile("^x-test/004$"), markerHandler(&a))
	AddHandler(regexp.MustCompile("^x-test/004$"), markerHandler(&b))

	handlers := GetHandlers("x-test/004")
	if len(handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %d", len(handlers))
	}
	for _, h := range handlers {
		if _, err := h(context.Background(), nil, ""); err != nil {
			t.Fatal(err)
		}
	}
	if !a || !b {
		t.Fatal("expected both matching handlers to be invoked")
	}
}

// Regression test: a handler registered after a content type has already
// been resolved (and thus cached) must still be picked up on the next
// lookup, rather than the stale cached entry being returned forever.
func Test_metadata_005(t *testing.T) {
	var initial, later bool
	AddHandler(regexp.MustCompile("^x-test/005$"), markerHandler(&initial))

	// Prime the cache
	handlers := GetHandlers("x-test/005")
	if len(handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(handlers))
	}
	if _, err := handlers[0](context.Background(), nil, ""); err != nil {
		t.Fatal(err)
	}
	if !initial {
		t.Fatal("expected the initial handler to be invoked")
	}

	// Register another handler for the same content type
	AddHandler(regexp.MustCompile("^x-test/005$"), markerHandler(&later))

	handlers2 := GetHandlers("x-test/005")
	if len(handlers2) != 2 {
		t.Fatalf("expected 2 handlers after registering a new one, but the stale cache returned %d", len(handlers2))
	}
}

// Regression test: concurrent calls to GetHandlers (with AddHandler
// interleaved) must not race on the internal cache.
func Test_metadata_006(t *testing.T) {
	AddHandler(regexp.MustCompile("^x-test/006-a$"), markerHandler(new(bool)))
	AddHandler(regexp.MustCompile("^x-test/006-b$"), markerHandler(new(bool)))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			GetHandlers("x-test/006-a")
		}()
		go func() {
			defer wg.Done()
			GetHandlers("x-test/006-b")
		}()
	}
	wg.Wait()
}

// Regression test: GetMetadata should only run handlers registered for the
// namespace named by an explicit "namespace:" or "namespace:name" filter.
func Test_metadata_007(t *testing.T) {
	var tiffCalled, exifCalled bool
	AddHandler(regexp.MustCompile("^x-test/007$"), func(context.Context, io.Reader, string) ([]gomedia.Metadata, error) {
		tiffCalled = true
		return nil, nil
	}, "tiff")
	AddHandler(regexp.MustCompile("^x-test/007$"), func(context.Context, io.Reader, string) ([]gomedia.Metadata, error) {
		exifCalled = true
		return nil, nil
	}, "exif")

	if _, err := GetMetadata(context.Background(), strings.NewReader("data"), "x-test/007", "tiff:Make"); err != nil {
		t.Fatal(err)
	}
	if !tiffCalled {
		t.Fatal(`expected the "tiff" handler to be invoked for filter "tiff:Make"`)
	}
	if exifCalled {
		t.Fatal(`expected the "exif" handler NOT to be invoked for filter "tiff:Make"`)
	}
}

// A bare name filter (no namespace prefix) can't be pruned by namespace,
// since any handler's namespace could contain a tag with that name.
func Test_metadata_008(t *testing.T) {
	var tiffCalled, exifCalled bool
	AddHandler(regexp.MustCompile("^x-test/008$"), func(context.Context, io.Reader, string) ([]gomedia.Metadata, error) {
		tiffCalled = true
		return nil, nil
	}, "tiff")
	AddHandler(regexp.MustCompile("^x-test/008$"), func(context.Context, io.Reader, string) ([]gomedia.Metadata, error) {
		exifCalled = true
		return nil, nil
	}, "exif")

	if _, err := GetMetadata(context.Background(), strings.NewReader("data"), "x-test/008", "Make"); err != nil {
		t.Fatal(err)
	}
	if !tiffCalled || !exifCalled {
		t.Fatal("expected both handlers to be invoked for a bare-name filter")
	}
}

// A namespace filter that matches no registered handler's namespace should
// return no metadata and no error, since the content type itself is
// supported, just not by anything in that namespace.
func Test_metadata_009(t *testing.T) {
	AddHandler(regexp.MustCompile("^x-test/009$"), markerHandler(new(bool)), "tiff")

	meta, err := GetMetadata(context.Background(), strings.NewReader("data"), "x-test/009", "gps:Latitude")
	if err != nil {
		t.Fatal(err)
	}
	if len(meta) != 0 {
		t.Fatalf("expected no metadata, got %d", len(meta))
	}
}

// Regression test: if one handler fails and another succeeds for the same
// content type, GetMetadata should still return the metadata gathered from
// the successful handler, alongside the failing handler's error as a
// warning, rather than discarding it.
func Test_metadata_010(t *testing.T) {
	AddHandler(regexp.MustCompile("^x-test/010$"), func(context.Context, io.Reader, string) ([]gomedia.Metadata, error) {
		return nil, errors.New("boom")
	}, "broken")
	AddHandler(regexp.MustCompile("^x-test/010$"), func(context.Context, io.Reader, string) ([]gomedia.Metadata, error) {
		return []gomedia.Metadata{fakeMetadata("ok:value")}, nil
	}, "ok")

	meta, err := GetMetadata(context.Background(), strings.NewReader("data"), "x-test/010", "")
	if err == nil {
		t.Fatal("expected the failing handler's error to be returned")
	}
	if len(meta) != 1 || meta[0].Key() != "ok:value" {
		t.Fatalf("expected metadata from the successful handler despite the other failing, got %v", meta)
	}
}

// Regression test: an already-canceled context should make GetMetadata
// return immediately without invoking any handlers.
func Test_metadata_011(t *testing.T) {
	var called bool
	AddHandler(regexp.MustCompile("^x-test/011$"), markerHandler(&called), "x")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := GetMetadata(ctx, strings.NewReader("data"), "x-test/011", ""); err == nil {
		t.Fatal("expected an error for an already-canceled context")
	}
	if called {
		t.Fatal("did not expect any handler to be invoked for an already-canceled context")
	}
}

// Regression test: the context passed into GetMetadata should reach each
// handler unchanged, so a handler can honor cancellation/values itself if
// it does its own I/O or long-running work.
func Test_metadata_012(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "hello")

	var got any
	AddHandler(regexp.MustCompile("^x-test/012$"), func(ctx context.Context, _ io.Reader, _ string) ([]gomedia.Metadata, error) {
		got = ctx.Value(ctxKey{})
		return nil, nil
	}, "x")

	if _, err := GetMetadata(ctx, strings.NewReader("data"), "x-test/012", ""); err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("expected the context passed to GetMetadata to reach the handler, got %v", got)
	}
}

// Regression test: GetMetadata waits for every handler to finish even if
// ctx is canceled mid-flight, rather than abandoning them, since metadata
// extraction isn't preemptible and abandoning a handler goroutine would
// leak it.
func Test_metadata_013(t *testing.T) {
	var ran bool
	AddHandler(regexp.MustCompile("^x-test/013$"), func(_ context.Context, _ io.Reader, _ string) ([]gomedia.Metadata, error) {
		time.Sleep(20 * time.Millisecond)
		ran = true
		return nil, nil
	}, "x")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	if _, err := GetMetadata(ctx, strings.NewReader("data"), "x-test/013", ""); err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Fatal("expected GetMetadata to wait for the handler to finish despite mid-flight cancellation")
	}
}
