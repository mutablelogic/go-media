package ffmpeg_test

import (
	"testing"

	"github.com/djthorpe/gopi-media/ffmpeg"
)

////////////////////////////////////////////////////////////////////////////////
// TEST ENUMS

func Test_avutil_000(t *testing.T) {
	t.Log("Test_avutil_000")
}

func Test_avutil_001(t *testing.T) {
	for i := 0; i < 100; i++ {
		dict := ffmpeg.NewAVDictionary()
		dict.Close()
	}
}

func Test_avutil_002(t *testing.T) {
	dict := ffmpeg.NewAVDictionary()
	if dict.Count() != 0 {
		t.Error("Expecting count==0")
	}
	if err := dict.Set("a", "b", ffmpeg.AV_DICT_NONE); err != nil {
		t.Error(err)
	} else if dict.Count() != 1 {
		t.Error("Expecting count==1")
	} else if err := dict.Set("b", "a", ffmpeg.AV_DICT_NONE); err != nil {
		t.Error(err)
	} else if dict.Count() != 2 {
		t.Error("Expecting count==2")
	}
	dict.Close()
}
