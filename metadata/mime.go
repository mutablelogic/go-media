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
	// Authoritative means byte-sniffing gets this extension wrong or
	// leaves it ambiguous (e.g. .m4a sniffs as video/mp4), so the
	// extension itself is trusted immediately, without reading the stream
	// at all. Preferred alone (the common case) only affects
	// ExtensionByType's reverse mapping and must NOT skip sniffing - a
	// .jpg's actual byte signature is reliable, so there's no reason to
	// trust a mislabeled or corrupted file over actually checking it.
	Authoritative bool
}

// extensionContentTypes stores extension -> media type mappings and whether
// this extension should be preferred when a type has multiple aliases.
// It is used both for content type detection and extension selection.
var extensionContentTypes = map[string]extensionRule{
	".m4a":  {ContentType: "audio/mp4", Preferred: true, Authoritative: true},
	".jpg":  {ContentType: "image/jpeg", Preferred: true},
	".jpeg": {ContentType: "image/jpeg"},
	// Markdown sniffs as text/plain, so its extension refines it (see
	// ContentType), but isn't Authoritative, so binary content named .md
	// is still detected from its bytes
	".md":       {ContentType: "text/markdown", Preferred: true},
	".markdown": {ContentType: "text/markdown"},
	// Documents which are zip archives sniff as application/zip, so their
	// extensions refine it (see ContentType), and they're listed here as
	// the system's types often don't include them
	".docx": {ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Preferred: true},
	".xlsx": {ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", Preferred: true},
	".pptx": {ContentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation", Preferred: true},
	".odt":  {ContentType: "application/vnd.oasis.opendocument.text", Preferred: true},
	".ods":  {ContentType: "application/vnd.oasis.opendocument.spreadsheet", Preferred: true},
	".odp":  {ContentType: "application/vnd.oasis.opendocument.presentation", Preferred: true},
	".epub": {ContentType: "application/epub+zip", Preferred: true},
}

// zipTypes are the prefixes and suffixes of the content types of documents
// which are zip archives, except the content types with the excluded
// suffixes, such as OpenDocument's flat XML documents (.fodt), which are XML
var (
	zipTypePrefixes  = []string{"application/vnd.openxmlformats-officedocument.", "application/vnd.oasis.opendocument."}
	zipTypeSuffixes  = []string{"+zip"}
	zipTypeExcluding = []string{"-flat-xml"}
)

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
			if forced.Authoritative {
				return mime.ParseMediaType(forced.ContentType)
			}
			// Preferred without Authoritative (e.g. .jpg/.jpeg): use this as
			// the extension-based fallback below, but still sniff first -
			// unlike .m4a, there's no reason not to. Extensions not in this
			// map at all (e.g. .ts, which is either TypeScript source or an
			// MPEG transport stream) go through mime.TypeByExtension the
			// same way.
			extType = forced.ContentType
		} else if ext != "" {
			extType = mime.TypeByExtension(ext)
		}
	}

	// Try the http.DetectContentType function first
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", nil, gomedia.ErrInternalError.With(err.Error())
	}
	if mediaType := http.DetectContentType(buf[:n]); mediaType != types.ContentTypeBinary {
		contentType, params, err := mime.ParseMediaType(mediaType)
		if err != nil {
			return "", nil, err
		}
		// Plain text is refined by an extension with a more specific text
		// type, such as text/markdown for .md, keeping the sniffed
		// parameters, such as its charset, and a zip archive by an extension
		// of a document which is a zip archive, such as .docx
		if extType != "" {
			if extContentType, _, err := mime.ParseMediaType(extType); err == nil && refines(contentType, extContentType) {
				return extContentType, params, nil
			}
		}
		return contentType, params, nil
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

// refines returns true when the content type of an extension is a more
// specific type of the sniffed content type: a text type for plain text, or a
// document which is a zip archive, and not a flat XML document, for a zip
// archive
func refines(sniffed, ext string) bool {
	switch sniffed {
	case "text/plain":
		return strings.HasPrefix(ext, "text/") && ext != sniffed
	case "application/zip":
		for _, suffix := range zipTypeExcluding {
			if strings.HasSuffix(ext, suffix) {
				return false
			}
		}
		for _, prefix := range zipTypePrefixes {
			if strings.HasPrefix(ext, prefix) {
				return true
			}
		}
		for _, suffix := range zipTypeSuffixes {
			if strings.HasSuffix(ext, suffix) {
				return true
			}
		}
	}
	return false
}
