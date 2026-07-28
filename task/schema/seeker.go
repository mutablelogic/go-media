package schema

import (
	"bytes"
	"encoding/json"
	"io"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ReadSeeker struct {
	io.ReadSeeker
	name string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new in-memory ReadSeeker from a Reader. If the reader is already a
// ReadSeeker, it is returned as-is. Otherwise, the reader is read into memory
// and a new ReadSeeker is created.
func NewReadSeeker(r io.Reader) (*ReadSeeker, error) {
	var self ReadSeeker

	// Check arguments and set name
	if r == nil {
		return nil, gomedia.ErrBadParameter.With("nil reader")
	} else if named, ok := r.(gomedia.NamedReader); ok && named != nil {
		self.name = named.Name()
	}

	// Make the input replayable so we can read once for content type
	// detection and a second time for metadata extraction.
	if seeker, ok := r.(io.ReadSeeker); ok {
		self.ReadSeeker = seeker
	} else {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			return nil, err
		} else {
			self.ReadSeeker = bytes.NewReader(buf.Bytes())
		}
	}

	// Seek to start
	if _, err := self.ReadSeeker.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	// Return success
	return &self, nil
}

////////////////////////////////////////////////////////////////////////////////
// MARSHALING

// MarshalJSON renders the input's name - the embedded io.ReadSeeker has no
// exported fields of its own (and couldn't be usefully serialized as a
// byte stream anyway), so the default encoding would otherwise produce {}.
func (r ReadSeeker) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name string `json:"name,omitempty" help:"Name of the input." example:"sample.mp3"`
	}{
		Name: r.name,
	})
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r ReadSeeker) Name() string {
	return r.name
}
