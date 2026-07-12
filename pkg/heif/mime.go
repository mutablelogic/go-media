package heif

import "mime"

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	mime.AddExtensionType(".heic", "image/heic")
	mime.AddExtensionType(".heics", "image/heic-sequence")
	mime.AddExtensionType(".heif", "image/heif")
	mime.AddExtensionType(".heifs", "image/heif-sequence")
	mime.AddExtensionType(".avif", "image/avif")
	mime.AddExtensionType(".avis", "image/avif-sequence")
}
