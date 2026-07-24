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

	var opaque uintptr
	codec := ff.AVCodec_iterate(&opaque)
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
		codec = ff.AVCodec_iterate(&opaque)
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
	var pkt ff.AVPacket

	// This will fail, but we're just testing the API exists
	sub, err := ff.AVCodec_decode_subtitle(&ctx, &pkt)
	if err == nil && sub == nil {
		t.Log("No subtitle decoded (expected with invalid context)")
	} else if err != nil {
		t.Logf("Expected error: %v", err)
	} else {
		t.Logf("Decoded subtitle: %v", sub)
	}
}

// SetText/SetASS build rects using FFmpeg's own allocator (av_mallocz,
// av_strdup) so AVSubtitle_free can release them the same way it releases a
// decoded subtitle's rects - this exercises that round trip.
func Test_subtitle_settext_001(t *testing.T) {
	sub := ff.NewSubtitle(12345)
	defer ff.AVSubtitle_free(sub)

	if err := sub.SetText("hello world", 100, 2000); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if sub.PTS() != 12345 {
		t.Fatalf("PTS() = %d, want 12345", sub.PTS())
	}
	if sub.NumRects() != 1 {
		t.Fatalf("NumRects() = %d, want 1", sub.NumRects())
	}

	rects := sub.Rects()
	if len(rects) != 1 {
		t.Fatalf("Rects() len = %d, want 1", len(rects))
	}
	if rects[0].Type() != ff.SUBTITLE_TEXT {
		t.Fatalf("Rects()[0].Type() = %v, want SUBTITLE_TEXT", rects[0].Type())
	}
	if rects[0].Text() != "hello world" {
		t.Fatalf("Rects()[0].Text() = %q, want %q", rects[0].Text(), "hello world")
	}
	if sub.StartDisplayTime() != 100 || sub.EndDisplayTime() != 2000 {
		t.Fatalf("display times = [%d,%d], want [100,2000]", sub.StartDisplayTime(), sub.EndDisplayTime())
	}
}

// SetText/SetASS replace any prior rects, rather than leaking or appending
// to them.
func Test_subtitle_settext_replaces_prior_rect(t *testing.T) {
	sub := ff.NewSubtitle(0)
	defer ff.AVSubtitle_free(sub)

	if err := sub.SetText("first", 0, 1000); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	if err := sub.SetText("second", 0, 1000); err != nil {
		t.Fatalf("SetText (replace): %v", err)
	}
	if sub.NumRects() != 1 {
		t.Fatalf("NumRects() = %d, want 1", sub.NumRects())
	}
	if got := sub.Rects()[0].Text(); got != "second" {
		t.Fatalf("Rects()[0].Text() = %q, want %q", got, "second")
	}
}

// End-to-end: encode a SetASS-built subtitle with the real "ass" encoder.
// (Of the sampled text subtitle encoders - ass, srt, mov_text, webvtt - only
// "ass" opens cleanly against this build of libavcodec; the others fail
// AVCodec_open with AVERROR_INVALIDDATA for reasons unrelated to these
// bindings, so "ass" is the one used here to prove the construction/encode
// path genuinely works end to end.)
func Test_subtitle_encode_ass_001(t *testing.T) {
	codec := ff.AVCodec_find_encoder_by_name("ass")
	if codec == nil {
		t.Skip("ass encoder not available")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	ctx.SetTimeBase(ff.AVUtil_rational(1, 100))
	if err := ff.AVCodec_open(ctx, codec, nil); err != nil {
		t.Fatalf("AVCodec_open: %v", err)
	}

	sub := ff.NewSubtitle(0)
	defer ff.AVSubtitle_free(sub)
	if err := sub.SetASS("0,Default,,0,0,0,,Hello world", 0, 1000); err != nil {
		t.Fatalf("SetASS: %v", err)
	}

	buf := make([]byte, 65536)
	n, err := ff.AVCodec_encode_subtitle(ctx, buf, sub)
	if err != nil {
		t.Fatalf("AVCodec_encode_subtitle: %v", err)
	}
	if n == 0 {
		t.Fatal("expected some encoded bytes")
	}
	if got := string(buf[:n]); got != "0,Default,,0,0,0,,Hello world" {
		t.Fatalf("encoded = %q, want %q", got, "0,Default,,0,0,0,,Hello world")
	}
}

func Test_subtitle_encode_api_001(t *testing.T) {
	// Test that the encode function exists and has correct signature
	// We can't actually encode without a proper context and subtitle
	// This just ensures the binding compiles and links correctly

	// Just verify the function signature exists by calling it
	// with nil context (will return error, not crash)
	buf := make([]byte, 0) // Empty buffer should return EINVAL
	ret, err := ff.AVCodec_encode_subtitle(nil, buf, nil)
	if err == nil {
		t.Error("Expected error with nil context, got success")
	}
	t.Logf("Encode returned: %d bytes, error: %v (expected error with nil context)", ret, err)
}
