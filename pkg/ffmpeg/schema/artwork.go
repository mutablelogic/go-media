package schema

// Artwork represents embedded artwork (cover art, thumbnails, etc.) as raw bytes.
// When JSON encoded, it is represented as a base64 string.
type Artwork []byte
