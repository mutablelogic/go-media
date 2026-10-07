package ffmpeg

import (
	"strings"
	"testing"
)

////////////////////////////////////////////////////////////////////////////////
// TEST AVError.Error()

func TestAVError_Error_Zero(t *testing.T) {
	var err AVError
	result := err.Error()
	if result != "" {
		t.Errorf("Expected empty string for zero error, got: %q", result)
	}
}

func TestAVError_Error_EOF(t *testing.T) {
	err := AVError(AVERROR_EOF)
	result := err.Error()
	if result == "" {
		t.Error("Expected non-empty error message for AVERROR_EOF")
	}
	// The actual message may vary depending on FFmpeg version,
	// but it should contain something meaningful
	t.Logf("AVERROR_EOF message: %q", result)
}

func TestAVError_Error_InvalidData(t *testing.T) {
	err := AVError(AVERROR_INVALIDDATA)
	result := err.Error()
	if result == "" {
		t.Error("Expected non-empty error message for AVERROR_INVALIDDATA")
	}
	t.Logf("AVERROR_INVALIDDATA message: %q", result)
}

func TestAVError_Error_DecoderNotFound(t *testing.T) {
	err := AVError(AVERROR_DECODER_NOT_FOUND)
	result := err.Error()
	if result == "" {
		t.Error("Expected non-empty error message for AVERROR_DECODER_NOT_FOUND")
	}
	// Should contain something about decoder
	if !strings.Contains(strings.ToLower(result), "decoder") &&
		!strings.Contains(strings.ToLower(result), "not found") {
		t.Logf("Warning: expected 'decoder' or 'not found' in message, got: %q", result)
	}
	t.Logf("AVERROR_DECODER_NOT_FOUND message: %q", result)
}

func TestAVError_Error_EncoderNotFound(t *testing.T) {
	err := AVError(AVERROR_ENCODER_NOT_FOUND)
	result := err.Error()
	if result == "" {
		t.Error("Expected non-empty error message for AVERROR_ENCODER_NOT_FOUND")
	}
	t.Logf("AVERROR_ENCODER_NOT_FOUND message: %q", result)
}

func TestAVError_Error_BufferTooSmall(t *testing.T) {
	err := AVError(AVERROR_BUFFER_TOO_SMALL)
	result := err.Error()
	if result == "" {
		t.Error("Expected non-empty error message for AVERROR_BUFFER_TOO_SMALL")
	}
	t.Logf("AVERROR_BUFFER_TOO_SMALL message: %q", result)
}

func TestAVError_Error_AllConstants(t *testing.T) {
	// Test all error constants to ensure they return non-empty messages
	errors := []struct {
		name string
		code AVError
	}{
		{"AVERROR_BSF_NOT_FOUND", AVError(AVERROR_BSF_NOT_FOUND)},
		{"AVERROR_BUG", AVError(AVERROR_BUG)},
		{"AVERROR_BUFFER_TOO_SMALL", AVError(AVERROR_BUFFER_TOO_SMALL)},
		{"AVERROR_DECODER_NOT_FOUND", AVError(AVERROR_DECODER_NOT_FOUND)},
		{"AVERROR_DEMUXER_NOT_FOUND", AVError(AVERROR_DEMUXER_NOT_FOUND)},
		{"AVERROR_ENCODER_NOT_FOUND", AVError(AVERROR_ENCODER_NOT_FOUND)},
		{"AVERROR_EOF", AVError(AVERROR_EOF)},
		{"AVERROR_EXIT", AVError(AVERROR_EXIT)},
		{"AVERROR_EXTERNAL", AVError(AVERROR_EXTERNAL)},
		{"AVERROR_FILTER_NOT_FOUND", AVError(AVERROR_FILTER_NOT_FOUND)},
		{"AVERROR_INVALIDDATA", AVError(AVERROR_INVALIDDATA)},
		{"AVERROR_MUXER_NOT_FOUND", AVError(AVERROR_MUXER_NOT_FOUND)},
		{"AVERROR_OPTION_NOT_FOUND", AVError(AVERROR_OPTION_NOT_FOUND)},
		{"AVERROR_PATCHWELCOME", AVError(AVERROR_PATCHWELCOME)},
		{"AVERROR_PROTOCOL_NOT_FOUND", AVError(AVERROR_PROTOCOL_NOT_FOUND)},
		{"AVERROR_STREAM_NOT_FOUND", AVError(AVERROR_STREAM_NOT_FOUND)},
		{"AVERROR_BUG2", AVError(AVERROR_BUG2)},
		{"AVERROR_UNKNOWN", AVError(AVERROR_UNKNOWN)},
		{"AVERROR_EXPERIMENTAL", AVError(AVERROR_EXPERIMENTAL)},
		{"AVERROR_INPUT_CHANGED", AVError(AVERROR_INPUT_CHANGED)},
		{"AVERROR_OUTPUT_CHANGED", AVError(AVERROR_OUTPUT_CHANGED)},
		{"AVERROR_HTTP_BAD_REQUEST", AVError(AVERROR_HTTP_BAD_REQUEST)},
		{"AVERROR_HTTP_UNAUTHORIZED", AVError(AVERROR_HTTP_UNAUTHORIZED)},
		{"AVERROR_HTTP_FORBIDDEN", AVError(AVERROR_HTTP_FORBIDDEN)},
		{"AVERROR_HTTP_NOT_FOUND", AVError(AVERROR_HTTP_NOT_FOUND)},
		{"AVERROR_HTTP_OTHER_4XX", AVError(AVERROR_HTTP_OTHER_4XX)},
		{"AVERROR_HTTP_SERVER_ERROR", AVError(AVERROR_HTTP_SERVER_ERROR)},
	}

	for _, tc := range errors {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.code.Error()
			if result == "" {
				t.Errorf("%s returned empty error message", tc.name)
			}
			t.Logf("%s: %q (code: %d)", tc.name, result, int(tc.code))
		})
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST AVError.IsEAGAIN / IsEINVAL / IsENOSYS
//
// Constructed from testErrnoEAGAIN/testErrnoEINVAL/testErrnoENOSYS (the C
// compiler's own <errno.h> values, exposed by avutil_error.go), not
// syscall.EAGAIN etc - on Windows those are synthetic Go-invented values
// with no relation to the real errno.h constants mingw-compiled FFmpeg
// actually returns, so comparing against them would be tautological and
// never catch a real mismatch between the two.

func TestAVError_IsEAGAIN(t *testing.T) {
	err := AVError(-testErrnoEAGAIN)

	if !err.IsEAGAIN() {
		t.Errorf("Expected IsEAGAIN() to return true for error code %d", int(err))
	}
	if err.IsEINVAL() {
		t.Error("Expected IsEINVAL() to return false for an EAGAIN error")
	}
	if err.IsENOSYS() {
		t.Error("Expected IsENOSYS() to return false for an EAGAIN error")
	}
}

func TestAVError_IsEINVAL(t *testing.T) {
	err := AVError(-testErrnoEINVAL)

	if !err.IsEINVAL() {
		t.Errorf("Expected IsEINVAL() to return true for error code %d", int(err))
	}
	if err.IsEAGAIN() {
		t.Error("Expected IsEAGAIN() to return false for an EINVAL error")
	}
}

func TestAVError_IsENOSYS(t *testing.T) {
	err := AVError(-testErrnoENOSYS)

	if !err.IsENOSYS() {
		t.Errorf("Expected IsENOSYS() to return true for error code %d", int(err))
	}
	if err.IsEAGAIN() {
		t.Error("Expected IsEAGAIN() to return false for an ENOSYS error")
	}
}

func TestAVError_IsEAGAIN_Zero(t *testing.T) {
	var err AVError

	if err.IsEAGAIN() {
		t.Error("Expected zero error not to match EAGAIN")
	}
	if err.IsEINVAL() {
		t.Error("Expected zero error not to match EINVAL")
	}
	if err.IsENOSYS() {
		t.Error("Expected zero error not to match ENOSYS")
	}
}

func TestAVError_IsEAGAIN_NonErrnoError(t *testing.T) {
	// Test with an FFmpeg-specific error that is not errno-based
	err := AVError(AVERROR_EOF)

	if err.IsEAGAIN() {
		t.Error("Expected AVERROR_EOF not to match EAGAIN")
	}
	if err.IsEINVAL() {
		t.Error("Expected AVERROR_EOF not to match EINVAL")
	}
	if err.IsENOSYS() {
		t.Error("Expected AVERROR_EOF not to match ENOSYS")
	}
}
