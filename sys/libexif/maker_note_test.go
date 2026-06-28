package libexif_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libexif"
)

// loadMakerNote returns the MakerNoteData from the given file or skips.
func loadMakerNote(t *testing.T, path string) *MakerNoteData {
	t.Helper()
	data := Exif_data_new_from_file(path)
	if data == nil {
		t.Skipf("%s: no EXIF data", path)
	}
	t.Cleanup(func() { Exif_data_unref(data) })
	mn := Exif_data_get_maker_note_data(data)
	if mn == nil {
		t.Skipf("%s: no maker note data", path)
	}
	return mn
}

func Test_mnote_000(t *testing.T) {
	mn := loadMakerNote(t, testJPEGMakerNote)
	n := Exif_mnote_data_count(mn)
	if n == 0 {
		t.Fatal("expected non-zero maker note count")
	}
	t.Log("maker note count=", n)
}

func Test_mnote_001(t *testing.T) {
	mn := loadMakerNote(t, testJPEGMakerNote)
	n := Exif_mnote_data_count(mn)
	for i := uint(0); i < n; i++ {
		id := Exif_mnote_data_get_id(mn, i)
		name := Exif_mnote_data_get_name(mn, i)
		title := Exif_mnote_data_get_title(mn, i)
		val := Exif_mnote_data_get_value(mn, i)
		t.Logf("[%d] id=0x%04x name=%q title=%q value=%q", i, id, name, title, val)
	}
}

func Test_mnote_002(t *testing.T) {
	mn := loadMakerNote(t, testJPEGMakerNote)
	n := Exif_mnote_data_count(mn)
	for i := uint(0); i < n; i++ {
		desc := Exif_mnote_data_get_description(mn, i)
		t.Logf("[%d] description=%q", i, desc)
	}
}

func Test_mnote_003(t *testing.T) {
	mn := loadMakerNote(t, testJPEGMakerNote)
	raw := Exif_mnote_data_save(mn)
	if len(raw) == 0 {
		t.Fatal("expected non-empty saved maker note data")
	}
	t.Log("saved bytes=", len(raw))
}

func Test_mnote_004(t *testing.T) {
	// Exercise all JPEG fixtures with maker note data.
	for _, path := range testJPEGs {
		if path == testJPEG {
			continue // sample.jpg has no maker note
		}
		mn := loadMakerNote(t, path)
		n := Exif_mnote_data_count(mn)
		t.Logf("%s: %d maker note entries", path, n)
		for i := uint(0); i < n; i++ {
			name := Exif_mnote_data_get_name(mn, i)
			val := Exif_mnote_data_get_value(mn, i)
			t.Logf("  [%d] %q = %q", i, name, val)
		}
	}
}

func Test_mnote_005(t *testing.T) {
	mn := loadMakerNote(t, testJPEGMakerNote)
	Exif_mnote_data_ref(mn)
	Exif_mnote_data_unref(mn)
}
