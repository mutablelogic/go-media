package libheif_test

import (
	"testing"

	. "github.com/mutablelogic/go-media/sys/libheif"
)

func Test_library_000(t *testing.T) {
	v := Libheif_get_version()
	if v == "" {
		t.Fatal("Libheif_get_version returned empty string")
	}
	if n := Libheif_get_version_number(); n == 0 {
		t.Fatal("Libheif_get_version_number returned 0")
	}
	if m := Libheif_get_version_number_major(); m <= 0 {
		t.Fatalf("Libheif_get_version_number_major=%d", m)
	}
	minor := Libheif_get_version_number_minor()
	maint := Libheif_get_version_number_maintenance()

	numeric := Libheif_get_version_number()
	recomposed := (uint32(Libheif_get_version_number_major()) << 24) | (uint32(minor) << 16) | (uint32(maint) << 8)
	if numeric != recomposed {
		t.Fatalf("version number mismatch: got=0x%x recomposed=0x%x", numeric, recomposed)
	}

	dir := Libheif_get_plugin_directory()
	if dir == "" {
		t.Fatal("Libheif_get_plugin_directory returned empty string")
	}
}

func Test_library_001(t *testing.T) {
	if err := Libheif_init(); err != nil {
		t.Fatalf("Libheif_init error=%v", err)
	}
	Libheif_deinit()
}
