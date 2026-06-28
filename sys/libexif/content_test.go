package libexif_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libexif"
)

func Test_content_000(t *testing.T) {
	content := Exif_content_new()
	if content == nil {
		t.Fatal("Exif_content_new returned nil")
	}
	defer Exif_content_unref(content)
}

func Test_content_001(t *testing.T) {
	content := Exif_content_new()
	if content == nil {
		t.Fatal("Exif_content_new returned nil")
	}
	Exif_content_ref(content)
	Exif_content_unref(content)
	Exif_content_unref(content)
}

func Test_content_002(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	for ifd := EXIF_IFD_0; ifd < EXIF_IFD_COUNT; ifd++ {
		content := Exif_data_get_content(data, ifd)
		if content == nil {
			continue
		}
		got := Exif_content_get_ifd(content)
		if got != ifd {
			t.Errorf("IFD %v: get_ifd returned %v", ifd, got)
		}
		t.Logf("IFD %s: %d entries", Exif_ifd_get_name(ifd), Exif_content_get_entry_count(content))
	}
}

func Test_content_003(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	for ifd := EXIF_IFD_0; ifd < EXIF_IFD_COUNT; ifd++ {
		content := Exif_data_get_content(data, ifd)
		if content == nil {
			continue
		}
		n := Exif_content_get_entry_count(content)
		for i := uint(0); i < n; i++ {
			entry := Exif_content_get_entry_at(content, i)
			if entry == nil {
				t.Errorf("IFD %v entry %d: nil", ifd, i)
			}
		}
	}
}

func Test_content_004(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	content := Exif_data_get_content(data, EXIF_IFD_0)
	if content == nil {
		t.Skip("no IFD0 content in test JPEG")
	}

	entry := Exif_content_get_entry(content, EXIF_TAG_ORIENTATION)
	if entry == nil {
		t.Skip("no Orientation tag in IFD0")
	}
	t.Log("Orientation=", Exif_entry_get_value(entry))
}

func Test_content_005(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	var visited []Tag
	for ifd := EXIF_IFD_0; ifd < EXIF_IFD_COUNT; ifd++ {
		content := Exif_data_get_content(data, ifd)
		if content == nil {
			continue
		}
		Exif_content_foreach_entry(content, func(entry *Entry) {
			visited = append(visited, Exif_entry_get_tag(entry))
		})
	}
	if len(visited) == 0 {
		t.Fatal("foreach_entry visited no entries")
	}
	t.Log("visited tags=", len(visited))
}

func Test_content_006(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	content := Exif_data_get_content(data, EXIF_IFD_0)
	if content == nil {
		t.Skip("no IFD0 content in test JPEG")
	}
	Exif_content_fix(content)
}

func Test_content_007(t *testing.T) {
	data := Exif_data_new_from_file(testJPEG)
	if data == nil {
		t.Fatal("Exif_data_new_from_file returned nil")
	}
	defer Exif_data_unref(data)

	content := Exif_data_get_content(data, EXIF_IFD_0)
	if content == nil {
		t.Skip("no IFD0 content in test JPEG")
	}
	Exif_content_dump(content, 0)
}
