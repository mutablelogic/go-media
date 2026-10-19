package metadata

import "testing"

func Test_mime_refines(t *testing.T) {
	for _, tc := range []struct {
		sniffed, ext string
		want         bool
	}{
		// Plain text is refined by a more specific text type
		{"text/plain", "text/markdown", true},
		{"text/plain", "text/plain", false},
		{"text/plain", "application/json", true},
		{"text/plain", "application/vnd.google-apps.document", true},
		{"text/plain", "application/octet-stream", false},
		{"text/plain", "application/zip", false},
		{"text/plain", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false},
		{"text/plain", "image/png", false},
		{"text/plain", "audio/mpeg", false},

		// A zip archive is refined by a document which is a zip archive
		{"application/zip", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", true},
		{"application/zip", "application/vnd.oasis.opendocument.text", true},
		{"application/zip", "application/epub+zip", true},
		{"application/zip", "image/jpeg", false},

		// But not by a flat XML document, which isn't a zip archive
		{"application/zip", "application/vnd.oasis.opendocument.text-flat-xml", false},
		{"application/zip", "application/vnd.oasis.opendocument.spreadsheet-flat-xml", false},

		// Text which is plain, XML or HTML is refined by an XML type
		{"text/plain", "image/svg+xml", true},
		{"text/xml", "image/svg+xml", true},
		{"text/html", "image/svg+xml", true},
		{"text/xml", "application/atom+xml", true},
		{"text/xml", "text/markdown", false},
		{"text/html", "text/plain", false},

		// Other sniffed types aren't refined
		{"image/png", "image/jpeg", false},
		{"image/png", "image/svg+xml", false},
	} {
		if got := refines(tc.sniffed, tc.ext); got != tc.want {
			t.Errorf("refines(%q, %q) = %v, want %v", tc.sniffed, tc.ext, got, tc.want)
		}
	}
}
