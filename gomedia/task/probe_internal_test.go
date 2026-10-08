package task

import (
	"net/url"
	"testing"
)

func TestFileProtocolPath(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"posix, RFC 8089 triple-slash", "file:///tmp/sample.mp3", "/tmp/sample.mp3"},
		{"posix, single-slash opaque-looking form", "file:/tmp/sample.mp3", "/tmp/sample.mp3"},
		{"windows, RFC 8089 triple-slash", "file:///D:/a/go-media/sample.mp3", "D:/a/go-media/sample.mp3"},
		{"windows, opaque form (no //)", "file:D:/a/go-media/sample.mp3", "D:/a/go-media/sample.mp3"},
		{"ordinary posix path isn't mistaken for a drive letter", "file:///foo/bar", "/foo/bar"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.url)
			if err != nil {
				t.Fatalf("url.Parse(%q): %v", tc.url, err)
			}
			if got := fileProtocolPath(u); got != tc.want {
				t.Fatalf("fileProtocolPath(%q) = %q, want %q", tc.url, got, tc.want)
			}
		})
	}
}
