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
	// TextType is the type to use when the content sniffs as plain text,
	// which every text file does. For a text extension, such as .go, it is
	// the extension's own type. For .ts, which is an MPEG transport stream
	// unless it is text, it is TypeScript.
	TextType string
}

// text returns the rule for a text extension, whose type is used when the
// content sniffs as plain text
func text(contentType string) extensionRule {
	return extensionRule{ContentType: contentType, TextType: contentType}
}

// extensionContentTypes stores extension -> media type mappings and whether
// this extension should be preferred when a type has multiple aliases.
// It is used both for content type detection and extension selection, and
// the types are registered with the mime package, so that every platform
// returns the same types for them. Camera raw and HEIF types are registered
// by pkg/raw and pkg/heif.
var extensionContentTypes = map[string]extensionRule{
	// Video. Matroska sniffs as video/webm, the MP4 family as video/mp4 and
	// Ogg as application/ogg, so their extensions are authoritative.
	".3gp":  {ContentType: "video/3gpp", Authoritative: true},
	".avi":  {ContentType: "video/x-msvideo"},
	".flv":  {ContentType: "video/x-flv"},
	".m2ts": {ContentType: "video/mp2t"},
	".m4v":  {ContentType: "video/x-m4v", Authoritative: true},
	".mk3d": {ContentType: "video/x-matroska-3d", Authoritative: true},
	".mkv":  {ContentType: "video/x-matroska", Preferred: true, Authoritative: true},
	".mov":  {ContentType: "video/quicktime", Authoritative: true},
	".mpeg": {ContentType: "video/mpeg"},
	".mpg":  {ContentType: "video/mpeg", Preferred: true},
	".ogv":  {ContentType: "video/ogg", Authoritative: true},
	".ts":   {ContentType: "video/mp2t", Preferred: true, TextType: "text/typescript"},
	".vob":  {ContentType: "video/mpeg"},
	".webm": {ContentType: "video/webm"},
	".wmv":  {ContentType: "video/x-ms-wmv"},

	// Audio
	".aac":  {ContentType: "audio/aac"},
	".aif":  {ContentType: "audio/aiff"},
	".aiff": {ContentType: "audio/aiff", Preferred: true},
	".ape":  {ContentType: "audio/x-ape"},
	".dsf":  {ContentType: "audio/x-dsf"},
	".m4a":  {ContentType: "audio/mp4", Preferred: true, Authoritative: true},
	".m4b":  {ContentType: "audio/mp4", Authoritative: true},
	".mid":  {ContentType: "audio/midi", Preferred: true},
	".midi": {ContentType: "audio/midi"},
	".mka":  {ContentType: "audio/x-matroska", Authoritative: true},
	".oga":  {ContentType: "audio/ogg", Authoritative: true},
	".ogg":  {ContentType: "audio/ogg", Preferred: true, Authoritative: true},
	".opus": {ContentType: "audio/opus", Authoritative: true},
	".wav":  {ContentType: "audio/wav"},
	".wma":  {ContentType: "audio/x-ms-wma"},

	// Playlists and cue sheets. The .m3u and .pls types are audio types, so
	// aren't used for their text content, which would otherwise be read by
	// the audio handlers.
	".cue":  text("application/x-cue"),
	".m3u":  {ContentType: "audio/x-mpegurl"},
	".m3u8": text("application/vnd.apple.mpegurl"),
	".pls":  {ContentType: "audio/x-scpls"},

	// Subtitles
	".ass": text("text/x-ssa"),
	".srt": text("application/x-subrip"),
	".ssa": text("text/x-ssa"),
	".vtt": text("text/vtt"),

	// Images
	".jpg":  {ContentType: "image/jpeg", Preferred: true},
	".jpeg": {ContentType: "image/jpeg"},
	".jxl":  {ContentType: "image/jxl"},
	".tif":  {ContentType: "image/tiff"},
	".tiff": {ContentType: "image/tiff", Preferred: true},

	// Documents and metadata. EPUB sniffs as application/zip.
	".csv":  text("text/csv"),
	".epub": {ContentType: "application/epub+zip", Authoritative: true},
	".md":   text("text/markdown"),
	".nfo":  text("text/x-nfo"),
	".txt":  text("text/plain"),

	// Archives
	".7z":  {ContentType: "application/x-7z-compressed"},
	".bz2": {ContentType: "application/x-bzip2"},
	".gz":  {ContentType: "application/gzip", Preferred: true},
	".rar": {ContentType: "application/vnd.rar"},
	".tar": {ContentType: "application/x-tar"},
	".tgz": {ContentType: "application/gzip"},
	".xz":  {ContentType: "application/x-xz"},
	".zip": {ContentType: "application/zip"},
	".zst": {ContentType: "application/zstd"},

	// Source code and configuration
	".c":     text("text/x-c"),
	".cpp":   text("text/x-c++"),
	".go":    text("text/x-go"),
	".h":     text("text/x-c"),
	".hpp":   text("text/x-c++"),
	".java":  text("text/x-java"),
	".jsx":   text("text/jsx"),
	".proto": text("text/x-protobuf"),
	".py":    text("text/x-python"),
	".rb":    text("text/x-ruby"),
	".rs":    text("text/x-rust"),
	".sh":    text("text/x-shellscript"),
	".sql":   text("application/sql"),
	".swift": text("text/x-swift"),
	".toml":  text("application/toml"),
	".tsx":   text("text/tsx"),
	".yaml":  text("application/yaml"),
	".yml":   text("application/yaml"),
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	for ext, rule := range extensionContentTypes {
		if err := mime.AddExtensionType(ext, rule.ContentType); err != nil {
			panic(err)
		}
	}
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

	var extType, textType string
	if named, ok := r.(NamedStream); ok {
		ext := strings.ToLower(filepath.Ext(named.Name()))
		if forced, ok := extensionContentTypes[ext]; ok {
			if forced.Authoritative {
				return mime.ParseMediaType(forced.ContentType)
			}
			// Not Authoritative (e.g. .jpg/.jpeg): use this as the
			// extension-based fallback below, but still sniff first -
			// unlike .m4a, there's no reason not to. Extensions not in this
			// map at all go through mime.TypeByExtension the same way.
			extType, textType = forced.ContentType, forced.TextType
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
	// An empty stream sniffs as plain text, so isn't sniffed
	if mediaType := http.DetectContentType(buf[:n]); n > 0 && mediaType != types.ContentTypeBinary {
		// Plain text is more specific when the extension says what kind of
		// text it is, such as .go source or a .ts TypeScript file. Keep the
		// parameters from sniffing, such as the charset.
		contentType, params, err := mime.ParseMediaType(mediaType)
		if err == nil && contentType == "text/plain" && textType != "" {
			return textType, params, nil
		}
		return contentType, params, err
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
