package libheif_test

import (
	"strings"
	"testing"

	. "github.com/mutablelogic/go-media/sys/libheif"
)

func Test_error_000(t *testing.T) {
	err := HeifError{Code: HEIF_ERROR_OK, Subcode: HEIF_SUBERROR_UNSPECIFIED}
	if err.Code != HEIF_ERROR_OK {
		t.Fatalf("HeifError code=%d want=%d", err.Code, HEIF_ERROR_OK)
	}
	if err.Subcode != HEIF_SUBERROR_UNSPECIFIED {
		t.Fatalf("HeifError subcode=%d want=%d", err.Subcode, HEIF_SUBERROR_UNSPECIFIED)
	}
}

func Test_error_001(t *testing.T) {
	err := HeifError{Code: HEIF_ERROR_INVALID_INPUT, Subcode: HEIF_SUBERROR_INVALID_BOX_SIZE}
	msg := err.Error()
	if msg == "" {
		t.Fatal("HeifError.Error returned empty string")
	}
	if !strings.Contains(msg, "code=") {
		t.Fatalf("HeifError fallback message does not include code details: %q", msg)
	}
}
