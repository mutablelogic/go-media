package schema_test

import (
	"strings"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func TestNewSubtitleProfile_Print(t *testing.T) {
	codecs := []string{"srt", "ass", "webvtt", "mov_text"}

	for _, candidate := range codecs {
		if ff.AVCodec_find_encoder_by_name(candidate) == nil {
			continue
		}
		profile, err := schema.NewSubtitleProfile(candidate)
		if err != nil {
			t.Fatalf("NewSubtitleProfile(%q): %v", candidate, err)
		}
		if profile == nil {
			t.Fatalf("NewSubtitleProfile(%q): nil profile", candidate)
		}

		options := schema.OptionsForCodec(ff.AVCodec_find_encoder_by_name(candidate))
		t.Logf("codec=%s profile=%s options=%v", candidate, profile.String(), options)
	}
}

func TestNewSubtitleProfile_NotFound(t *testing.T) {
	if _, err := schema.NewSubtitleProfile("does-not-exist"); err == nil {
		t.Fatal("NewSubtitleProfile: expected error for unknown codec")
	}
}

func TestNewSubtitleProfile_NotSubtitle(t *testing.T) {
	// aac is an audio codec, not a subtitle one
	if _, err := schema.NewSubtitleProfile("aac"); err == nil {
		t.Fatal("NewSubtitleProfile: expected error for non-subtitle codec")
	}
}

func TestSubtitleProfile_TimeBase(t *testing.T) {
	profile, err := schema.NewSubtitleProfile("srt")
	if err != nil {
		t.Skipf("srt encoder not available: %v", err)
	}
	if tb := profile.TimeBase(); tb != nil {
		t.Fatalf("TimeBase() = %v, want nil (subtitles have no rate field to derive one from)", tb)
	}
}

// mov_text declares its own private "height" option (frame height, usually
// video height) — a codec-specific knob, not a universal field, so it must
// flow through Options() like any other private option.
func TestSubtitleProfile_Set_CodecSpecificOption(t *testing.T) {
	profile, err := schema.NewSubtitleProfile("mov_text")
	if err != nil {
		t.Skipf("mov_text encoder not available: %v", err)
	}

	if err := profile.Set("height", uint64(1080)); err != nil {
		t.Fatalf("Set(height): %v", err)
	}
	if !strings.Contains(string(profile.Options()), "1080") {
		t.Fatalf("Options() = %s, want it to contain the height value", profile.Options())
	}
}

func TestSubtitleProfile_Set_UnsupportedOption(t *testing.T) {
	profile, err := schema.NewSubtitleProfile("srt")
	if err != nil {
		t.Skipf("srt encoder not available: %v", err)
	}

	if err := profile.Set("not_a_real_option", "x"); err == nil {
		t.Fatal("Set: expected error for unsupported option")
	}
}

// No sampled subtitle codec declares a static profile list, and subtitles
// have no other universal option — confirm the synthetic option machinery
// used for audio/video doesn't leak a "profile" (or any other phantom)
// option into subtitle codecs that don't support it.
func TestOptionsForCodec_NoProfileOption_Subtitle(t *testing.T) {
	codec := ff.AVCodec_find_encoder_by_name("srt")
	if codec == nil {
		t.Skip("srt encoder not available")
	}
	for _, opt := range schema.OptionsForCodec(codec) {
		if opt.Name == schema.OptionProfile {
			t.Fatalf("OptionsForCodec(srt): unexpected %q option for a codec with no profile concept", schema.OptionProfile)
		}
	}
}
