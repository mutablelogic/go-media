package schema_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

// Regression test: NewOutputFormat used to build Audio/Video/Subtitle by
// iterating the codec registry and keeping whatever was compatible, in
// registry order - an order with no relationship to which codec the format
// actually prefers. The doc comment claimed "the first one is the default"
// even though it wasn't; this checks it actually is, against the format's
// own AudioCodec()/VideoCodec()/SubtitleCodec() (FFmpeg's own declared
// default for that container).
func TestNewOutputFormat_DefaultCodecFirst(t *testing.T) {
	for _, name := range []string{"mp4", "matroska", "mp3", "ogg", "avi", "webm"} {
		t.Run(name, func(t *testing.T) {
			fmt_ := ff.AVFormat_guess_format(name, "", "")
			if fmt_ == nil {
				t.Skipf("format %q not available", name)
			}
			f := schema.NewOutputFormat(fmt_)
			if f == nil {
				t.Fatal("NewOutputFormat: nil")
			}

			checkFirst := func(label string, codecs []string, id ff.AVCodecID) {
				encoder := ff.AVCodec_find_encoder(id)
				if encoder == nil {
					// No default declared/resolvable for this format+kind -
					// nothing to assert.
					return
				}
				// Some formats declare a default that AVFormat_query_codec
				// itself doesn't consider compatible (e.g. mp3's default
				// video codec is "png", for embedded cover art, which
				// query_codec rejects for regular stream muxing) - in that
				// case the codec never appears in the derived list at all,
				// which is a separate, pre-existing limitation of how the
				// list itself is built, not of the ordering this test covers.
				found := false
				for _, c := range codecs {
					if c == encoder.Name() {
						found = true
						break
					}
				}
				if !found {
					t.Skipf("%s: default %q not present in derived list %v (query_codec limitation, not an ordering issue)", label, encoder.Name(), codecs)
				}
				if codecs[0] != encoder.Name() {
					t.Errorf("%s[0] = %q, want default %q", label, codecs[0], encoder.Name())
				}
			}

			checkFirst("Audio", f.Audio, fmt_.AudioCodec())
			checkFirst("Video", f.Video, fmt_.VideoCodec())
			checkFirst("Subtitle", f.Subtitle, fmt_.SubtitleCodec())
		})
	}
}
