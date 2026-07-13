//go:build chromaprint

package manager_test

import (
	"path/filepath"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	test "github.com/mutablelogic/go-media/gomedia/test"
)

func TestAudioFingerprint_Input(t *testing.T) {
	m, ctx := test.Begin(t)

	testFile := filepath.Join("..", "..", "etc", "test", "sample.mp3")
	resp, err := m.AudioFingerprint(ctx, schema.AudioFingerprintRequest{Input: testFile})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.Fingerprint == "" {
		t.Fatal("expected non-empty fingerprint")
	}
	if resp.Duration <= 0 {
		t.Fatalf("expected positive duration, got %v", resp.Duration)
	}
}
