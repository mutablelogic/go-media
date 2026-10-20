package metadata

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"sync"

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
	".m4p":  {ContentType: "audio/mp4", Authoritative: true}, // iTunes and Apple Music protected audio
	".jpg":  {ContentType: "image/jpeg", Preferred: true},
	".jpeg": {ContentType: "image/jpeg"},
	".tga":  {ContentType: "image/x-tga", Preferred: true},

	// Playlists, which are text, have different types on each platform
	".m3u":  {ContentType: "audio/x-mpegurl", Preferred: true},
	".m3u8": {ContentType: "application/vnd.apple.mpegurl", Preferred: true},
	".pls":  {ContentType: "audio/x-scpls", Preferred: true},
}

// textExtensionTypes are the types of text, by extension, which only refine
// content sniffed as plain text, and are never the type of content by its
// extension alone, as content of another type has the extension too, such
// as .ts, which is TypeScript, or an MPEG transport stream
// (see AddTextExtensionType)
var (
	textExtensionTypes   = map[string]string{}
	textExtensionTypesMu sync.RWMutex
)

// iconTypes are the content types of an icon, which differ on each platform,
// so an icon isn't refined by another type of icon
var iconTypes = []string{"image/x-icon", "image/vnd.microsoft.icon"}

// textTypes are the content types of text documents which aren't text or
// application types, such as playlists, so they refine plain text
var textTypes = []string{"audio/x-mpegurl", "audio/mpegurl", "audio/x-scpls"}

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

	var extType, textType string
	if named, ok := r.(NamedStream); ok {
		ext := strings.ToLower(filepath.Ext(named.Name()))
		textType = textExtensionType(ext)
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
		// Plain text is refined by the type of text of its extension first,
		// such as text/x-typescript for .ts, which is otherwise a video
		if contentType == "text/plain" && textType != "" {
			return textType, params, nil
		}
		// Plain text, or a zip archive, is refined by the type of its
		// extension, when that's more specific (see refines), such as
		// text/markdown for .md, keeping the sniffed parameters, such as its
		// charset. Extensions are registered with mime.AddExtensionType.
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

// AddTextExtensionType registers the type of text with the extension, such
// as text/x-typescript for .ts, which refines content sniffed as plain text,
// as mime.AddExtensionType's does, but is never the type of content by its
// extension alone, so content which isn't text, such as an MPEG transport
// stream, has the type it otherwise has. The extension starts with a dot,
// and is matched without case.
func AddTextExtensionType(ext, contentType string) error {
	if !strings.HasPrefix(ext, ".") || len(ext) < 2 {
		return gomedia.ErrBadParameter.Withf("invalid extension %q", ext)
	} else if mediatype, _, err := mime.ParseMediaType(contentType); err != nil {
		return gomedia.ErrBadParameter.Withf("invalid content type %q: %v", contentType, err)
	} else {
		contentType = mediatype
	}
	textExtensionTypesMu.Lock()
	defer textExtensionTypesMu.Unlock()
	textExtensionTypes[strings.ToLower(ext)] = contentType
	return nil
}

// ExtensionByType returns the preferred extension for contentType, falling
// back to the first registered extension if no preferred extension matches,
// or else the first extension of the type of text (see AddTextExtensionType).
func ExtensionByType(contentType string) string {
	for ext, rule := range extensionContentTypes {
		if rule.ContentType == contentType && rule.Preferred {
			return ext
		}
	}

	exts, err := mime.ExtensionsByType(contentType)
	if err != nil || len(exts) == 0 {
		return textExtension(contentType)
	}

	return exts[0]
}

// refines returns true when the content type of an extension is a more
// specific type of the sniffed content type. Plain text is refined by a
// text type, such as text/markdown, or an application type which is text,
// such as application/json, but not a binary type, such as an image or a
// zip archive, as sniffing would have detected it. A zip archive is refined
// by a document which is a zip archive, such as a Word document, but not a
// flat XML document. Plain text is also refined by a text type which isn't a
// text or application type, such as a playlist. Text which is plain, XML or HTML is refined by an XML
// type, such as image/svg+xml, as XML is sniffed only by how it starts: an
// SVG image sniffs as XML with a declaration, as HTML with a comment, and as
// plain text otherwise. An icon is refined by another image type, as it's
// sniffed by only four bytes, which an uncompressed TGA image starts with too.
func refines(sniffed, ext string) bool {
	switch sniffed {
	case "text/plain", "text/xml", "text/html":
		if isXMLType(ext) {
			return true
		}
	}
	switch sniffed {
	case "text/plain":
		switch {
		case ext == sniffed:
			return false
		case strings.HasPrefix(ext, "text/"), slices.Contains(textTypes, ext):
			return true
		case strings.HasPrefix(ext, "application/"):
			return ext != types.ContentTypeBinary && ext != "application/zip" && !isZipType(ext)
		}
	case "application/zip":
		return isZipType(ext)
	case "image/x-icon":
		return strings.HasPrefix(ext, "image/") && !slices.Contains(iconTypes, ext)
	}
	return false
}

// textExtensionType returns the type of text registered with the extension,
// or an empty string
func textExtensionType(ext string) string {
	textExtensionTypesMu.RLock()
	defer textExtensionTypesMu.RUnlock()
	return textExtensionTypes[ext]
}

// textExtension returns the first extension, in order, registered with the
// type of text, without its parameters or case, or an empty string
func textExtension(contentType string) string {
	contentType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	textExtensionTypesMu.RLock()
	defer textExtensionTypesMu.RUnlock()
	var exts []string
	for ext, textType := range textExtensionTypes {
		if textType == contentType {
			exts = append(exts, ext)
		}
	}
	slices.Sort(exts)
	if len(exts) == 0 {
		return ""
	}
	return exts[0]
}

// isXMLType returns true when the content type is of an XML document with a
// more specific type, such as image/svg+xml
func isXMLType(contentType string) bool {
	return strings.HasSuffix(contentType, "+xml")
}

// isZipType returns true when the content type is of a document which is a
// zip archive, but not a flat XML document
func isZipType(contentType string) bool {
	for _, suffix := range zipTypeExcluding {
		if strings.HasSuffix(contentType, suffix) {
			return false
		}
	}
	for _, prefix := range zipTypePrefixes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}
	for _, suffix := range zipTypeSuffixes {
		if strings.HasSuffix(contentType, suffix) {
			return true
		}
	}
	return false
}
