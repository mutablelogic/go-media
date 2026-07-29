package schema_test

import (
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func TestNewCodec_IsHardware(t *testing.T) {
	// A well-known software encoder must never report itself as hardware.
	if sw := ff.AVCodec_find_encoder_by_name("libx264"); sw != nil {
		if codec := schema.NewCodec(sw); codec.IsHardware {
			t.Fatalf("NewCodec(%q): expected IsHardware=false for a software codec", sw.Name())
		}
	}

	// Every codec this ffmpeg build marks with AV_CODEC_CAP_HARDWARE must
	// come back from NewCodec with IsHardware set - this build may not
	// register any (e.g. a CI build without vaapi/nvenc/videotoolbox), in
	// which case there's nothing to check.
	var opaque uintptr
	found := false
	for {
		codec := ff.AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}
		if !codec.Capabilities().Is(ff.AV_CODEC_CAP_HARDWARE) {
			continue
		}
		found = true
		if c := schema.NewCodec(codec); !c.IsHardware {
			t.Fatalf("NewCodec(%q): expected IsHardware=true for a codec with AV_CODEC_CAP_HARDWARE", codec.Name())
		}
	}
	if !found {
		t.Skip("no hardware-accelerated codec registered in this ffmpeg build")
	}
}
