package chromaprint_test

import (
	"io"
	"os"
	"testing"

	// Namespace imports
	. "github.com/djthorpe/go-media"
	. "github.com/djthorpe/go-media/pkg/chromaprint"
)

const (
	SAMPLE_FILE_S16 = "../../etc/s16le_22050_1ch_audio.raw"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_Stream_001(t *testing.T) {
	stream, err := NewStream(AUDIO_FMT_U8, AudioLayoutMono, 44100)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Release()
	t.Log(stream)
}

func Test_Stream_002(t *testing.T) {
	stream, err := NewStream(AUDIO_FMT_S16, AudioLayoutMono, 44100)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Release()

	buf := make([]byte, 44100*2) // One second buffer of silence
	for i := 0; i < 10; i++ {
		if err := stream.Write(buf); err != nil {
			t.Error(err)
		} else {
			t.Log(" i=", i, " duration=", stream.Duration())
		}
	}
	if fp, err := stream.Fingerprint(); err != nil {
		t.Error(err)
	} else {
		t.Log("duration=", stream.Duration(), " fp=", fp)
	}
}

func Test_Stream_003(t *testing.T) {
	r, err := os.Open(SAMPLE_FILE_S16)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	/* Create fingerprint */
	ch := AudioLayoutMono
	stream, err := NewStream(AUDIO_FMT_S16, ch, 22050)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Release()

	/* Read into buffer until EOF or max time reached */
	buf := make([]byte, 256)
	for {
		n, err := r.Read(buf)
		if err == io.EOF || n == 0 {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if err := stream.Write(buf[:n]); err != nil {
			t.Error(err)
		}
		t.Log(" duration=", stream.Duration())
	}

	// Make fingerprint
	fp, err := stream.Fingerprint()
	if err != nil {
		t.Error(err)
	} else {
		t.Log("duration=", stream.Duration().Seconds(), " fp=", fp)
	}

	// Make client
	key := os.Getenv("CHROMAPRINT_KEY")
	if key == "" {
		t.Skip("No API key set, set using environment variable CHROMAPRINT_KEY")
	}
	client, err := NewClientWithConfig(Config{Key: key})
	if err != nil {
		t.Fatal(err)
	}

	// Lookup fingerprint
	if matches, err := client.Lookup(fp, stream.Duration(), META_ALL); err != nil {
		t.Error(err)
	} else if len(matches) == 0 {
		t.Error("No matches")
	} else {
		t.Log("Matches=", matches)
	}
}
