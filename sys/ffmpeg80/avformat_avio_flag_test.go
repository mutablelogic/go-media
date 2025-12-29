package ffmpeg

import (
	"encoding/json"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS - STRING REPRESENTATION

func Test_AVIOFlag_String_001(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("AVIO_FLAG_NONE", AVIO_FLAG_NONE.String())
}

func Test_AVIOFlag_String_002(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("AVIO_FLAG_READ", AVIO_FLAG_READ.String())
	assert.Equal("AVIO_FLAG_WRITE", AVIO_FLAG_WRITE.String())
	// AVIO_FLAG_READ_WRITE is a combination, so it shows both flags
	readWrite := AVIO_FLAG_READ_WRITE.String()
	assert.Contains(readWrite, "AVIO_FLAG_READ")
	assert.Contains(readWrite, "AVIO_FLAG_WRITE")
	assert.Equal("AVIO_FLAG_NONBLOCK", AVIO_FLAG_NONBLOCK.String())
	assert.Equal("AVIO_FLAG_DIRECT", AVIO_FLAG_DIRECT.String())
}

func Test_AVIOFlag_String_003(t *testing.T) {
	// Test combined flags
	assert := assert.New(t)
	flags := AVIO_FLAG_READ | AVIO_FLAG_NONBLOCK
	assert.Contains(flags.String(), "AVIO_FLAG_READ")
	assert.Contains(flags.String(), "AVIO_FLAG_NONBLOCK")
}

func Test_AVIOFlag_String_004(t *testing.T) {
	// Test unknown flag value
	assert := assert.New(t)
	flag := AVIOFlag(0x9999)
	assert.Contains(flag.String(), "AVIOFlag")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - FLAG STRING

func Test_AVIOFlag_FlagString_001(t *testing.T) {
	assert := assert.New(t)

	flags := []AVIOFlag{
		AVIO_FLAG_NONE,
		AVIO_FLAG_READ,
		AVIO_FLAG_WRITE,
		AVIO_FLAG_READ_WRITE,
		AVIO_FLAG_NONBLOCK,
		AVIO_FLAG_DIRECT,
	}

	for _, flag := range flags {
		result := flag.FlagString()
		assert.NotEmpty(result)
		assert.Contains(result, "AVIO_FLAG_")
	}
}

func Test_AVIOFlag_FlagString_002(t *testing.T) {
	// Test FlagString returns formatted hex for unknown/combined values
	assert := assert.New(t)
	flag := AVIO_FLAG_READ | AVIO_FLAG_WRITE
	result := flag.FlagString()
	// READ_WRITE is a special combined flag
	assert.Contains(result, "AVIO_FLAG")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - JSON MARSHALING

func Test_AVIOFlag_JSON_001(t *testing.T) {
	assert := assert.New(t)

	data, err := json.Marshal(AVIO_FLAG_READ)
	assert.NoError(err)
	assert.Contains(string(data), "AVIO_FLAG_READ")
}

func Test_AVIOFlag_JSON_002(t *testing.T) {
	assert := assert.New(t)

	flags := []AVIOFlag{
		AVIO_FLAG_NONE,
		AVIO_FLAG_READ,
		AVIO_FLAG_WRITE,
		AVIO_FLAG_NONBLOCK,
		AVIO_FLAG_DIRECT,
	}

	for _, flag := range flags {
		data, err := json.Marshal(flag)
		assert.NoError(err)
		assert.NotEmpty(data)
	}
}

func Test_AVIOFlag_JSON_003(t *testing.T) {
	// Test JSON marshaling of combined flags
	assert := assert.New(t)

	flag := AVIO_FLAG_READ | AVIO_FLAG_NONBLOCK
	data, err := json.Marshal(flag)
	assert.NoError(err)
	assert.Contains(string(data), "AVIO_FLAG_")
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - IS METHOD

func Test_AVIOFlag_Is_001(t *testing.T) {
	assert := assert.New(t)

	flags := AVIO_FLAG_READ | AVIO_FLAG_NONBLOCK
	assert.True(flags.Is(AVIO_FLAG_READ))
	assert.True(flags.Is(AVIO_FLAG_NONBLOCK))
	assert.False(flags.Is(AVIO_FLAG_WRITE))
}

func Test_AVIOFlag_Is_002(t *testing.T) {
	assert := assert.New(t)

	flags := AVIO_FLAG_WRITE | AVIO_FLAG_DIRECT
	assert.True(flags.Is(AVIO_FLAG_WRITE))
	assert.True(flags.Is(AVIO_FLAG_DIRECT))
	assert.False(flags.Is(AVIO_FLAG_READ))
}

func Test_AVIOFlag_Is_003(t *testing.T) {
	// Test Is with READ_WRITE pseudo flag
	assert := assert.New(t)

	flags := AVIO_FLAG_READ_WRITE
	assert.True(flags.Is(AVIO_FLAG_READ))
	assert.True(flags.Is(AVIO_FLAG_WRITE))
}

func Test_AVIOFlag_Is_004(t *testing.T) {
	// Test Is with NONE
	assert := assert.New(t)

	flags := AVIO_FLAG_READ | AVIO_FLAG_WRITE
	assert.False(flags.Is(AVIO_FLAG_NONE))
	assert.False(AVIO_FLAG_NONE.Is(AVIO_FLAG_READ))
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - BITWISE OPERATIONS

func Test_AVIOFlag_Bitwise_001(t *testing.T) {
	// Test OR operation
	assert := assert.New(t)

	flags := AVIO_FLAG_READ | AVIO_FLAG_NONBLOCK
	assert.True(flags.Is(AVIO_FLAG_READ))
	assert.True(flags.Is(AVIO_FLAG_NONBLOCK))
}

func Test_AVIOFlag_Bitwise_002(t *testing.T) {
	// Test AND operation
	assert := assert.New(t)

	flags := AVIO_FLAG_READ | AVIO_FLAG_NONBLOCK
	result := flags & AVIO_FLAG_READ
	assert.Equal(AVIO_FLAG_READ, result)
}

func Test_AVIOFlag_Bitwise_003(t *testing.T) {
	// Test XOR operation
	assert := assert.New(t)

	flags := AVIO_FLAG_READ | AVIO_FLAG_NONBLOCK
	result := flags ^ AVIO_FLAG_READ
	assert.Equal(AVIO_FLAG_NONBLOCK, result)
}

func Test_AVIOFlag_Bitwise_004(t *testing.T) {
	// Test clearing a flag
	assert := assert.New(t)

	flags := AVIO_FLAG_READ | AVIO_FLAG_WRITE | AVIO_FLAG_NONBLOCK
	flags = flags &^ AVIO_FLAG_WRITE // Clear WRITE
	assert.True(flags.Is(AVIO_FLAG_READ))
	assert.False(flags.Is(AVIO_FLAG_WRITE))
	assert.True(flags.Is(AVIO_FLAG_NONBLOCK))
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - CONSTANTS VALIDATION

func Test_AVIOFlag_Constants_001(t *testing.T) {
	// Verify basic constants are unique
	assert := assert.New(t)

	flags := []AVIOFlag{
		AVIO_FLAG_READ,
		AVIO_FLAG_WRITE,
		AVIO_FLAG_NONBLOCK,
		AVIO_FLAG_DIRECT,
	}

	// Check all flags are unique
	seen := make(map[AVIOFlag]bool)
	for _, flag := range flags {
		assert.False(seen[flag], "Duplicate flag found: %v", flag)
		seen[flag] = true
	}
}

func Test_AVIOFlag_Constants_002(t *testing.T) {
	// Verify NONE is zero
	assert := assert.New(t)
	assert.Equal(AVIOFlag(0), AVIO_FLAG_NONE)
}

func Test_AVIOFlag_Constants_003(t *testing.T) {
	// Verify MIN and MAX are set correctly
	assert := assert.New(t)
	assert.Equal(AVIO_FLAG_READ, AVIO_FLAG_MIN)
	assert.Equal(AVIO_FLAG_DIRECT, AVIO_FLAG_MAX)
}

func Test_AVIOFlag_Constants_004(t *testing.T) {
	// Verify READ_WRITE is combination of READ and WRITE
	assert := assert.New(t)
	expected := AVIO_FLAG_READ | AVIO_FLAG_WRITE
	assert.Equal(expected, AVIO_FLAG_READ_WRITE)
}

////////////////////////////////////////////////////////////////////////////////
// TESTS - COMMON PATTERNS

func Test_AVIOFlag_Pattern_001(t *testing.T) {
	// Test read-only mode
	assert := assert.New(t)

	readFlags := AVIO_FLAG_READ
	assert.True(readFlags.Is(AVIO_FLAG_READ))
	assert.False(readFlags.Is(AVIO_FLAG_WRITE))
	assert.Contains(readFlags.String(), "READ")
}

func Test_AVIOFlag_Pattern_002(t *testing.T) {
	// Test write-only mode
	assert := assert.New(t)

	writeFlags := AVIO_FLAG_WRITE
	assert.True(writeFlags.Is(AVIO_FLAG_WRITE))
	assert.False(writeFlags.Is(AVIO_FLAG_READ))
	assert.Contains(writeFlags.String(), "WRITE")
}

func Test_AVIOFlag_Pattern_003(t *testing.T) {
	// Test read-write mode
	assert := assert.New(t)

	readWriteFlags := AVIO_FLAG_READ_WRITE
	assert.True(readWriteFlags.Is(AVIO_FLAG_READ))
	assert.True(readWriteFlags.Is(AVIO_FLAG_WRITE))
}

func Test_AVIOFlag_Pattern_004(t *testing.T) {
	// Test non-blocking mode
	assert := assert.New(t)

	nonBlockFlags := AVIO_FLAG_READ | AVIO_FLAG_NONBLOCK
	assert.True(nonBlockFlags.Is(AVIO_FLAG_READ))
	assert.True(nonBlockFlags.Is(AVIO_FLAG_NONBLOCK))
	assert.Contains(nonBlockFlags.String(), "NONBLOCK")
}

func Test_AVIOFlag_Pattern_005(t *testing.T) {
	// Test direct mode
	assert := assert.New(t)

	directFlags := AVIO_FLAG_WRITE | AVIO_FLAG_DIRECT
	assert.True(directFlags.Is(AVIO_FLAG_WRITE))
	assert.True(directFlags.Is(AVIO_FLAG_DIRECT))
	assert.Contains(directFlags.String(), "DIRECT")
}
