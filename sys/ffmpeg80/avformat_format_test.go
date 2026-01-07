package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS - STRING REPRESENTATION

func Test_AVFormat_String_001(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("AVFMT_NONE", AVFMT_NONE.String())
}

func Test_AVFormat_String_002(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("AVFMT_NOFILE", AVFMT_NOFILE.String())
	assert.Equal("AVFMT_NEEDNUMBER", AVFMT_NEEDNUMBER.String())
	assert.Equal("AVFMT_EXPERIMENTAL", AVFMT_EXPERIMENTAL.String())
	assert.Equal("AVFMT_SHOWIDS", AVFMT_SHOWIDS.String())
	assert.Equal("AVFMT_GLOBALHEADER", AVFMT_GLOBALHEADER.String())
	assert.Equal("AVFMT_NOTIMESTAMPS", AVFMT_NOTIMESTAMPS.String())
	assert.Equal("AVFMT_GENERICINDEX", AVFMT_GENERICINDEX.String())
	assert.Equal("AVFMT_TSDISCONT", AVFMT_TSDISCONT.String())
	assert.Equal("AVFMT_VARIABLEFPS", AVFMT_VARIABLEFPS.String())
	assert.Equal("AVFMT_NODIMENSIONS", AVFMT_NODIMENSIONS.String())
	assert.Equal("AVFMT_NOSTREAMS", AVFMT_NOSTREAMS.String())
	assert.Equal("AVFMT_NOBINSEARCH", AVFMT_NOBINSEARCH.String())
	assert.Equal("AVFMT_NOGENSEARCH", AVFMT_NOGENSEARCH.String())
	assert.Equal("AVFMT_NOBYTESEEK", AVFMT_NOBYTESEEK.String())
	assert.Equal("AVFMT_TS_NONSTRICT", AVFMT_TS_NONSTRICT.String())
	assert.Equal("AVFMT_TS_NEGATIVE", AVFMT_TS_NEGATIVE.String())
	assert.Equal("AVFMT_SEEK_TO_PTS", AVFMT_SEEK_TO_PTS.String())
}

func Test_AVFormat_String_003(t *testing.T) {
	// Test combined formats
	assert := assert.New(t)
	formats := AVFMT_NOFILE | AVFMT_NEEDNUMBER
	assert.Contains(formats.String(), "AVFMT_NOFILE")
	assert.Contains(formats.String(), "AVFMT_NEEDNUMBER")
}

func Test_AVFormat_String_004(t *testing.T) {
	// Test unknown format value
	assert := assert.New(t)
	format := AVFormat(0x9999999)
	assert.Contains(format.String(), "AVFormat")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - FLAG STRING

func Test_AVFormat_FlagString_001(t *testing.T) {
	assert := assert.New(t)

	formats := []AVFormat{
		AVFMT_NONE,
		AVFMT_NOFILE,
		AVFMT_NEEDNUMBER,
		AVFMT_EXPERIMENTAL,
		AVFMT_SHOWIDS,
		AVFMT_GLOBALHEADER,
		AVFMT_NOTIMESTAMPS,
		AVFMT_GENERICINDEX,
		AVFMT_TSDISCONT,
		AVFMT_VARIABLEFPS,
		AVFMT_NODIMENSIONS,
		AVFMT_NOSTREAMS,
		AVFMT_NOBINSEARCH,
		AVFMT_NOGENSEARCH,
		AVFMT_NOBYTESEEK,
		AVFMT_TS_NONSTRICT,
		AVFMT_TS_NEGATIVE,
		AVFMT_SEEK_TO_PTS,
	}

	for _, format := range formats {
		result := format.FlagString()
		assert.NotEmpty(result)
		assert.Contains(result, "AVFMT_")
	}
}

func Test_AVFormat_FlagString_002(t *testing.T) {
	// Test FlagString returns formatted hex for unknown/combined values
	assert := assert.New(t)
	format := AVFMT_NOFILE | AVFMT_NEEDNUMBER
	result := format.FlagString()
	assert.Contains(result, "AVFormat")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - JSON MARSHALING

func Test_AVFormat_JSON_001(t *testing.T) {
	assert := assert.New(t)

	data, err := json.Marshal(AVFMT_NOFILE)
	assert.NoError(err)
	assert.Contains(string(data), "AVFMT_NOFILE")
}

func Test_AVFormat_JSON_002(t *testing.T) {
	assert := assert.New(t)

	formats := []AVFormat{
		AVFMT_NONE,
		AVFMT_NEEDNUMBER,
		AVFMT_EXPERIMENTAL,
		AVFMT_GLOBALHEADER,
		AVFMT_SEEK_TO_PTS,
	}

	for _, format := range formats {
		data, err := json.Marshal(format)
		assert.NoError(err)
		assert.NotEmpty(data)
	}
}

func Test_AVFormat_JSON_003(t *testing.T) {
	// Test JSON marshaling of combined formats
	assert := assert.New(t)

	format := AVFMT_NOFILE | AVFMT_GLOBALHEADER | AVFMT_VARIABLEFPS
	data, err := json.Marshal(format)
	assert.NoError(err)
	assert.Contains(string(data), "AVFMT_")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - IS METHOD

func Test_AVFormat_Is_001(t *testing.T) {
	assert := assert.New(t)

	formats := AVFMT_NOFILE | AVFMT_NEEDNUMBER
	assert.True(formats.Is(AVFMT_NOFILE))
	assert.True(formats.Is(AVFMT_NEEDNUMBER))
	assert.False(formats.Is(AVFMT_EXPERIMENTAL))
}

func Test_AVFormat_Is_002(t *testing.T) {
	assert := assert.New(t)

	formats := AVFMT_GLOBALHEADER | AVFMT_VARIABLEFPS
	assert.True(formats.Is(AVFMT_GLOBALHEADER))
	assert.True(formats.Is(AVFMT_VARIABLEFPS))
	assert.False(formats.Is(AVFMT_NOTIMESTAMPS))
}

func Test_AVFormat_Is_003(t *testing.T) {
	// Test Is with combined formats
	assert := assert.New(t)

	formats := AVFMT_NOFILE | AVFMT_NEEDNUMBER | AVFMT_EXPERIMENTAL
	combined := AVFMT_NOFILE | AVFMT_NEEDNUMBER
	// Note: Is() checks if ANY of the bits are set, not all
	assert.True(formats.Is(combined))
}

func Test_AVFormat_Is_004(t *testing.T) {
	// Test Is with NONE
	assert := assert.New(t)

	formats := AVFMT_NOFILE | AVFMT_NEEDNUMBER
	assert.False(formats.Is(AVFMT_NONE))
	assert.False(AVFMT_NONE.Is(AVFMT_NOFILE))
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - BITWISE OPERATIONS

func Test_AVFormat_Bitwise_001(t *testing.T) {
	// Test OR operation
	assert := assert.New(t)

	formats := AVFMT_NOFILE | AVFMT_NEEDNUMBER
	assert.True(formats.Is(AVFMT_NOFILE))
	assert.True(formats.Is(AVFMT_NEEDNUMBER))
}

func Test_AVFormat_Bitwise_002(t *testing.T) {
	// Test AND operation
	assert := assert.New(t)

	formats := AVFMT_NOFILE | AVFMT_NEEDNUMBER
	result := formats & AVFMT_NOFILE
	assert.Equal(AVFMT_NOFILE, result)
}

func Test_AVFormat_Bitwise_003(t *testing.T) {
	// Test XOR operation
	assert := assert.New(t)

	formats := AVFMT_NOFILE | AVFMT_NEEDNUMBER
	result := formats ^ AVFMT_NOFILE
	assert.Equal(AVFMT_NEEDNUMBER, result)
}

func Test_AVFormat_Bitwise_004(t *testing.T) {
	// Test clearing a format flag
	assert := assert.New(t)

	formats := AVFMT_NOFILE | AVFMT_NEEDNUMBER | AVFMT_EXPERIMENTAL
	formats = formats &^ AVFMT_NEEDNUMBER // Clear NEEDNUMBER
	assert.True(formats.Is(AVFMT_NOFILE))
	assert.False(formats.Is(AVFMT_NEEDNUMBER))
	assert.True(formats.Is(AVFMT_EXPERIMENTAL))
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - CONSTANTS VALIDATION

func Test_AVFormat_Constants_001(t *testing.T) {
	// Verify all constants are unique
	assert := assert.New(t)

	formats := []AVFormat{
		AVFMT_NOFILE,
		AVFMT_NEEDNUMBER,
		AVFMT_EXPERIMENTAL,
		AVFMT_SHOWIDS,
		AVFMT_GLOBALHEADER,
		AVFMT_NOTIMESTAMPS,
		AVFMT_GENERICINDEX,
		AVFMT_TSDISCONT,
		AVFMT_VARIABLEFPS,
		AVFMT_NODIMENSIONS,
		AVFMT_NOSTREAMS,
		AVFMT_NOBINSEARCH,
		AVFMT_NOGENSEARCH,
		AVFMT_NOBYTESEEK,
		AVFMT_TS_NONSTRICT,
		AVFMT_TS_NEGATIVE,
		AVFMT_SEEK_TO_PTS,
	}

	// Check all formats are unique
	seen := make(map[AVFormat]bool)
	for _, format := range formats {
		assert.False(seen[format], "Duplicate format found: %v", format)
		seen[format] = true
	}
}

func Test_AVFormat_Constants_002(t *testing.T) {
	// Verify NONE is zero
	assert := assert.New(t)
	assert.Equal(AVFormat(0), AVFMT_NONE)
}

func Test_AVFormat_Constants_003(t *testing.T) {
	// Verify MIN and MAX are set correctly
	assert := assert.New(t)
	assert.Equal(AVFMT_NOFILE, AVFMT_MIN)
	assert.Equal(AVFMT_SEEK_TO_PTS, AVFMT_MAX)
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - COMMON PATTERNS

func Test_AVFormat_Pattern_001(t *testing.T) {
	// Test common demuxer format flags
	assert := assert.New(t)

	demuxerFlags := AVFMT_NOFILE | AVFMT_GENERICINDEX
	assert.True(demuxerFlags.Is(AVFMT_NOFILE))
	assert.True(demuxerFlags.Is(AVFMT_GENERICINDEX))
	assert.Contains(demuxerFlags.String(), "NOFILE")
	assert.Contains(demuxerFlags.String(), "GENERICINDEX")
}

func Test_AVFormat_Pattern_002(t *testing.T) {
	// Test muxer format flags
	assert := assert.New(t)

	muxerFlags := AVFMT_GLOBALHEADER | AVFMT_VARIABLEFPS | AVFMT_NOTIMESTAMPS
	assert.True(muxerFlags.Is(AVFMT_GLOBALHEADER))
	assert.True(muxerFlags.Is(AVFMT_VARIABLEFPS))
	assert.True(muxerFlags.Is(AVFMT_NOTIMESTAMPS))
}

func Test_AVFormat_Pattern_003(t *testing.T) {
	// Test timestamp-related flags
	assert := assert.New(t)

	timestampFlags := AVFMT_NOTIMESTAMPS | AVFMT_TSDISCONT | AVFMT_TS_NONSTRICT
	assert.True(timestampFlags.Is(AVFMT_NOTIMESTAMPS))
	assert.True(timestampFlags.Is(AVFMT_TSDISCONT))
	assert.True(timestampFlags.Is(AVFMT_TS_NONSTRICT))
}

func Test_AVFormat_Pattern_004(t *testing.T) {
	// Test seeking-related flags
	assert := assert.New(t)

	seekFlags := AVFMT_NOBINSEARCH | AVFMT_NOGENSEARCH | AVFMT_NOBYTESEEK | AVFMT_SEEK_TO_PTS
	assert.True(seekFlags.Is(AVFMT_NOBINSEARCH))
	assert.True(seekFlags.Is(AVFMT_NOGENSEARCH))
	assert.True(seekFlags.Is(AVFMT_NOBYTESEEK))
	assert.True(seekFlags.Is(AVFMT_SEEK_TO_PTS))
}

func Test_AVFormat_Pattern_005(t *testing.T) {
	// Test format requirements
	assert := assert.New(t)

	requirementFlags := AVFMT_NEEDNUMBER | AVFMT_NODIMENSIONS | AVFMT_NOSTREAMS
	assert.True(requirementFlags.Is(AVFMT_NEEDNUMBER))
	assert.True(requirementFlags.Is(AVFMT_NODIMENSIONS))
	assert.True(requirementFlags.Is(AVFMT_NOSTREAMS))
}

func Test_AVFormat_Pattern_006(t *testing.T) {
	// Test experimental format
	assert := assert.New(t)

	experimentalFormat := AVFMT_EXPERIMENTAL | AVFMT_NOFILE
	assert.True(experimentalFormat.Is(AVFMT_EXPERIMENTAL))
	assert.Contains(experimentalFormat.String(), "EXPERIMENTAL")
}

////////////////////////////////////////////////////////////////////////////////
// TEST PrivClass on AVInputFormat

func Test_AVInputFormat_PrivClass(t *testing.T) {
	assert := assert.New(t)

	// Test input format with priv_class
	format := AVFormat_find_input_format("mpegts")
	if format == nil {
		t.Skip("mpegts input format not found")
	}

	class := format.PrivClass()
	if class == nil {
		t.Skip("mpegts has no priv_class")
	}

	assert.NotNil(class, "mpegts should have priv_class")

	// The class should have a valid name
	className := class.Name()
	assert.NotEmpty(className, "AVClass should have a name")
	t.Logf("mpegts priv_class name: %s", className)
}

func Test_AVInputFormat_PrivClass_Options(t *testing.T) {
	assert := assert.New(t)

	// Test that we can enumerate options via PrivClass
	format := AVFormat_find_input_format("mpegts")
	if format == nil {
		t.Skip("mpegts input format not found")
	}

	class := format.PrivClass()
	if class == nil {
		t.Skip("mpegts has no priv_class")
	}

	// Use FAKE_OBJ trick to enumerate options
	options := AVUtil_opt_list_from_class(class)
	assert.NotEmpty(options, "mpegts should have options")

	t.Logf("Found %d options for mpegts via PrivClass", len(options))

	// Look for known options
	foundResyncSize := false
	for _, opt := range options {
		if opt.Name() == "resync_size" {
			foundResyncSize = true
			t.Logf("Found resync_size: help=%s, type=%v", opt.Help(), opt.Type())
			break
		}
	}

	assert.True(foundResyncSize, "Expected to find 'resync_size' option")
}

////////////////////////////////////////////////////////////////////////////////
// TEST PrivClass on AVOutputFormat

func Test_AVOutputFormat_PrivClass(t *testing.T) {
	assert := assert.New(t)

	// Test output format with priv_class
	format := AVFormat_guess_format("mp4", "", "")
	if format == nil {
		t.Skip("mp4 output format not found")
	}

	class := format.PrivClass()
	// Note: Some formats may not have priv_class, that's OK
	if class == nil {
		t.Skip("mp4 has no priv_class")
	}

	assert.NotNil(class, "mp4 should have priv_class")

	className := class.Name()
	assert.NotEmpty(className, "AVClass should have a name")
	t.Logf("mp4 priv_class name: %s", className)
}

func Test_AVOutputFormat_PrivClass_Options(t *testing.T) {
	// Test matroska which is known to have options
	format := AVFormat_guess_format("matroska", "", "")
	if format == nil {
		t.Skip("matroska output format not found")
	}

	class := format.PrivClass()
	if class == nil {
		t.Skip("matroska has no priv_class")
	}

	// Use FAKE_OBJ trick to enumerate options
	options := AVUtil_opt_list_from_class(class)
	t.Logf("Found %d options for matroska via PrivClass", len(options))

	// Matroska should have options
	if len(options) > 0 {
		for i := 0; i < min(5, len(options)); i++ {
			opt := options[i]
			t.Logf("Option %d: %s (%v) - %s", i, opt.Name(), opt.Type(), opt.Help())
		}
	}
}
