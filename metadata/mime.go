package metadata

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	_ "github.com/mutablelogic/go-media/pkg/raw"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type NamedStream interface {
	// Name returns the name (path) of the stream, which can be used to
	// determine the MIME type.
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Type returns the MIME type of the given file, along with a map of any additional
// metadata that was extracted from the file. If the MIME type cannot be determined,
// an error is returned.
func ContentType(r io.Reader) (string, map[string]string, error) {
	if r == nil {
		return "", nil, gomedia.ErrBadParameter.With("nil reader")
	}

	// Try the http.DetectContentType function first
	buf := make([]byte, 512)
	if _, err := r.Read(buf); err != nil && !errors.Is(err, io.EOF) {
		return "", nil, gomedia.ErrInternalError.With(err.Error())
	}
	if mediaType := http.DetectContentType(buf); mediaType != types.ContentTypeBinary {
		return mime.ParseMediaType(mediaType)
	}

	// By extension second
	if named, ok := r.(NamedStream); ok {
		if ext := filepath.Ext(named.Name()); ext != "" {
			if mediaType := mime.TypeByExtension(ext); mediaType != "" {
				return mime.ParseMediaType(mediaType)
			}
		}
	}

	// Unknown type
	return types.ContentTypeBinary, nil, nil
}
