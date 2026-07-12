package metadata

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

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

type extensionRule struct {
	ContentType string
	Preferred   bool
}

// extensionContentTypes stores extension -> media type mappings and whether
// this extension should be preferred when a type has multiple aliases.
// It is used both for content type detection and extension selection.
var extensionContentTypes = map[string]extensionRule{
	".m4a":  {ContentType: "audio/mp4", Preferred: true},
	".jpg":  {ContentType: "image/jpeg", Preferred: true},
	".jpeg": {ContentType: "image/jpeg"},
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

	var extType string
	if named, ok := r.(NamedStream); ok {
		ext := strings.ToLower(filepath.Ext(named.Name()))
		if forced, ok := extensionContentTypes[ext]; ok {
			extType = forced.ContentType
		} else if ext != "" {
			extType = mime.TypeByExtension(ext)
		}
	}

	// Try the http.DetectContentType function first
	buf := make([]byte, 512)
	if _, err := r.Read(buf); err != nil && !errors.Is(err, io.EOF) {
		return "", nil, gomedia.ErrInternalError.With(err.Error())
	}
	if mediaType := http.DetectContentType(buf); mediaType != types.ContentTypeBinary {
		// Extension-based override for known cases like .m4a, where MP4 byte
		// signatures are otherwise reported as video/mp4.
		if extType != "" {
			if mediaType == "video/mp4" || mediaType == "application/mp4" {
				return mime.ParseMediaType(extType)
			}
		}
		return mime.ParseMediaType(mediaType)
	}

	// By extension second
	if extType != "" {
		return mime.ParseMediaType(extType)
	}

	// Unknown type
	return types.ContentTypeBinary, nil, nil
}

// ExtensionByType returns the preferred extension for contentType, falling
// back to the first registered extension if no preferred extension matches.
func ExtensionByType(contentType string) string {
	for ext, rule := range extensionContentTypes {
		if rule.ContentType == contentType && rule.Preferred {
			return ext
		}
	}

	exts, err := mime.ExtensionsByType(contentType)
	if err != nil || len(exts) == 0 {
		return ""
	}

	return exts[0]
}
