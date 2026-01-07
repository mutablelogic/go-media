package ffmpeg_test

import (
	"encoding/json"
	"testing"
	"unsafe"

	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set and av_opt_get with AVCodecContext

func Test_avutil_opt_001(t *testing.T) {
	// Get H264 encoder for testing
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	// Allocate codec context
	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting string option
	if err := ff.AVUtil_opt_set(unsafe.Pointer(ctx), "preset", "fast", 0); err != nil {
		t.Logf("Warning: could not set preset option: %v", err)
	}

	// Test getting string option (if it was set)
	value, err := ff.AVUtil_opt_get(unsafe.Pointer(ctx), "preset", 0)
	if err == nil {
		t.Logf("preset = %s", value)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_int and av_opt_get_int

func Test_avutil_opt_002(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting integer option (bitrate)
	const testBitrate = 400000
	if err := ff.AVUtil_opt_set_int(unsafe.Pointer(ctx), "b", testBitrate, 0); err != nil {
		t.Logf("Warning: could not set bitrate option: %v", err)
	}

	// Test getting integer option
	bitrate, err := ff.AVUtil_opt_get_int(unsafe.Pointer(ctx), "b", 0)
	if err == nil {
		t.Logf("bitrate = %d", bitrate)
		if bitrate != testBitrate {
			t.Logf("Note: bitrate value differs: expected %d, got %d", testBitrate, bitrate)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_double and av_opt_get_double

func Test_avutil_opt_003(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting double option (qcompress)
	const testQCompress = 0.6
	if err := ff.AVUtil_opt_set_double(unsafe.Pointer(ctx), "qcompress", testQCompress, 0); err != nil {
		t.Logf("Warning: could not set qcompress option: %v", err)
	}

	// Test getting double option
	qcompress, err := ff.AVUtil_opt_get_double(unsafe.Pointer(ctx), "qcompress", 0)
	if err == nil {
		t.Logf("qcompress = %f", qcompress)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_q and av_opt_get_q

func Test_avutil_opt_004(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting rational option (time_base)
	testTimeBase := ff.AVUtil_rational(1, 25)
	if err := ff.AVUtil_opt_set_q(unsafe.Pointer(ctx), "time_base", testTimeBase, 0); err != nil {
		t.Logf("Warning: could not set time_base option: %v", err)
	}

	// Test getting rational option
	timeBase, err := ff.AVUtil_opt_get_q(unsafe.Pointer(ctx), "time_base", 0)
	if err == nil {
		t.Logf("time_base = %d/%d", timeBase.Num(), timeBase.Den())
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_image_size and av_opt_get_image_size

func Test_avutil_opt_005(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting image size option (video_size)
	const testWidth, testHeight = 1920, 1080
	if err := ff.AVUtil_opt_set_image_size(unsafe.Pointer(ctx), "video_size", testWidth, testHeight, 0); err != nil {
		t.Logf("Warning: could not set video_size option: %v", err)
	}

	// Test getting image size option
	width, height, err := ff.AVUtil_opt_get_image_size(unsafe.Pointer(ctx), "video_size", 0)
	if err == nil {
		t.Logf("video_size = %dx%d", width, height)
		if width != testWidth || height != testHeight {
			t.Logf("Note: video_size value differs: expected %dx%d, got %dx%d", testWidth, testHeight, width, height)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_pixel_fmt and av_opt_get_pixel_fmt

func Test_avutil_opt_006(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting pixel format option
	const testPixFmt = ff.AV_PIX_FMT_YUV420P
	if err := ff.AVUtil_opt_set_pixel_fmt(unsafe.Pointer(ctx), "pixel_format", testPixFmt, 0); err != nil {
		t.Logf("Warning: could not set pixel_format option: %v", err)
	}

	// Test getting pixel format option
	pixFmt, err := ff.AVUtil_opt_get_pixel_fmt(unsafe.Pointer(ctx), "pixel_format", 0)
	if err == nil {
		t.Logf("pixel_format = %s", pixFmt.String())
		if pixFmt != testPixFmt {
			t.Logf("Note: pixel_format value differs: expected %s, got %s", testPixFmt.String(), pixFmt.String())
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_sample_fmt and av_opt_get_sample_fmt with audio encoder

func Test_avutil_opt_007(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_MP2)
	if codec == nil {
		t.Skip("MP2 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting sample format option
	const testSampleFmt = ff.AV_SAMPLE_FMT_FLTP
	if err := ff.AVUtil_opt_set_sample_fmt(unsafe.Pointer(ctx), "sample_fmt", testSampleFmt, 0); err != nil {
		t.Logf("Warning: could not set sample_fmt option: %v", err)
	}

	// Test getting sample format option
	sampleFmt, err := ff.AVUtil_opt_get_sample_fmt(unsafe.Pointer(ctx), "sample_fmt", 0)
	if err == nil {
		t.Logf("sample_fmt = %s", sampleFmt.String())
		if sampleFmt != testSampleFmt {
			t.Logf("Note: sample_fmt value differs: expected %s, got %s", testSampleFmt.String(), sampleFmt.String())
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_video_rate and av_opt_get_video_rate

func Test_avutil_opt_008(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting video rate option (framerate)
	testRate := ff.AVUtil_rational(30, 1)
	if err := ff.AVUtil_opt_set_video_rate(unsafe.Pointer(ctx), "r", testRate, 0); err != nil {
		t.Logf("Warning: could not set framerate option: %v", err)
	}

	// Test getting video rate option
	rate, err := ff.AVUtil_opt_get_video_rate(unsafe.Pointer(ctx), "r", 0)
	if err == nil {
		t.Logf("framerate = %d/%d (%.2f fps)", rate.Num(), rate.Den(), ff.AVUtil_rational_q2d(rate))
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_channel_layout and av_opt_get_channel_layout

func Test_avutil_opt_009(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_MP2)
	if codec == nil {
		t.Skip("MP2 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting channel layout option
	testLayout := ff.AV_CHANNEL_LAYOUT_STEREO()
	if err := ff.AVUtil_opt_set_channel_layout(unsafe.Pointer(ctx), "channel_layout", &testLayout, 0); err != nil {
		t.Logf("Warning: could not set channel_layout option: %v", err)
	}

	// Test getting channel layout option
	layout, err := ff.AVUtil_opt_get_channel_layout(unsafe.Pointer(ctx), "channel_layout", 0)
	if err == nil {
		if desc, err2 := ff.AVUtil_channel_layout_describe(layout); err2 == nil {
			t.Logf("channel_layout = %s (channels: %d)", desc, layout.NumChannels())
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_bin

func Test_avutil_opt_010(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting binary option (extradata if supported)
	testData := []byte{0x01, 0x02, 0x03, 0x04}
	if err := ff.AVUtil_opt_set_bin(unsafe.Pointer(ctx), "extradata", testData, 0); err != nil {
		t.Logf("Info: binary option test completed (extradata may not be settable via opt): %v", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_find

func Test_avutil_opt_011(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test finding an option
	opt := ff.AVUtil_opt_find(unsafe.Pointer(ctx), "b", "", 0, 0)
	if opt != nil {
		t.Logf("Found option 'b' (bitrate)")
	} else {
		t.Error("Failed to find option 'b'")
	}

	// Test finding non-existent option
	opt = ff.AVUtil_opt_find(unsafe.Pointer(ctx), "nonexistent_option", "", 0, 0)
	if opt == nil {
		t.Logf("Correctly returned nil for non-existent option")
	} else {
		t.Error("Should have returned nil for non-existent option")
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_find2

func Test_avutil_opt_012(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test finding an option with target object
	opt, targetObj := ff.AVUtil_opt_find2(unsafe.Pointer(ctx), "b", "", 0, 0)
	if opt != nil {
		t.Logf("Found option 'b' (bitrate)")
		if targetObj != nil {
			t.Logf("Target object is non-nil")
		}
	} else {
		t.Error("Failed to find option 'b'")
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_defaults

func Test_avutil_opt_013(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting defaults
	ff.AVUtil_opt_set_defaults(unsafe.Pointer(ctx))
	t.Logf("Set defaults on codec context")

	// Verify a default value was set
	if bitrate, err := ff.AVUtil_opt_get_int(unsafe.Pointer(ctx), "b", 0); err == nil {
		t.Logf("Default bitrate = %d", bitrate)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_set_defaults2

func Test_avutil_opt_014(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test setting defaults with flags
	ff.AVUtil_opt_set_defaults2(unsafe.Pointer(ctx), 0, 0)
	t.Logf("Set defaults2 on codec context")
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_copy

func Test_avutil_opt_015(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	src := ff.AVCodec_alloc_context(codec)
	if src == nil {
		t.Fatal("Failed to allocate source codec context")
	}
	defer ff.AVCodec_free_context(src)

	dst := ff.AVCodec_alloc_context(codec)
	if dst == nil {
		t.Fatal("Failed to allocate destination codec context")
	}
	defer ff.AVCodec_free_context(dst)

	// Set some options on source
	const testBitrate = 500000
	if err := ff.AVUtil_opt_set_int(unsafe.Pointer(src), "b", testBitrate, 0); err != nil {
		t.Logf("Warning: could not set bitrate on source: %v", err)
	}

	// Copy options from source to destination
	if err := ff.AVUtil_opt_copy(unsafe.Pointer(dst), unsafe.Pointer(src)); err != nil {
		t.Errorf("Failed to copy options: %v", err)
	} else {
		t.Logf("Successfully copied options")

		// Verify the copy
		if bitrate, err := ff.AVUtil_opt_get_int(unsafe.Pointer(dst), "b", 0); err == nil {
			t.Logf("Destination bitrate = %d", bitrate)
			if bitrate != testBitrate {
				t.Logf("Note: bitrate value differs: expected %d, got %d", testBitrate, bitrate)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_is_set_to_default_by_name

func Test_avutil_opt_017(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Set defaults first
	ff.AVUtil_opt_set_defaults(unsafe.Pointer(ctx))

	// Check if option is set to default
	result := ff.AVUtil_opt_is_set_to_default_by_name(unsafe.Pointer(ctx), "b", 0)
	if result == 1 {
		t.Logf("Option 'b' is set to default")
	} else if result == 0 {
		t.Logf("Option 'b' is not set to default")
	} else {
		t.Logf("Error checking if option is set to default: %d", result)
	}

	// Change the option value
	if err := ff.AVUtil_opt_set_int(unsafe.Pointer(ctx), "b", 1000000, 0); err == nil {
		// Check again
		result := ff.AVUtil_opt_is_set_to_default_by_name(unsafe.Pointer(ctx), "b", 0)
		if result == 0 {
			t.Logf("Option 'b' is now not set to default (as expected)")
		} else if result == 1 {
			t.Logf("Note: Option 'b' still reports as default")
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_serialize

func Test_avutil_opt_018(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Set some options
	ff.AVUtil_opt_set_int(unsafe.Pointer(ctx), "b", 400000, 0)
	ff.AVUtil_opt_set(unsafe.Pointer(ctx), "preset", "fast", 0)

	// Serialize options
	serialized := ff.AVUtil_opt_serialize(unsafe.Pointer(ctx), 0, 0, '=', ':')
	if serialized != "" {
		t.Logf("Serialized options: %s", serialized)
	} else {
		t.Logf("Serialization returned empty string (may be normal)")
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_show2

func Test_avutil_opt_019(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Show options (outputs to FFmpeg log)
	if err := ff.AVUtil_opt_show2(unsafe.Pointer(ctx), nil, 0, 0); err != nil {
		t.Errorf("Failed to show options: %v", err)
	} else {
		t.Logf("Options shown (check FFmpeg log output)")
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_query_ranges

func Test_avutil_opt_020(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Query ranges for an option (not all options support range queries)
	var ranges *ff.AVOptionRanges
	if err := ff.AVUtil_opt_query_ranges(&ranges, unsafe.Pointer(ctx), "b", 0); err != nil {
		t.Logf("Option 'b' does not support range queries (expected): %v", err)
	} else if ranges != nil {
		t.Logf("Successfully queried ranges for option 'b'")
		ff.AVUtil_opt_freep_ranges(&ranges)
		if ranges == nil {
			t.Logf("Ranges successfully freed")
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST av_opt_next and av_opt_list

func Test_avutil_opt_021(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test iterating with av_opt_next
	count := 0
	var prev *ff.AVOption
	for {
		opt := ff.AVUtil_opt_next(unsafe.Pointer(ctx), prev)
		if opt == nil {
			break
		}
		count++
		if count <= 5 {
			t.Logf("Option %d: name=%s, type=%v, help=%s", count, opt.Name(), opt.Type(), opt.Help())
		}
		prev = opt
	}
	t.Logf("Total options found with av_opt_next: %d", count)

	if count == 0 {
		t.Error("Expected to find at least some options")
	}
}

func Test_avutil_opt_022(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Test getting all options as a slice
	options := ff.AVUtil_opt_list(unsafe.Pointer(ctx))
	t.Logf("Total options found with av_opt_list: %d", len(options))

	if len(options) == 0 {
		t.Error("Expected to find at least some options")
	}

	// Check first few options
	for i, opt := range options {
		if i >= 5 {
			break
		}
		t.Logf("Option %d: name=%s, type=%v", i+1, opt.Name(), opt.Type())
	}

	// Verify we can find known options
	foundBitrate := false
	for _, opt := range options {
		if opt.Name() == "b" {
			foundBitrate = true
			t.Logf("Found bitrate option: offset=%d", opt.Offset())
			break
		}
	}

	if !foundBitrate {
		t.Error("Expected to find 'b' (bitrate) option")
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST AVOption JSON marshaling

func Test_avutil_opt_023(t *testing.T) {
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		t.Fatal("Failed to allocate codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Get all options
	options := ff.AVUtil_opt_list(unsafe.Pointer(ctx))
	if len(options) == 0 {
		t.Fatal("Expected to find options")
	}

	// Marshal first option to JSON
	data, err := json.Marshal(options[0])
	if err != nil {
		t.Fatalf("Failed to marshal option to JSON: %v", err)
	}
	t.Logf("First option JSON: %s", string(data))

	// Verify JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check required fields
	if _, ok := result["name"]; !ok {
		t.Error("JSON missing 'name' field")
	}
	if _, ok := result["type"]; !ok {
		t.Error("JSON missing 'type' field")
	}

	// Verify type is a string, not a number
	if typeVal, ok := result["type"].(string); !ok {
		t.Errorf("JSON 'type' field should be a string, got %T", result["type"])
	} else {
		t.Logf("Type is string: %s", typeVal)
	}

	// Marshal array of options
	firstFive := options[:5]
	arrayData, err := json.Marshal(firstFive)
	if err != nil {
		t.Fatalf("Failed to marshal options array to JSON: %v", err)
	}
	t.Logf("First 5 options JSON: %s", string(arrayData))

	// Pretty print
	prettyData, err := json.MarshalIndent(firstFive, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal pretty JSON: %v", err)
	}
	t.Logf("Pretty JSON:\n%s", string(prettyData))
}

////////////////////////////////////////////////////////////////////////////////
// TEST AVUtil_opt_list_from_class (FAKE_OBJ trick)

func Test_avutil_opt_024(t *testing.T) {
	// Test FAKE_OBJ trick with AVCodec
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_H264)
	if codec == nil {
		t.Skip("H264 encoder not found")
	}

	class := codec.PrivClass()
	if class == nil {
		t.Skip("H264 encoder has no priv_class")
	}

	// Use FAKE_OBJ trick to enumerate options without context allocation
	options := ff.AVUtil_opt_list_from_class(class)
	if len(options) == 0 {
		t.Fatal("Expected options from FAKE_OBJ trick, got none")
	}

	t.Logf("Found %d options via FAKE_OBJ for H264 encoder", len(options))

	// Verify first few options have valid data
	for i := 0; i < min(5, len(options)); i++ {
		opt := options[i]
		name := opt.Name()
		help := opt.Help()
		optType := opt.Type()

		if name == "" {
			t.Errorf("Option %d has empty name", i)
		}
		t.Logf("Option %d: name=%s, type=%v, help=%s", i, name, optType, help)
	}
}

func Test_avutil_opt_025(t *testing.T) {
	// Test FAKE_OBJ with AVInputFormat
	format := ff.AVFormat_find_input_format("mpegts")
	if format == nil {
		t.Skip("mpegts input format not found")
	}

	class := format.PrivClass()
	if class == nil {
		t.Skip("mpegts has no priv_class")
	}

	options := ff.AVUtil_opt_list_from_class(class)
	if len(options) == 0 {
		t.Fatal("Expected options from FAKE_OBJ trick for mpegts, got none")
	}

	t.Logf("Found %d options via FAKE_OBJ for mpegts format", len(options))

	// Look for known mpegts options
	foundResyncSize := false
	for _, opt := range options {
		if opt.Name() == "resync_size" {
			foundResyncSize = true
			t.Logf("Found resync_size option: help=%s, type=%v", opt.Help(), opt.Type())
			break
		}
	}

	if !foundResyncSize {
		t.Error("Expected to find 'resync_size' option in mpegts format")
	}
}

func Test_avutil_opt_026(t *testing.T) {
	// Test FAKE_OBJ with AVOutputFormat
	format := ff.AVFormat_guess_format("mp4", "", "")
	if format == nil {
		t.Skip("mp4 output format not found")
	}

	class := format.PrivClass()
	if class == nil {
		t.Skip("mp4 has no priv_class")
	}

	options := ff.AVUtil_opt_list_from_class(class)
	t.Logf("Found %d options via FAKE_OBJ for mp4 format", len(options))

	// Some formats may have no options, that's OK
	if len(options) > 0 {
		t.Logf("First option: %s (%v)", options[0].Name(), options[0].Type())
	}
}

func Test_avutil_opt_027(t *testing.T) {
	// Test FAKE_OBJ with AVFilter
	filter := ff.AVFilter_get_by_name("scale")
	if filter == nil {
		t.Skip("scale filter not found")
	}

	class := filter.PrivClass()
	if class == nil {
		t.Skip("scale filter has no priv_class")
	}

	options := ff.AVUtil_opt_list_from_class(class)
	if len(options) == 0 {
		t.Fatal("Expected options from FAKE_OBJ trick for scale filter, got none")
	}

	t.Logf("Found %d options via FAKE_OBJ for scale filter", len(options))

	// Look for known scale options
	foundWidth := false
	foundHeight := false
	for _, opt := range options {
		switch opt.Name() {
		case "w", "width":
			foundWidth = true
			t.Logf("Found width option: help=%s, type=%v", opt.Help(), opt.Type())
		case "h", "height":
			foundHeight = true
			t.Logf("Found height option: help=%s, type=%v", opt.Help(), opt.Type())
		}
	}

	if !foundWidth {
		t.Error("Expected to find width option in scale filter")
	}
	if !foundHeight {
		t.Error("Expected to find height option in scale filter")
	}
}
