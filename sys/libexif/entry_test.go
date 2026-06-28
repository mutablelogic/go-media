package libexif_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libexif"
)

// firstEntry returns the first Entry found in any IFD of the loaded Data,
// or nil if none exist.
func firstEntry(t *testing.T) *Entry {
	t.Helper()
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	t.Cleanup(func() { Exif_data_unref(data) })

	for ifd := EXIF_IFD_0; ifd < EXIF_IFD_COUNT; ifd++ {
		content := Exif_data_get_content(data, ifd)
		if content == nil {
			continue
		}
		if Exif_content_get_entry_count(content) > 0 {
			return Exif_content_get_entry_at(content, 0)
		}
	}
	return nil
}

func Test_entry_000(t *testing.T) {
	entry := Exif_entry_new()
	if entry == nil {
		t.Fatal("Exif_entry_new returned nil")
	}
	defer Exif_entry_unref(entry)
}

func Test_entry_001(t *testing.T) {
	entry := Exif_entry_new()
	if entry == nil {
		t.Fatal("Exif_entry_new returned nil")
	}
	Exif_entry_ref(entry)
	Exif_entry_unref(entry)
	Exif_entry_unref(entry)
}

func Test_entry_002(t *testing.T) {
	entry := firstEntry(t)
	if entry == nil {
		t.Fatal("no entries found in test JPEG")
	}

	tag := Exif_entry_get_tag(entry)
	ifd := Exif_entry_get_ifd(entry)
	name := Exif_tag_get_name_in_ifd(tag, ifd)
	t.Logf("tag=0x%04x name=%q", tag, name)
	t.Logf("format=%q size=%d components=%d",
		Exif_format_get_name(Exif_entry_get_format(entry)),
		Exif_entry_get_size(entry),
		Exif_entry_get_components(entry))
}

func Test_entry_003(t *testing.T) {
	entry := firstEntry(t)
	if entry == nil {
		t.Fatal("no entries found in test JPEG")
	}

	data := Exif_entry_get_data(entry)
	if len(data) == 0 {
		t.Fatal("expected non-empty data")
	}
	t.Log("raw data=", data)
}

func Test_entry_004(t *testing.T) {
	entry := firstEntry(t)
	if entry == nil {
		t.Fatal("no entries found in test JPEG")
	}

	val := Exif_entry_get_value(entry)
	if val == "" {
		t.Fatal("expected non-empty value string")
	}
	t.Log("value=", val)
}

func Test_entry_005(t *testing.T) {
	entry := firstEntry(t)
	if entry == nil {
		t.Fatal("no entries found in test JPEG")
	}

	ifd := Exif_entry_get_ifd(entry)
	if ifd == EXIF_IFD_COUNT {
		t.Fatal("expected valid IFD, got EXIF_IFD_COUNT")
	}
	t.Log("ifd=", Exif_ifd_get_name(ifd))
}

func Test_entry_006(t *testing.T) {
	entry := firstEntry(t)
	if entry == nil {
		t.Fatal("no entries found in test JPEG")
	}
	Exif_entry_fix(entry)
}

func Test_entry_007(t *testing.T) {
	entry := firstEntry(t)
	if entry == nil {
		t.Fatal("no entries found in test JPEG")
	}
	Exif_entry_dump(entry, 0)
}

func Test_entry_008(t *testing.T) {
	// Walk all IFDs and log every entry in the sample JPEG.
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	total := 0
	for ifd := EXIF_IFD_0; ifd < EXIF_IFD_COUNT; ifd++ {
		content := Exif_data_get_content(data, ifd)
		if content == nil {
			continue
		}
		n := Exif_content_get_entry_count(content)
		for i := uint(0); i < n; i++ {
			entry := Exif_content_get_entry_at(content, i)
			if entry == nil {
				continue
			}
			tag := Exif_entry_get_tag(entry)
			val := Exif_entry_get_value(entry)
			t.Logf("IFD %s tag=0x%04x %q = %q",
				Exif_ifd_get_name(ifd), tag,
				Exif_tag_get_name_in_ifd(tag, ifd), val)
			total++
		}
	}
	if total == 0 {
		t.Fatal("no entries found in any IFD")
	}
	t.Log("total entries=", total)
}
