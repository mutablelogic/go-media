package metadata

import (
	"bytes"
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"sync"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// HandlerFunc is a function that can be used to extract metadata from a given
// reader. The context can be used to cancel or time out long-running
// extraction. Opts carries the options passed to GetMetadata; in particular
// Opts.HasNamespace can be used to check whether a specific namespace (e.g.
// "artwork") was requested via WithNamespace, which handlers that do
// expensive work for a single namespace should use to opt in only when that
// namespace was explicitly requested. A handler that produces several
// namespaces can use FilterMetadata to return only the ones requested.
type HandlerFunc func(context.Context, io.Reader, *Opts) ([]gomedia.Metadata, error)

type entry struct {
	re         *regexp.Regexp
	namespaces []string
	handler    HandlerFunc
}

// namedReader wraps a *bytes.Reader with a Name(), so a handler can recover
// the original input's filename (via gomedia.NamedReader) after GetMetadata
// has buffered it into a fresh reader for each handler.
type namedReader struct {
	*bytes.Reader
	name string
}

func (r namedReader) Name() string { return r.name }

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var handlerlock sync.Mutex
var handlers []entry
var cached = make(map[string][]entry)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Add a metadata handler for a given regular expression, along with the
// namespaces (e.g. "exif", "tiff") of metadata it can produce. A caller
// that requests specific namespaces via WithNamespace will only run
// handlers registered for one of those namespaces.
func AddHandler(re *regexp.Regexp, fn HandlerFunc, namespaces ...string) {
	if re == nil || fn == nil {
		panic(gomedia.ErrBadParameter.With("nil regex or handler"))
	}
	handlerlock.Lock()
	defer handlerlock.Unlock()
	handlers = append(handlers, entry{re: re, namespaces: namespaces, handler: fn})
	cached = make(map[string][]entry)
}

// GetHandlers returns all handlers registered for a given content type, or
// nil if no handler is registered for that content type.
func GetHandlers(contentType string) []HandlerFunc {
	entries := getEntries(contentType)
	if len(entries) == 0 {
		return nil
	}
	fns := make([]HandlerFunc, len(entries))
	for i, entry := range entries {
		fns[i] = entry.handler
	}
	return fns
}

// GetMetadata runs every handler registered for contentType against r,
// concurrently, and returns the combined metadata from all of them.
// Metadata from handlers that succeed is always returned, even if other
// handlers for the same content type fail; a non-nil error is the joined
// errors from any failing handlers, and should be treated as a warning rather than a
// reason to discard the metadata that was returned alongside it. ctx is
// passed to every handler and checked before any work starts, but
// GetMetadata otherwise waits for all handlers to finish rather than
// returning early on cancellation.
func GetMetadata(ctx context.Context, r io.Reader, contentType string, opts ...Option) ([]gomedia.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries := getEntries(contentType)
	if len(entries) == 0 {
		return nil, gomedia.ErrNotImplemented.With("no handler for content type ", contentType)
	}

	// Apply options
	o, err := applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	// Narrow down to the handlers that can produce one of the requested
	// namespaces, if any were given via WithNamespace. No namespaces
	// requested means no pruning, since any handler could be relevant.
	var selected []HandlerFunc
	if len(o.namespaces) > 0 {
		for _, entry := range entries {
			for _, namespace := range o.namespaces {
				if containsFold(entry.namespaces, namespace) {
					selected = append(selected, entry.handler)
					break
				}
			}
		}
	} else {
		selected = make([]HandlerFunc, len(entries))
		for i, entry := range entries {
			selected[i] = entry.handler
		}
	}
	if len(selected) == 0 {
		return nil, nil
	}

	// Tee the reader into a buffer once, so every handler can read the
	// data independently and concurrently. If r carries a name (see
	// gomedia.NamedReader), preserve it on each handler's copy, since a
	// handler may need the filename itself (e.g. to look up TMDB metadata)
	// rather than the file's contents.
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}
	data := buf.Bytes()
	named, _ := r.(gomedia.NamedReader)

	newReader := func() io.Reader {
		br := bytes.NewReader(data)
		if named == nil {
			return br
		}
		return namedReader{br, named.Name()}
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allMeta []gomedia.Metadata
		errs    error
	)
	wg.Add(len(selected))
	for _, handler := range selected {
		go func(handler HandlerFunc) {
			defer wg.Done()
			meta, err := handler(ctx, newReader(), o)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = errors.Join(errs, err)
				return
			}
			allMeta = append(allMeta, meta...)
		}(handler)
	}

	wg.Wait()

	return allMeta, errs
}

// FilterMetadata returns the entries whose namespace was requested via o
// (see WithNamespace), or every entry if no namespace was requested. Keys
// are of the form "namespace:name" and matched case-insensitively. This is
// a helper for handlers that build up a map of "namespace:name"-keyed
// metadata and need to apply the namespace filter passed to their
// HandlerFunc.
func FilterMetadata(entries map[string]gomedia.Metadata, o *Opts) []gomedia.Metadata {
	result := make([]gomedia.Metadata, 0, len(entries))
	for key, m := range entries {
		if len(o.namespaces) == 0 {
			result = append(result, m)
			continue
		}
		namespace, _, _ := strings.Cut(key, ":")
		if containsFold(o.namespaces, namespace) {
			result = append(result, m)
		}
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// getEntries returns all registered entries matching contentType, caching
// the result until the next AddHandler call.
func getEntries(contentType string) []entry {
	handlerlock.Lock()
	defer handlerlock.Unlock()

	// Check the cache first
	if entries, ok := cached[contentType]; ok {
		return entries
	}

	// Make a list of the handlers that match the content type
	var matches []entry
	for _, entry := range handlers {
		if entry.re.MatchString(contentType) {
			matches = append(matches, entry)
		}
	}

	cached[contentType] = matches
	return matches
}

// containsFold reports whether namespaces contains s, case-insensitively.
func containsFold(namespaces []string, s string) bool {
	for _, namespace := range namespaces {
		if strings.EqualFold(namespace, s) {
			return true
		}
	}
	return false
}
