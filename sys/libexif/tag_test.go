package libexif_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libexif"
)

func Test_ifd_000(t *testing.T) {
	for ifd := EXIF_IFD_0; ifd < EXIF_IFD_COUNT; ifd++ {
		name := Exif_ifd_get_name(ifd)
		if name == "" {
			t.Errorf("IFD %v: empty name", ifd)
		}
		t.Logf("IFD %v: name=%q", ifd, name)
	}
}

func Test_tag_000(t *testing.T) {
	n := Exif_tag_table_count()
	if n == 0 {
		t.Fatal("expected non-zero tag table count")
	}
	t.Log("tag table count=", n)
}

func Test_tag_001(t *testing.T) {
	n := Exif_tag_table_count()
	for i := uint(0); i < n; i++ {
		name := Exif_tag_table_get_name(i)
		if name == "" {
			break // sentinel at end of table
		}
		tag := Exif_tag_table_get_tag(i)
		t.Logf("entry %d: tag=0x%04x name=%q", i, tag, name)
	}
}

func Test_tag_002(t *testing.T) {
	tag := Exif_tag_from_name("Orientation")
	if tag != EXIF_TAG_ORIENTATION {
		t.Errorf("got tag 0x%04x, want 0x%04x", tag, EXIF_TAG_ORIENTATION)
	}
}

func Test_tag_003(t *testing.T) {
	for _, ifd := range []IFD{EXIF_IFD_0, EXIF_IFD_EXIF} {
		name := Exif_tag_get_name_in_ifd(EXIF_TAG_ORIENTATION, ifd)
		title := Exif_tag_get_title_in_ifd(EXIF_TAG_ORIENTATION, ifd)
		t.Logf("IFD %v: name=%q title=%q", ifd, name, title)
	}
	// Orientation is defined in IFD0
	name := Exif_tag_get_name_in_ifd(EXIF_TAG_ORIENTATION, EXIF_IFD_0)
	if name == "" {
		t.Fatal("expected non-empty name for ORIENTATION in IFD0")
	}
}

func Test_tag_004(t *testing.T) {
	desc := Exif_tag_get_description_in_ifd(EXIF_TAG_ORIENTATION, EXIF_IFD_0)
	if desc == "" {
		t.Fatal("expected non-empty description for ORIENTATION in IFD0")
	}
	t.Log("description=", desc)
}

func Test_tag_005(t *testing.T) {
	level := Exif_tag_get_support_level_in_ifd(EXIF_TAG_ORIENTATION, EXIF_IFD_0, EXIF_DATA_TYPE_COMPRESSED)
	t.Logf("support level for ORIENTATION in IFD0 (compressed)= %v", level)

	level = Exif_tag_get_support_level_in_ifd(EXIF_TAG_ORIENTATION, EXIF_IFD_GPS, EXIF_DATA_TYPE_COMPRESSED)
	if level != EXIF_SUPPORT_LEVEL_NOT_RECORDED {
		t.Errorf("expected NOT_RECORDED for ORIENTATION in GPS IFD, got %v", level)
	}
}
