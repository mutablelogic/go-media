package schema_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func TestNewAudioProfile_Print(t *testing.T) {
	codecs := []string{"aac", "libmp3lame", "flac", "opus", "pcm_s16le"}

	for _, candidate := range codecs {
		if ff.AVCodec_find_encoder_by_name(candidate) == nil {
			continue
		}
		profile, err := schema.NewAudioProfile(candidate)
		if err != nil {
			t.Fatalf("NewAudioProfile(%q): %v", candidate, err)
		}
		if profile == nil {
			t.Fatalf("NewAudioProfile(%q): nil profile", candidate)
		}

		options := schema.OptionsForCodec(ff.AVCodec_find_encoder_by_name(candidate))
		t.Logf("codec=%s profile=%s options=%v", candidate, profile.String(), options)
	}

}

func TestAudioProfileOptionsValidate(t *testing.T) {
	_, err := schema.NewAudioProfile("aac")
	if err != nil {
		t.Fatalf("NewAudioProfile(%q): %v", "aac", err)
	}

	var coder *schema.Option
	var ms *schema.Option
	options := schema.OptionsForCodec(ff.AVCodec_find_encoder_by_name("aac"))
	for i := range options {
		switch options[i].Name {
		case "aac_coder":
			coder = &options[i]
		case "aac_ms":
			ms = &options[i]
		}
	}
	if coder == nil {
		t.Skip("aac_coder option not available")
	}
	if ms == nil {
		t.Skip("aac_ms option not available")
	}

	if _, err := coder.Validate("fast"); err != nil {
		t.Fatalf("Validate(fast): %v", err)
	}

	if _, err := coder.Validate("invalid"); err == nil {
		t.Fatal("expected enum validation error for invalid aac_coder value")
	}

	if _, err := ms.Validate(true); err != nil {
		t.Fatalf("Validate(true): %v", err)
	}

	if _, err := ms.Validate("true"); err == nil {
		t.Fatal("expected type validation error for string boolean value")
	}
}

// FFmpeg's native aac encoder declares no static AVCodec.profiles list and
// no private "profile" AVOption of its own, so "profile" isn't a real,
// discoverable option for it at all — Set() should reject it rather than
// silently accepting a value the codec can never act on.
func TestAudioProfile_Set_Profile_Unsupported(t *testing.T) {
	profile, err := schema.NewAudioProfile("aac")
	if err != nil {
		t.Fatalf("NewAudioProfile(aac): %v", err)
	}

	options := schema.OptionsForCodec(ff.AVCodec_find_encoder_by_name("aac"))
	for _, opt := range options {
		if opt.Name == schema.OptionProfile {
			t.Skip("aac encoder now declares a profile option; this codec no longer demonstrates the unsupported case")
		}
	}

	if err := profile.Set(schema.OptionProfile, "LC"); err == nil {
		t.Fatal("Set(profile): expected error for codec with no profile concept")
	}
}

// libmp3lame has neither a static profile list nor a private "profile"
// option, so the synthetic option must not be injected for it — confirming
// the fix for codecs that have no concept of "profile" at all.
func TestOptionsForCodec_NoProfileOption_Audio(t *testing.T) {
	codec := ff.AVCodec_find_encoder_by_name("libmp3lame")
	if codec == nil {
		t.Skip("libmp3lame encoder not available")
	}
	for _, opt := range schema.OptionsForCodec(codec) {
		if opt.Name == schema.OptionProfile {
			t.Fatalf("OptionsForCodec(libmp3lame): unexpected %q option for a codec with no profile concept", schema.OptionProfile)
		}
	}
}

func TestAudioProfile_Set_Bitrate(t *testing.T) {
	profile, err := schema.NewAudioProfile("aac")
	if err != nil {
		t.Fatalf("NewAudioProfile(aac): %v", err)
	}

	if err := profile.Set(schema.OptionBitrate, uint64(128000)); err != nil {
		t.Fatalf("Set(bitrate): %v", err)
	}
	if got := profile.Par().BitRate(); got != 128000 {
		t.Fatalf("Par().BitRate() = %d, want 128000", got)
	}
}
