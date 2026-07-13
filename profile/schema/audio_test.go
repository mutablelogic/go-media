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

		t.Logf("codec=%s profile=%s options=%v", candidate, profile.String(), profile.Options())
	}

}

func TestAudioProfileOptionsValidate(t *testing.T) {
	profile, err := schema.NewAudioProfile("aac")
	if err != nil {
		t.Fatalf("NewAudioProfile(%q): %v", "aac", err)
	}

	var coder *schema.Option
	var ms *schema.Option
	options := profile.Options()
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

	coder.Value = "fast"
	if err := coder.Validate(); err != nil {
		t.Fatalf("Validate(fast): %v", err)
	}

	coder.Value = "invalid"
	if err := coder.Validate(); err == nil {
		t.Fatal("expected enum validation error for invalid aac_coder value")
	}

	ms.Value = true
	if err := ms.Validate(); err != nil {
		t.Fatalf("Validate(true): %v", err)
	}

	ms.Value = "true"
	if err := ms.Validate(); err == nil {
		t.Fatal("expected type validation error for string boolean value")
	}
}
