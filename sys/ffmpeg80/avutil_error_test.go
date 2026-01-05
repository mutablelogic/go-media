package ffmpeg

import (
	"strings"
	"syscall"
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
// TEST AVError.IsErrno()

func TestAVError_IsErrno_EPERM(t *testing.T) {
	// Create an AVERROR from EPERM (Operation not permitted)
	// Note: FFmpeg errors from errno are typically negative values
	err := AVError(-int(syscall.EPERM))

	if !err.IsErrno(syscall.EPERM) {
		t.Errorf("Expected IsErrno(EPERM) to return true for error code %d", int(err))
	}
}

func TestAVError_IsErrno_ENOENT(t *testing.T) {
	// Create an AVERROR from ENOENT (No such file or directory)
	err := AVError(-int(syscall.ENOENT))

	if !err.IsErrno(syscall.ENOENT) {
		t.Errorf("Expected IsErrno(ENOENT) to return true for error code %d", int(err))
	}

	// Should not match other errors
	if err.IsErrno(syscall.EPERM) {
		t.Error("Expected IsErrno(EPERM) to return false for ENOENT error")
	}
}

func TestAVError_IsErrno_EINVAL(t *testing.T) {
	// Create an AVERROR from EINVAL (Invalid argument)
	err := AVError(-int(syscall.EINVAL))

	if !err.IsErrno(syscall.EINVAL) {
		t.Errorf("Expected IsErrno(EINVAL) to return true for error code %d", int(err))
	}
}

func TestAVError_IsErrno_Zero(t *testing.T) {
	var err AVError

	// Zero error should not match any errno
	if err.IsErrno(syscall.EPERM) {
		t.Error("Expected zero error not to match EPERM")
	}
	if err.IsErrno(syscall.ENOENT) {
		t.Error("Expected zero error not to match ENOENT")
	}
}

func TestAVError_IsErrno_NonErrnoError(t *testing.T) {
	// Test with FFmpeg-specific errors that are not errno-based
	err := AVError(AVERROR_EOF)

	// Should not match system errno values
	if err.IsErrno(syscall.EPERM) {
		t.Error("Expected AVERROR_EOF not to match EPERM")
	}
	if err.IsErrno(syscall.ENOENT) {
		t.Error("Expected AVERROR_EOF not to match ENOENT")
	}
}

func TestAVError_IsErrno_MultipleErrno(t *testing.T) {
	testCases := []struct {
		name  string
		errno syscall.Errno
	}{
		{"EPERM", syscall.EPERM},
		{"ENOENT", syscall.ENOENT},
		{"EINTR", syscall.EINTR},
		{"EIO", syscall.EIO},
		{"ENOMEM", syscall.ENOMEM},
		{"EACCES", syscall.EACCES},
		{"EINVAL", syscall.EINVAL},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := AVError(-int(tc.errno))

			if !err.IsErrno(tc.errno) {
				t.Errorf("Expected IsErrno(%s) to return true for error code %d", tc.name, int(err))
			}

			// Test that it doesn't match other errno values
			for _, other := range testCases {
				if other.errno != tc.errno {
					if err.IsErrno(other.errno) {
						t.Errorf("Expected IsErrno(%s) to return false for %s error", other.name, tc.name)
					}
				}
			}
		})
	}
}
