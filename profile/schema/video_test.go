package schema_test

import (
	"strings"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func TestNewVideoProfile_Print(t *testing.T) {
	codecs := []string{"libx264", "libx265", "mpeg4", "rawvideo"}

	for _, candidate := range codecs {
		if ff.AVCodec_find_encoder_by_name(candidate) == nil {
			continue
		}
		profile, err := schema.NewVideoProfile(candidate)
		if err != nil {
			t.Fatalf("NewVideoProfile(%q): %v", candidate, err)
		}
		if profile == nil {
			t.Fatalf("NewVideoProfile(%q): nil profile", candidate)
		}

		options := schema.OptionsForCodec(ff.AVCodec_find_encoder_by_name(candidate))
		t.Logf("codec=%s profile=%s options=%v", candidate, profile.String(), options)
	}
}

func TestNewVideoProfile_NotFound(t *testing.T) {
	if _, err := schema.NewVideoProfile("does-not-exist"); err == nil {
		t.Fatal("NewVideoProfile: expected error for unknown codec")
	}
}

func TestNewVideoProfile_NotVideo(t *testing.T) {
	// aac is an audio codec, not a video one
	if _, err := schema.NewVideoProfile("aac"); err == nil {
		t.Fatal("NewVideoProfile: expected error for non-video codec")
	}
}

func TestVideoProfile_Set(t *testing.T) {
	profile, err := schema.NewVideoProfile("rawvideo")
	if err != nil {
		t.Skipf("rawvideo encoder not available: %v", err)
	}

	if err := profile.Set(schema.OptionWidth, uint64(640)); err != nil {
		t.Fatalf("Set(width): %v", err)
	}
	if err := profile.Set(schema.OptionHeight, uint64(480)); err != nil {
		t.Fatalf("Set(height): %v", err)
	}
	if err := profile.Set(schema.OptionFrameRate, float64(25)); err != nil {
		t.Fatalf("Set(frame_rate): %v", err)
	}

	if got := profile.Par().Width(); got != 640 {
		t.Fatalf("Par().Width() = %d, want 640", got)
	}
	if got := profile.Par().Height(); got != 480 {
		t.Fatalf("Par().Height() = %d, want 480", got)
	}

	tb := profile.TimeBase()
	if tb == nil {
		t.Fatal("TimeBase(): expected non-nil timebase after setting frame_rate")
	}
	if got := ff.AVUtil_rational_q2d(*tb); got != 0.04 {
		t.Fatalf("TimeBase() = %v (%.4f), want 1/25 (0.04)", tb, got)
	}
}

func TestVideoProfile_Set_UnsupportedOption(t *testing.T) {
	profile, err := schema.NewVideoProfile("rawvideo")
	if err != nil {
		t.Skipf("rawvideo encoder not available: %v", err)
	}

	if err := profile.Set("not_a_real_option", "x"); err == nil {
		t.Fatal("Set: expected error for unsupported option")
	}
}

// rawvideo has neither a static profile list nor a private "profile"
// option, so the synthetic option must not be injected for it — confirming
// the fix for codecs that have no concept of "profile" at all.
func TestOptionsForCodec_NoProfileOption_Video(t *testing.T) {
	codec := ff.AVCodec_find_encoder_by_name("rawvideo")
	if codec == nil {
		t.Skip("rawvideo encoder not available")
	}
	for _, opt := range schema.OptionsForCodec(codec) {
		if opt.Name == schema.OptionProfile {
			t.Fatalf("OptionsForCodec(rawvideo): unexpected %q option for a codec with no profile concept", schema.OptionProfile)
		}
	}
}

func TestVideoProfile_Set_InvalidPixelFormat(t *testing.T) {
	profile, err := schema.NewVideoProfile("rawvideo")
	if err != nil {
		t.Skipf("rawvideo encoder not available: %v", err)
	}

	if err := profile.Set(schema.OptionPixelFormat, "not_a_real_pixel_format"); err == nil {
		t.Fatal("Set: expected error for invalid pixel format")
	}
}

// prores_ks is one of the few encoders that declares a static profile list
// (codec.Profiles()), so Set(profile, ...) here goes through the dedicated
// numeric AVCodecParameters.profile field, resolved by name.
func TestVideoProfile_Set_Profile(t *testing.T) {
	profile, err := schema.NewVideoProfile("prores_ks")
	if err != nil {
		t.Skipf("prores_ks encoder not available: %v", err)
	}

	if err := profile.Set(schema.OptionProfile, "proxy"); err != nil {
		t.Fatalf("Set(profile, %q): %v", "proxy", err)
	}
	if got := profile.Par().Profile(); got == int(ff.AV_PROFILE_UNKNOWN) {
		t.Fatal("Par().Profile() left as AV_PROFILE_UNKNOWN after Set(profile, \"proxy\")")
	}
}

func TestVideoProfile_Set_Profile_Invalid(t *testing.T) {
	profile, err := schema.NewVideoProfile("prores_ks")
	if err != nil {
		t.Skipf("prores_ks encoder not available: %v", err)
	}
	if err := profile.Set(schema.OptionProfile, "not_a_real_profile"); err == nil {
		t.Fatal("Set(profile): expected error for unknown profile")
	}
}

// libx264 exposes "profile" only as its own private string option (baseline,
// main, high, ...) rather than the generic AVCodecParameters.profile field,
// and declares nothing in codec.Profiles(). Set() must defer to the codec's
// own option dict for these rather than the dedicated numeric field, or
// every profile value would be rejected as "unknown".
func TestVideoProfile_Set_Profile_PrivateOptionFallback(t *testing.T) {
	profile, err := schema.NewVideoProfile("libx264")
	if err != nil {
		t.Skipf("libx264 encoder not available: %v", err)
	}

	if err := profile.Set(schema.OptionProfile, "high"); err != nil {
		t.Fatalf("Set(profile, \"high\"): %v", err)
	}
	if got := profile.Par().Profile(); got != int(ff.AV_PROFILE_UNKNOWN) {
		t.Fatalf("Par().Profile() = %d, want AV_PROFILE_UNKNOWN (value should live in Options(), not Par())", got)
	}
	if !strings.Contains(string(profile.Options()), "high") {
		t.Fatalf("Options() = %s, want it to contain the profile value", profile.Options())
	}
}
