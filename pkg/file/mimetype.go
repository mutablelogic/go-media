package file

import (
	"mime"
	"net/http"
)

// MimeType returns the mimetype of the data, and returns the mimetype, file extension
func MimeType(data []byte) (string, string, error) {
	mimetype := http.DetectContentType(data)
	exts, err := mime.ExtensionsByType(mimetype)
	if err != nil || len(exts) == 0 {
		return mimetype, "", err
	} else {
		return mimetype, exts[0], nil
	}
}
