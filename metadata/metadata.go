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
// extraction. If the third argument is a non-empty string, if should return
// a specific named metadata, namespace, or artwork. For example,
// "exif:" => return all EXIF metadata
// "DateTimeOriginal" => return any DateTimeOriginal tags
// "exif:DateTimeOriginal" => return the DateTimeOriginal EXIF tag
// "artwork:" => return all artwork metadata
// "artwork:thumbnail" => return the thumbnail artwork metadata
type HandlerFunc func(context.Context, io.Reader, string) ([]gomedia.Metadata, error)

type entry struct {
	re         *regexp.Regexp
	namespaces []string
	handler    HandlerFunc
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var handlerlock sync.Mutex
var handlers []entry
var cached = make(map[string][]entry)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Add a metadata handler for a given regular expression, along with the
// namespaces (e.g. "exif", "tiff") of metadata it can produce. A filter
// that requests a specific namespace (e.g. "tiff:Make") will only run
// handlers registered for that namespace.
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
// concurrently, and returns the combined metadata from all of them. If
// filter names a specific namespace (e.g. "tiff:" or "tiff:Make"), only
// the handlers registered for that namespace are run. Metadata from
// handlers that succeed is always returned, even if other handlers for
// the same content type fail; a non-nil error is the joined errors from
// any failing handlers, and should be treated as a warning rather than a
// reason to discard the metadata that was returned alongside it. ctx is
// passed to every handler and checked before any work starts, but
// GetMetadata otherwise waits for all handlers to finish rather than
// returning early on cancellation, since metadata extraction isn't
// preemptible and this avoids leaking their goroutines.
func GetMetadata(ctx context.Context, r io.Reader, contentType, filter string) ([]gomedia.Metadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries := getEntries(contentType)
	if len(entries) == 0 {
		return nil, gomedia.ErrNotFound.With("no handler for content type ", contentType)
	}

	// Narrow down to the handlers that can produce the requested namespace,
	// if any. A bare name (no namespace) or an empty filter can't be pruned,
	// since any handler's namespace could contain a matching name.
	var selected []HandlerFunc
	if namespace, ok := filterNamespace(filter); ok {
		for _, entry := range entries {
			if containsFold(entry.namespaces, namespace) {
				selected = append(selected, entry.handler)
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
	// data independently and concurrently
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}
	data := buf.Bytes()

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
			meta, err := handler(ctx, bytes.NewReader(data), filter)

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

// FilterMetadata returns the entries whose key matches filter, which may be
// "namespace:" (match all entries in the namespace), "name" (match entries
// with this name in any namespace), "namespace:name" (match entries with
// this name in this namespace), or empty (match everything). Keys are
// matched case-insensitively. This is a helper for handlers that build up
// a map of "namespace:name"-keyed metadata and need to apply the filter
// passed to their HandlerFunc.
func FilterMetadata(entries map[string]gomedia.Metadata, filter string) []gomedia.Metadata {
	filter = strings.ToLower(filter)
	filterNS, filterName, hasNamespace := strings.Cut(filter, ":")
	matches := func(m gomedia.Metadata) bool {
		if filter == "" {
			return true
		}
		namespace, name, _ := strings.Cut(m.Key(), ":")
		switch {
		case !hasNamespace:
			return strings.EqualFold(name, filterNS)
		case filterName == "":
			return strings.EqualFold(namespace, filterNS)
		default:
			return strings.EqualFold(namespace, filterNS) && strings.EqualFold(name, filterName)
		}
	}

	result := make([]gomedia.Metadata, 0, len(entries))
	for _, m := range entries {
		if matches(m) {
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

// filterNamespace returns the namespace requested by filter (e.g. "tiff"
// for "tiff:" or "tiff:Make"), and whether a namespace was actually
// specified, as opposed to a bare name (e.g. "Make") or an empty filter.
func filterNamespace(filter string) (string, bool) {
	namespace, _, hasNamespace := strings.Cut(filter, ":")
	if !hasNamespace || namespace == "" {
		return "", false
	}
	return namespace, true
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
