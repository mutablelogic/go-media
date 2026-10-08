package media

import (
	"net/url"
	"strings"
)

// FileProtocolPath returns the raw filesystem path to hand to FFmpeg's file
// protocol, from a parsed "file" scheme URL. FFmpeg's file protocol only
// strips the literal "file:" prefix (libavformat/file.c: av_strstart(url,
// "file:", &url)) - it does no further URI parsing at all, so whatever
// follows the colon must already be a usable OS path.
//
// RFC 8089's standard form for a local Windows path, file:///C:/dir/file
// (an empty authority, then a path that still carries the leading "/"
// before the drive letter), parses in Go to Path="/C:/dir/file" - passed
// straight through, that leading slash makes it an invalid Windows path
// ("No such file or directory"), so it's stripped here when present.
//
// A non-empty, non-"localhost" Host is a UNC path, file://server/share/file
// (Host="server", Path="/share/file") - FFmpeg's file protocol expects this
// reassembled as "//server/share/file", so the host is preserved rather
// than silently dropped.
//
// The opaque form some callers may already send directly (file:C:/dir/file
// or file:/tmp/file, no "//") bypasses all of this: u.Opaque already holds
// exactly the right string.
func FileProtocolPath(u *url.URL) string {
	if u.Opaque != "" {
		return u.Opaque
	}

	p := u.Path
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' && isASCIILetter(p[1]) {
		p = p[1:]
	}

	if host := u.Host; host != "" && !strings.EqualFold(host, "localhost") {
		return "//" + host + p
	}

	return p
}

func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
