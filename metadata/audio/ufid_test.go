package audio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

const recordingID = "9fee7f18-fbb1-43f5-821a-4b88d2fe3cf5"

// syncsafeBytes returns the value as a syncsafe integer
func syncsafeBytes(value int) []byte {
	return []byte{byte(value >> 21 & 0x7f), byte(value >> 14 & 0x7f), byte(value >> 7 & 0x7f), byte(value & 0x7f)}
}

// id3Frame returns a frame of the version, with the flags, and the extra
// bytes before its data, such as a group
func id3Frame(version byte, id string, flags uint16, extra, data []byte) []byte {
	data = append(append([]byte{}, extra...), data...)
	var frame bytes.Buffer
	frame.WriteString(id)
	switch version {
	case 2:
		frame.Write([]byte{byte(len(data) >> 16), byte(len(data) >> 8), byte(len(data))})
	case 3:
		frame.Write(binary.BigEndian.AppendUint32(nil, uint32(len(data))))
		frame.Write(binary.BigEndian.AppendUint16(nil, flags))
	case 4:
		frame.Write(syncsafeBytes(len(data)))
		frame.Write(binary.BigEndian.AppendUint16(nil, flags))
	}
	frame.Write(data)
	return frame.Bytes()
}

// id3Tag returns a tag of the version, with the flags, and the frames,
// followed by padding and audio
func id3Tag(version, flags byte, frames ...[]byte) []byte {
	body := append(bytes.Join(frames, nil), make([]byte, 32)...)
	tag := append([]byte{'I', 'D', '3', version, 0, flags}, syncsafeBytes(len(body))...)
	return append(append(tag, body...), 0xff, 0xfb, 0x90, 0x00)
}

// ufid returns a UFID frame's data, with the owner and identifier
func ufid(owner, identifier string) []byte {
	return append(append([]byte(owner), 0), identifier...)
}

func Test_ufid(t *testing.T) {
	picture := bytes.Repeat([]byte{0xff, 0xd8}, 2000)
	tests := []struct {
		name string
		tag  []byte
		want string
	}{
		{"v2.3", id3Tag(3, 0,
			id3Frame(3, "TIT2", 0, nil, []byte("\x00Departure")),
			id3Frame(3, "APIC", 0, nil, picture),
			id3Frame(3, "UFID", 0, nil, ufid(musicbrainzOwner, recordingID)),
		), recordingID},
		{"v2.4", id3Tag(4, 0, id3Frame(4, "UFID", 0, nil, ufid(musicbrainzOwner, recordingID))), recordingID},
		{"v2.4 data length", id3Tag(4, 0, id3Frame(4, "UFID", 0x0001, syncsafeBytes(len(ufid(musicbrainzOwner, recordingID))), ufid(musicbrainzOwner, recordingID))), recordingID},
		{"v2.3 group", id3Tag(3, 0, id3Frame(3, "UFID", 0x0020, []byte{1}, ufid(musicbrainzOwner, recordingID))), recordingID},
		{"v2.2", id3Tag(2, 0, id3Frame(2, "UFI", 0, nil, ufid(musicbrainzOwner, recordingID))), recordingID},
		{"another owner first", id3Tag(3, 0,
			id3Frame(3, "UFID", 0, nil, ufid("http://www.cddb.com/id3/taginfo1.html", "3CD3N")),
			id3Frame(3, "UFID", 0, nil, ufid(musicbrainzOwner, recordingID)),
		), recordingID},
		{"v2.3 extended header", id3Tag(3, 0x40, append([]byte{0, 0, 0, 6, 0, 0, 0, 0, 0, 0}, id3Frame(3, "UFID", 0, nil, ufid(musicbrainzOwner, recordingID))...)), recordingID},
		{"v2.4 extended header", id3Tag(4, 0x40, append(append(syncsafeBytes(6), 1, 0), id3Frame(4, "UFID", 0, nil, ufid(musicbrainzOwner, recordingID))...)), recordingID},
		{"only another owner", id3Tag(3, 0, id3Frame(3, "UFID", 0, nil, ufid("http://www.cddb.com/id3/taginfo1.html", "3CD3N"))), ""},
		{"compressed", id3Tag(3, 0, id3Frame(3, "UFID", 0x0080, nil, ufid(musicbrainzOwner, recordingID))), ""},
		{"unsynchronised", id3Tag(3, 0x80, id3Frame(3, "UFID", 0, nil, ufid(musicbrainzOwner, recordingID))), ""},
		{"no frames", id3Tag(3, 0), ""},
		{"no tag", []byte{0xff, 0xfb, 0x90, 0x00, 0, 0, 0, 0, 0, 0, 0, 0}, ""},
		{"truncated", id3Tag(3, 0, id3Frame(3, "UFID", 0, nil, ufid(musicbrainzOwner, recordingID)))[:20], ""},
		{"empty", nil, ""},
	}
	for _, test := range tests {
		if got := ufidRecordingID(bytes.NewReader(test.tag)); got != test.want {
			t.Errorf("%s: got %q, want %q", test.name, got, test.want)
		}
	}
}
