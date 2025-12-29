package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS - STRING REPRESENTATION

func Test_AVFormatFlag_String_001(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("AVFMT_FLAG_NONE", AVFMT_FLAG_NONE.String())
}

func Test_AVFormatFlag_String_002(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("AVFMT_FLAG_GENPTS", AVFMT_FLAG_GENPTS.String())
	assert.Equal("AVFMT_FLAG_IGNIDX", AVFMT_FLAG_IGNIDX.String())
	assert.Equal("AVFMT_FLAG_NONBLOCK", AVFMT_FLAG_NONBLOCK.String())
	assert.Equal("AVFMT_FLAG_IGNDTS", AVFMT_FLAG_IGNDTS.String())
	assert.Equal("AVFMT_FLAG_NOFILLIN", AVFMT_FLAG_NOFILLIN.String())
	assert.Equal("AVFMT_FLAG_NOPARSE", AVFMT_FLAG_NOPARSE.String())
	assert.Equal("AVFMT_FLAG_NOBUFFER", AVFMT_FLAG_NOBUFFER.String())
	assert.Equal("AVFMT_FLAG_CUSTOM_IO", AVFMT_FLAG_CUSTOM_IO.String())
	assert.Equal("AVFMT_FLAG_DISCARD_CORRUPT", AVFMT_FLAG_DISCARD_CORRUPT.String())
	assert.Equal("AVFMT_FLAG_FLUSH_PACKETS", AVFMT_FLAG_FLUSH_PACKETS.String())
	assert.Equal("AVFMT_FLAG_BITEXACT", AVFMT_FLAG_BITEXACT.String())
	assert.Equal("AVFMT_FLAG_SORT_DTS", AVFMT_FLAG_SORT_DTS.String())
	assert.Equal("AVFMT_FLAG_FAST_SEEK", AVFMT_FLAG_FAST_SEEK.String())
	assert.Equal("AVFMT_FLAG_AUTO_BSF", AVFMT_FLAG_AUTO_BSF.String())
}

func Test_AVFormatFlag_String_003(t *testing.T) {
	// Test combined flags
	assert := assert.New(t)
	flags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	assert.Contains(flags.String(), "AVFMT_FLAG_GENPTS")
	assert.Contains(flags.String(), "AVFMT_FLAG_IGNIDX")
}

func Test_AVFormatFlag_String_004(t *testing.T) {
	// Test unknown flag value
	assert := assert.New(t)
	flag := AVFormatFlag(0x999999)
	assert.Contains(flag.String(), "AVFormatFlag")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - FLAG STRING

func Test_AVFormatFlag_FlagString_001(t *testing.T) {
	assert := assert.New(t)

	flags := []AVFormatFlag{
		AVFMT_FLAG_NONE,
		AVFMT_FLAG_GENPTS,
		AVFMT_FLAG_IGNIDX,
		AVFMT_FLAG_NONBLOCK,
		AVFMT_FLAG_IGNDTS,
		AVFMT_FLAG_NOFILLIN,
		AVFMT_FLAG_NOPARSE,
		AVFMT_FLAG_NOBUFFER,
		AVFMT_FLAG_CUSTOM_IO,
		AVFMT_FLAG_DISCARD_CORRUPT,
		AVFMT_FLAG_FLUSH_PACKETS,
		AVFMT_FLAG_BITEXACT,
		AVFMT_FLAG_SORT_DTS,
		AVFMT_FLAG_FAST_SEEK,
		AVFMT_FLAG_AUTO_BSF,
	}

	for _, flag := range flags {
		result := flag.FlagString()
		assert.NotEmpty(result)
		assert.Contains(result, "AVFMT_FLAG_")
	}
}

func Test_AVFormatFlag_FlagString_002(t *testing.T) {
	// Test FlagString returns single flag name even for combined flags
	assert := assert.New(t)
	flag := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	result := flag.FlagString()
	// FlagString should return formatted hex for unknown/combined values
	assert.Contains(result, "AVFormatFlag")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - JSON MARSHALING

func Test_AVFormatFlag_JSON_001(t *testing.T) {
	assert := assert.New(t)

	data, err := json.Marshal(AVFMT_FLAG_GENPTS)
	assert.NoError(err)
	assert.Contains(string(data), "AVFMT_FLAG_GENPTS")
}

func Test_AVFormatFlag_JSON_002(t *testing.T) {
	assert := assert.New(t)

	flags := []AVFormatFlag{
		AVFMT_FLAG_NONE,
		AVFMT_FLAG_IGNIDX,
		AVFMT_FLAG_NONBLOCK,
		AVFMT_FLAG_CUSTOM_IO,
		AVFMT_FLAG_BITEXACT,
	}

	for _, flag := range flags {
		data, err := json.Marshal(flag)
		assert.NoError(err)
		assert.NotEmpty(data)
	}
}

func Test_AVFormatFlag_JSON_003(t *testing.T) {
	// Test JSON marshaling of combined flags
	assert := assert.New(t)

	flag := AVFMT_FLAG_GENPTS | AVFMT_FLAG_NONBLOCK | AVFMT_FLAG_BITEXACT
	data, err := json.Marshal(flag)
	assert.NoError(err)
	assert.Contains(string(data), "AVFMT_FLAG_")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - IS METHOD

func Test_AVFormatFlag_Is_001(t *testing.T) {
	assert := assert.New(t)

	flags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	assert.True(flags.Is(AVFMT_FLAG_GENPTS))
	assert.True(flags.Is(AVFMT_FLAG_IGNIDX))
	assert.False(flags.Is(AVFMT_FLAG_NONBLOCK))
}

func Test_AVFormatFlag_Is_002(t *testing.T) {
	assert := assert.New(t)

	flags := AVFMT_FLAG_CUSTOM_IO | AVFMT_FLAG_NOBUFFER
	assert.True(flags.Is(AVFMT_FLAG_CUSTOM_IO))
	assert.True(flags.Is(AVFMT_FLAG_NOBUFFER))
	assert.False(flags.Is(AVFMT_FLAG_FLUSH_PACKETS))
}

func Test_AVFormatFlag_Is_003(t *testing.T) {
	// Test Is with combined flags
	assert := assert.New(t)

	flags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX | AVFMT_FLAG_NONBLOCK
	combined := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	assert.True(flags.Is(combined))
}

func Test_AVFormatFlag_Is_004(t *testing.T) {
	// Test Is with NONE
	assert := assert.New(t)

	flags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	assert.True(flags.Is(AVFMT_FLAG_NONE))
	assert.True(AVFMT_FLAG_NONE.Is(AVFMT_FLAG_NONE))
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - BITWISE OPERATIONS

func Test_AVFormatFlag_Bitwise_001(t *testing.T) {
	// Test OR operation
	assert := assert.New(t)

	flags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	assert.True(flags.Is(AVFMT_FLAG_GENPTS))
	assert.True(flags.Is(AVFMT_FLAG_IGNIDX))
}

func Test_AVFormatFlag_Bitwise_002(t *testing.T) {
	// Test AND operation
	assert := assert.New(t)

	flags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	result := flags & AVFMT_FLAG_GENPTS
	assert.Equal(AVFMT_FLAG_GENPTS, result)
}

func Test_AVFormatFlag_Bitwise_003(t *testing.T) {
	// Test XOR operation
	assert := assert.New(t)

	flags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	result := flags ^ AVFMT_FLAG_GENPTS
	assert.Equal(AVFMT_FLAG_IGNIDX, result)
}

func Test_AVFormatFlag_Bitwise_004(t *testing.T) {
	// Test clearing a flag
	assert := assert.New(t)

	flags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX | AVFMT_FLAG_NONBLOCK
	flags = flags &^ AVFMT_FLAG_IGNIDX // Clear IGNIDX
	assert.True(flags.Is(AVFMT_FLAG_GENPTS))
	assert.False(flags.Is(AVFMT_FLAG_IGNIDX))
	assert.True(flags.Is(AVFMT_FLAG_NONBLOCK))
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - CONSTANTS VALIDATION

func Test_AVFormatFlag_Constants_001(t *testing.T) {
	// Verify all constants are unique powers of 2 (except NONE and combined flags)
	assert := assert.New(t)

	flags := []AVFormatFlag{
		AVFMT_FLAG_GENPTS,
		AVFMT_FLAG_IGNIDX,
		AVFMT_FLAG_NONBLOCK,
		AVFMT_FLAG_IGNDTS,
		AVFMT_FLAG_NOFILLIN,
		AVFMT_FLAG_NOPARSE,
		AVFMT_FLAG_NOBUFFER,
		AVFMT_FLAG_CUSTOM_IO,
		AVFMT_FLAG_DISCARD_CORRUPT,
		AVFMT_FLAG_FLUSH_PACKETS,
		AVFMT_FLAG_BITEXACT,
		AVFMT_FLAG_SORT_DTS,
		AVFMT_FLAG_FAST_SEEK,
		AVFMT_FLAG_AUTO_BSF,
	}

	// Check all flags are unique
	seen := make(map[AVFormatFlag]bool)
	for _, flag := range flags {
		assert.False(seen[flag], "Duplicate flag found: %v", flag)
		seen[flag] = true
	}
}

func Test_AVFormatFlag_Constants_002(t *testing.T) {
	// Verify NONE is zero
	assert := assert.New(t)
	assert.Equal(AVFormatFlag(0), AVFMT_FLAG_NONE)
}

func Test_AVFormatFlag_Constants_003(t *testing.T) {
	// Verify MIN and MAX are set correctly
	assert := assert.New(t)
	assert.Equal(AVFMT_FLAG_GENPTS, AVFMT_FLAG_MIN)
	assert.Equal(AVFMT_FLAG_AUTO_BSF, AVFMT_FLAG_MAX)
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - COMMON PATTERNS

func Test_AVFormatFlag_Pattern_001(t *testing.T) {
	// Test common demuxer flags
	assert := assert.New(t)

	demuxerFlags := AVFMT_FLAG_GENPTS | AVFMT_FLAG_IGNIDX
	assert.True(demuxerFlags.Is(AVFMT_FLAG_GENPTS))
	assert.True(demuxerFlags.Is(AVFMT_FLAG_IGNIDX))
	assert.Contains(demuxerFlags.String(), "GENPTS")
	assert.Contains(demuxerFlags.String(), "IGNIDX")
}

func Test_AVFormatFlag_Pattern_002(t *testing.T) {
	// Test custom I/O flags
	assert := assert.New(t)

	customIOFlags := AVFMT_FLAG_CUSTOM_IO | AVFMT_FLAG_NOBUFFER
	assert.True(customIOFlags.Is(AVFMT_FLAG_CUSTOM_IO))
	assert.True(customIOFlags.Is(AVFMT_FLAG_NOBUFFER))
}

func Test_AVFormatFlag_Pattern_003(t *testing.T) {
	// Test muxer flags
	assert := assert.New(t)

	muxerFlags := AVFMT_FLAG_BITEXACT | AVFMT_FLAG_FLUSH_PACKETS | AVFMT_FLAG_AUTO_BSF
	assert.True(muxerFlags.Is(AVFMT_FLAG_BITEXACT))
	assert.True(muxerFlags.Is(AVFMT_FLAG_FLUSH_PACKETS))
	assert.True(muxerFlags.Is(AVFMT_FLAG_AUTO_BSF))
}

func Test_AVFormatFlag_Pattern_004(t *testing.T) {
	// Test parser-related flags
	assert := assert.New(t)

	noParseFlags := AVFMT_FLAG_NOPARSE | AVFMT_FLAG_NOFILLIN
	assert.True(noParseFlags.Is(AVFMT_FLAG_NOPARSE))
	assert.True(noParseFlags.Is(AVFMT_FLAG_NOFILLIN))
}
