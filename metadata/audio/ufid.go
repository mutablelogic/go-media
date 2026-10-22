package audio

import (
	"bytes"
	"encoding/binary"
	"io"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// musicbrainzOwner is the owner of the UFID frame which has a
	// recording's MusicBrainz identifier, as MusicBrainz Picard writes it
	musicbrainzOwner = "http://musicbrainz.org"

	// ufidMaxSize is the largest UFID frame read, whose owner and
	// identifier, which is at most 64 bytes, are much smaller
	ufidMaxSize = 1024
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// ufidRecordingID returns the recording's MusicBrainz identifier, from the
// UFID frame of the ID3v2 tag at the start of the stream, which ffmpeg
// doesn't return as a tag, or an empty string when there's none. Only the
// headers of the tag's frames are read, and the UFID frame, so pictures
// aren't. Versions 2.2, 2.3 and 2.4 are read; a tag whose frames are
// unsynchronised as a whole, which is rare, isn't.
func ufidRecordingID(r io.ReaderAt) string {
	// The tag's header: ID3, its version, flags and size
	var header [10]byte
	if _, err := r.ReadAt(header[:], 0); err != nil || !bytes.Equal(header[:3], []byte("ID3")) {
		return ""
	}
	version, flags := header[3], header[5]
	if version < 2 || version > 4 || (version < 4 && flags&0x80 != 0) {
		return ""
	}
	end := 10 + int64(syncsafe(header[6:10]))

	// Skip an extended header, whose size is its own in version 2.4, and
	// excludes its size in version 2.3
	offset := int64(10)
	if version > 2 && flags&0x40 != 0 {
		var size [4]byte
		if _, err := r.ReadAt(size[:], offset); err != nil {
			return ""
		}
		if version == 4 {
			offset += int64(syncsafe(size[:]))
		} else {
			offset += 4 + int64(binary.BigEndian.Uint32(size[:]))
		}
	}

	// The frames, until the end of the tag, or its padding
	idSize, headerSize := 4, 10
	if version == 2 {
		idSize, headerSize = 3, 6
	}
	frame := make([]byte, headerSize)
	for offset+int64(headerSize) <= end {
		if _, err := r.ReadAt(frame, offset); err != nil || frame[0] == 0 {
			return ""
		}
		id := string(frame[:idSize])
		var size int64
		var frameFlags uint16
		switch version {
		case 2:
			size = int64(frame[3])<<16 | int64(frame[4])<<8 | int64(frame[5])
		case 3:
			size = int64(binary.BigEndian.Uint32(frame[4:8]))
			frameFlags = binary.BigEndian.Uint16(frame[8:10])
		case 4:
			size = int64(syncsafe(frame[4:8]))
			frameFlags = binary.BigEndian.Uint16(frame[8:10])
		}
		body := offset + int64(headerSize)
		offset = body + size
		if offset > end || size <= 0 {
			return ""
		} else if id != "UFID" && id != "UFI" || size > ufidMaxSize {
			continue
		}

		// A frame which is compressed or encrypted isn't read, and a group,
		// or the data length, before the frame's data is skipped
		switch {
		case version == 3 && frameFlags&0x00c0 != 0, version == 4 && frameFlags&0x000c != 0:
			continue
		case version == 3 && frameFlags&0x0020 != 0:
			body++
		case version == 4:
			if frameFlags&0x0040 != 0 {
				body++
			}
			if frameFlags&0x0001 != 0 {
				body += 4
			}
		}
		if body >= offset {
			continue
		}
		data := make([]byte, offset-body)
		if _, err := r.ReadAt(data, body); err != nil {
			return ""
		}

		// The frame's owner, and its identifier
		owner, identifier, found := bytes.Cut(data, []byte{0})
		if found && string(owner) == musicbrainzOwner && len(identifier) > 0 {
			return string(bytes.TrimRight(identifier, "\x00"))
		}
	}
	return ""
}

// syncsafe returns the value of a syncsafe integer, whose bytes each have
// seven bits
func syncsafe(data []byte) uint32 {
	var value uint32
	for _, b := range data {
		value = value<<7 | uint32(b&0x7f)
	}
	return value
}
