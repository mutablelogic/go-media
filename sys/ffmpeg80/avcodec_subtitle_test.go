package ffmpeg_test

import (
	"testing"

	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func Test_subtitle_type_001(t *testing.T) {
	tests := []struct {
		value    ff.AVSubtitleType
		expected string
	}{
		{ff.SUBTITLE_NONE, "SUBTITLE_NONE"},
		{ff.SUBTITLE_BITMAP, "SUBTITLE_BITMAP"},
		{ff.SUBTITLE_TEXT, "SUBTITLE_TEXT"},
		{ff.SUBTITLE_ASS, "SUBTITLE_ASS"},
	}

	for _, test := range tests {
		if test.value.String() != test.expected {
			t.Errorf("Expected %q, got %q", test.expected, test.value.String())
		}
	}
}

func Test_subtitle_type_json_001(t *testing.T) {
	tests := []ff.AVSubtitleType{
		ff.SUBTITLE_NONE,
		ff.SUBTITLE_BITMAP,
		ff.SUBTITLE_TEXT,
		ff.SUBTITLE_ASS,
	}

	for _, subType := range tests {
		data, err := subType.MarshalJSON()
		if err != nil {
			t.Error(err)
		}
		t.Log(subType, "=>", string(data))
	}
}

func Test_subtitle_codec_001(t *testing.T) {
	// Test finding subtitle codecs
	subtitleCodecs := []struct {
		name string
		id   ff.AVCodecID
	}{
		// Text-based subtitle codecs
		{"ass", ff.AV_CODEC_ID_NONE},      // Will be set by avcodec_find_encoder_by_name
		{"srt", ff.AV_CODEC_ID_NONE},      // SubRip
		{"subrip", ff.AV_CODEC_ID_NONE},   // SubRip (alternative name)
		{"webvtt", ff.AV_CODEC_ID_NONE},   // WebVTT
		{"mov_text", ff.AV_CODEC_ID_NONE}, // MOV text (3GPP Timed Text)

		// Bitmap-based subtitle codecs
		{"dvdsub", ff.AV_CODEC_ID_NONE},            // DVD subtitles
		{"dvbsub", ff.AV_CODEC_ID_NONE},            // DVB subtitles
		{"hdmv_pgs_subtitle", ff.AV_CODEC_ID_NONE}, // Blu-ray PGS subtitles
	}

	for _, tc := range subtitleCodecs {
		// Try to find encoder
		if encoder := ff.AVCodec_find_encoder_by_name(tc.name); encoder != nil {
			if encoder.Type() != ff.AVMEDIA_TYPE_SUBTITLE {
				t.Errorf("Codec %q is not a subtitle codec, got type %v", tc.name, encoder.Type())
			}
			t.Logf("Encoder: %s (%s) - %s", encoder.Name(), encoder.ID(), encoder.LongName())
		}

		// Try to find decoder
		if decoder := ff.AVCodec_find_decoder_by_name(tc.name); decoder != nil {
			if decoder.Type() != ff.AVMEDIA_TYPE_SUBTITLE {
				t.Errorf("Codec %q is not a subtitle codec, got type %v", tc.name, decoder.Type())
			}
			t.Logf("Decoder: %s (%s) - %s", decoder.Name(), decoder.ID(), decoder.LongName())
		}
	}
}

func Test_subtitle_codec_iterate_001(t *testing.T) {
	// Iterate through all subtitle codecs
	count := 0
	encoders := 0
	decoders := 0

	codec := ff.AVCodec_iterate(nil)
	for codec != nil {
		if codec.Type() == ff.AVMEDIA_TYPE_SUBTITLE {
			count++
			if ff.AVCodec_is_encoder(codec) {
				encoders++
			}
			if ff.AVCodec_is_decoder(codec) {
				decoders++
			}
			t.Logf("Subtitle codec: %s (%s) - %s [encoder=%v, decoder=%v]",
				codec.Name(),
				codec.ID(),
				codec.LongName(),
				ff.AVCodec_is_encoder(codec),
				ff.AVCodec_is_decoder(codec),
			)
		}
		codec = ff.AVCodec_iterate(codec)
	}

	if count == 0 {
		t.Error("No subtitle codecs found")
	}

	t.Logf("Found %d subtitle codecs (%d encoders, %d decoders)", count, encoders, decoders)
}

func Test_subtitle_rect_properties_001(t *testing.T) {
	// Create a mock subtitle rect for testing property getters/setters
	// Note: In real usage, these would be allocated by FFmpeg
	var rect ff.AVSubtitleRect

	// Test position
	rect.SetX(100)
	rect.SetY(200)
	if rect.X() != 100 || rect.Y() != 200 {
		t.Errorf("Position mismatch: got (%d, %d), want (100, 200)", rect.X(), rect.Y())
	}

	// Test dimensions
	rect.SetWidth(640)
	rect.SetHeight(480)
	if rect.Width() != 640 || rect.Height() != 480 {
		t.Errorf("Dimensions mismatch: got (%d, %d), want (640, 480)", rect.Width(), rect.Height())
	}

	// Test type
	rect.SetType(ff.SUBTITLE_TEXT)
	if rect.Type() != ff.SUBTITLE_TEXT {
		t.Errorf("Type mismatch: got %v, want %v", rect.Type(), ff.SUBTITLE_TEXT)
	}

	// Test colors
	rect.SetNumColors(256)
	if rect.NumColors() != 256 {
		t.Errorf("NumColors mismatch: got %d, want 256", rect.NumColors())
	}

	// Test flags
	rect.SetFlags(1)
	if rect.Flags() != 1 {
		t.Errorf("Flags mismatch: got %d, want 1", rect.Flags())
	}

	t.Logf("SubtitleRect: %v", &rect)
}

func Test_subtitle_properties_001(t *testing.T) {
	// Create a mock subtitle for testing property getters/setters
	var sub ff.AVSubtitle

	// Test format
	sub.SetFormat(0)
	if sub.Format() != 0 {
		t.Errorf("Format mismatch: got %d, want 0", sub.Format())
	}

	// Test display times
	sub.SetStartDisplayTime(1000)
	sub.SetEndDisplayTime(5000)
	if sub.StartDisplayTime() != 1000 {
		t.Errorf("StartDisplayTime mismatch: got %d, want 1000", sub.StartDisplayTime())
	}
	if sub.EndDisplayTime() != 5000 {
		t.Errorf("EndDisplayTime mismatch: got %d, want 5000", sub.EndDisplayTime())
	}

	// Test PTS
	sub.SetPTS(90000)
	if sub.PTS() != 90000 {
		t.Errorf("PTS mismatch: got %d, want 90000", sub.PTS())
	}

	// Test rects (should be nil/empty without allocation)
	rects := sub.Rects()
	if rects != nil {
		t.Errorf("Expected nil rects, got %v", rects)
	}
	if sub.NumRects() != 0 {
		t.Errorf("Expected 0 rects, got %d", sub.NumRects())
	}

	t.Logf("Subtitle: %v", &sub)
}

func Test_subtitle_json_001(t *testing.T) {
	var sub ff.AVSubtitle
	sub.SetFormat(0)
	sub.SetStartDisplayTime(1000)
	sub.SetEndDisplayTime(5000)
	sub.SetPTS(90000)

	data, err := sub.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	t.Log("Subtitle JSON:", string(data))
}

func Test_subtitle_rect_json_001(t *testing.T) {
	var rect ff.AVSubtitleRect
	rect.SetType(ff.SUBTITLE_TEXT)
	rect.SetX(100)
	rect.SetY(200)
	rect.SetWidth(640)
	rect.SetHeight(480)

	data, err := rect.MarshalJSON()
	if err != nil {
		t.Error(err)
	}

	t.Log("SubtitleRect JSON:", string(data))
}

func Test_subtitle_decode_api_001(t *testing.T) {
	// Test that the decode function exists and has correct signature
	// We can't actually decode without a proper context and packet
	// This just ensures the binding compiles and links correctly

	var ctx ff.AVCodecContext
	var sub ff.AVSubtitle
	var got_sub int
	var pkt ff.AVPacket

	// This will fail, but we're just testing the API exists
	err := ff.AVCodec_decode_subtitle2(&ctx, &sub, &got_sub, &pkt)
	if err == nil {
		t.Log("Unexpected success - should fail with invalid context")
	} else {
		t.Logf("Expected error: %v", err)
	}
}

func Test_subtitle_encode_api_001(t *testing.T) {
	// Test that the encode function exists and has correct signature
	// We can't actually encode without a proper context and subtitle
	// This just ensures the binding compiles and links correctly

	var ctx ff.AVCodecContext
	var sub ff.AVSubtitle
	buf := make([]byte, 1024)

	// This will fail/return 0, but we're just testing the API exists
	ret := ff.AVCodec_encode_subtitle(&ctx, buf, &sub)
	t.Logf("Encode returned: %d (expected failure/0 with invalid context)", ret)
}
