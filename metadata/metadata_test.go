package metadata_test

import (
	"context"
	"errors"
	"image"
	"io"
	"os"
	"path/filepath"
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
	return func(context.Context, io.Reader, *Opts) ([]gomedia.Metadata, error) {
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
	AddHandler(nil, "test", markerHandler(new(bool)))
}

func Test_metadata_001(t *testing.T) {
	// Registering a nil handler should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil handler")
		}
	}()
	AddHandler(regexp.MustCompile("^x-test/001$"), "test", nil)
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
	AddHandler(regexp.MustCompile("^x-test/003$"), "test", markerHandler(&called))

	handlers := GetHandlers("x-test/003")
	if len(handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(handlers))
	}
	if _, err := handlers[0](context.Background(), nil, &Opts{}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected the registered handler to be invoked")
	}
}

func Test_metadata_004(t *testing.T) {
	// When multiple handlers match, all of them should be returned
	var a, b bool
	AddHandler(regexp.MustCompile("^x-test/004$"), "a", markerHandler(&a))
	AddHandler(regexp.MustCompile("^x-test/004$"), "b", markerHandler(&b))

	handlers := GetHandlers("x-test/004")
	if len(handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %d", len(handlers))
	}
	for _, h := range handlers {
		if _, err := h(context.Background(), nil, &Opts{}); err != nil {
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
	AddHandler(regexp.MustCompile("^x-test/005$"), "initial", markerHandler(&initial))

	// Prime the cache
	handlers := GetHandlers("x-test/005")
	if len(handlers) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(handlers))
	}
	if _, err := handlers[0](context.Background(), nil, &Opts{}); err != nil {
		t.Fatal(err)
	}
	if !initial {
		t.Fatal("expected the initial handler to be invoked")
	}

	// Register another handler for the same content type
	AddHandler(regexp.MustCompile("^x-test/005$"), "later", markerHandler(&later))

	handlers2 := GetHandlers("x-test/005")
	if len(handlers2) != 2 {
		t.Fatalf("expected 2 handlers after registering a new one, but the stale cache returned %d", len(handlers2))
	}
}

// Regression test: concurrent calls to GetHandlers (with AddHandler
// interleaved) must not race on the internal cache.
func Test_metadata_006(t *testing.T) {
	AddHandler(regexp.MustCompile("^x-test/006-a$"), "a", markerHandler(new(bool)))
	AddHandler(regexp.MustCompile("^x-test/006-b$"), "b", markerHandler(new(bool)))

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

// Regression test: GetMetadata should only run handlers registered for a
// namespace explicitly requested via WithNamespace.
func Test_metadata_007(t *testing.T) {
	var tiffCalled, exifCalled bool
	AddHandler(regexp.MustCompile("^x-test/007$"), "tiff", func(context.Context, io.Reader, *Opts) ([]gomedia.Metadata, error) {
		tiffCalled = true
		return nil, nil
	}, "tiff")
	AddHandler(regexp.MustCompile("^x-test/007$"), "exif", func(context.Context, io.Reader, *Opts) ([]gomedia.Metadata, error) {
		exifCalled = true
		return nil, nil
	}, "exif")

	if _, err := GetMetadata(context.Background(), strings.NewReader("data"), "x-test/007", WithNamespace("tiff")); err != nil {
		t.Fatal(err)
	}
	if !tiffCalled {
		t.Fatal(`expected the "tiff" handler to be invoked for WithNamespace("tiff")`)
	}
	if exifCalled {
		t.Fatal(`expected the "exif" handler NOT to be invoked for WithNamespace("tiff")`)
	}
}

// Requesting multiple namespaces via WithNamespace should run every handler
// that matches any one of them.
func Test_metadata_008(t *testing.T) {
	var tiffCalled, exifCalled bool
	AddHandler(regexp.MustCompile("^x-test/008$"), "tiff", func(context.Context, io.Reader, *Opts) ([]gomedia.Metadata, error) {
		tiffCalled = true
		return nil, nil
	}, "tiff")
	AddHandler(regexp.MustCompile("^x-test/008$"), "exif", func(context.Context, io.Reader, *Opts) ([]gomedia.Metadata, error) {
		exifCalled = true
		return nil, nil
	}, "exif")

	if _, err := GetMetadata(context.Background(), strings.NewReader("data"), "x-test/008", WithNamespace("tiff", "exif")); err != nil {
		t.Fatal(err)
	}
	if !tiffCalled || !exifCalled {
		t.Fatal("expected both handlers to be invoked when both their namespaces are requested")
	}
}

// A namespace filter that matches no registered handler's namespace should
// return no metadata and no error, since the content type itself is
// supported, just not by anything in that namespace.
func Test_metadata_009(t *testing.T) {
	AddHandler(regexp.MustCompile("^x-test/009$"), "tiff", markerHandler(new(bool)), "tiff")

	meta, err := GetMetadata(context.Background(), strings.NewReader("data"), "x-test/009", WithNamespace("gps"))
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
	AddHandler(regexp.MustCompile("^x-test/010$"), "broken", func(context.Context, io.Reader, *Opts) ([]gomedia.Metadata, error) {
		return nil, errors.New("boom")
	}, "broken")
	AddHandler(regexp.MustCompile("^x-test/010$"), "ok", func(context.Context, io.Reader, *Opts) ([]gomedia.Metadata, error) {
		return []gomedia.Metadata{fakeMetadata("ok:value")}, nil
	}, "ok")

	meta, err := GetMetadata(context.Background(), strings.NewReader("data"), "x-test/010")
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
	AddHandler(regexp.MustCompile("^x-test/011$"), "x", markerHandler(&called), "x")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := GetMetadata(ctx, strings.NewReader("data"), "x-test/011"); err == nil {
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
	AddHandler(regexp.MustCompile("^x-test/012$"), "x", func(ctx context.Context, _ io.Reader, _ *Opts) ([]gomedia.Metadata, error) {
		got = ctx.Value(ctxKey{})
		return nil, nil
	}, "x")

	if _, err := GetMetadata(ctx, strings.NewReader("data"), "x-test/012"); err != nil {
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
	AddHandler(regexp.MustCompile("^x-test/013$"), "x", func(_ context.Context, _ io.Reader, _ *Opts) ([]gomedia.Metadata, error) {
		time.Sleep(20 * time.Millisecond)
		ran = true
		return nil, nil
	}, "x")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	if _, err := GetMetadata(ctx, strings.NewReader("data"), "x-test/013"); err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Fatal("expected GetMetadata to wait for the handler to finish despite mid-flight cancellation")
	}
}

// A file is read by each handler where it is, from its current position,
// rather than being read into memory, and each handler can seek in it
func Test_metadata_014(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("skip:data"), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Seek(5, io.SeekStart); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var got []string
	handler := func(_ context.Context, r io.Reader, _ *Opts) ([]gomedia.Metadata, error) {
		seeker, ok := r.(io.ReadSeeker)
		if !ok {
			return nil, errors.New("expected an io.ReadSeeker")
		}
		if _, err := seeker.Seek(1, io.SeekStart); err != nil {
			return nil, err
		}
		data, err := io.ReadAll(seeker)
		if err != nil {
			return nil, err
		}
		name := ""
		if named, ok := r.(gomedia.NamedReader); ok {
			name = named.Name()
		}
		mu.Lock()
		defer mu.Unlock()
		got = append(got, string(data)+"@"+name)
		return nil, nil
	}
	AddHandler(regexp.MustCompile("^x-test/014$"), "a", handler, "a")
	AddHandler(regexp.MustCompile("^x-test/014$"), "b", handler, "b")

	if _, err := GetMetadata(context.Background(), f, "x-test/014"); err != nil {
		t.Fatal(err)
	}
	for _, g := range got {
		if g != "ata@"+path {
			t.Errorf("expected each handler to read the file from its position, got %q", g)
		}
	}
	if len(got) != 2 {
		t.Fatalf("expected two handlers to run, got %d", len(got))
	}

	// The file isn't read, so it's where it was
	if pos, err := f.Seek(0, io.SeekCurrent); err != nil {
		t.Fatal(err)
	} else if pos != 5 {
		t.Fatalf("expected the file to be at 5, got %d", pos)
	}
}

// A reader which can't be read at an offset is read into memory, and each
// handler can still seek in it
func Test_metadata_015(t *testing.T) {
	var got string
	AddHandler(regexp.MustCompile("^x-test/015$"), "x", func(_ context.Context, r io.Reader, _ *Opts) ([]gomedia.Metadata, error) {
		seeker, ok := r.(io.ReadSeeker)
		if !ok {
			return nil, errors.New("expected an io.ReadSeeker")
		}
		if _, err := seeker.Seek(2, io.SeekStart); err != nil {
			return nil, err
		}
		data, err := io.ReadAll(seeker)
		got = string(data)
		return nil, err
	}, "x")

	if _, err := GetMetadata(context.Background(), io.MultiReader(strings.NewReader("da"), strings.NewReader("ta")), "x-test/015"); err != nil {
		t.Fatal(err)
	}
	if got != "ta" {
		t.Fatalf("expected the handler to read the rest of the data, got %q", got)
	}
}
