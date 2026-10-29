package schema

////////////////////////////////////////////////////////////////////////////////
// TYPES

// DecodeRequest specifies parameters for decoding media.
type DecodeRequest struct {
	Request // Embed base request (Input, Reader)
}
