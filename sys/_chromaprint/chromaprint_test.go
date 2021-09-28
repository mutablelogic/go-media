package chromaprint_test

import (
	"testing"

	// Namespace imports
	. "github.com/djthorpe/go-media/sys/chromaprint"
)

////////////////////////////////////////////////////////////////////////////////
// TEST ENUMS

func Test_Chromaprint_000(t *testing.T) {
	t.Log("Test_chromaprint_000")
}

func Test_Chromaprint_001(t *testing.T) {
	if version := Version(); version != "" {
		t.Log("Version:", version)
	} else {
		t.Error("Expected non-empty version")
	}
}

func Test_Chromaprint_002(t *testing.T) {
	if ctx := NewChromaprint(ALGORITHM_DEFAULT); ctx == nil {
		t.Error("Unexpected nil return from NewChromaprint")
	} else {
		ctx.Free()
	}
}

func Test_Chromaprint_004(t *testing.T) {
	if ctx := NewChromaprint(ALGORITHM_DEFAULT); ctx == nil {
		t.Error("Unexpected nil return from NewChromaprint")
	} else if err := ctx.Start(44100, 2); err != nil {
		t.Error(err)
	} else {
		t.Log(ctx)
		ctx.Free()
	}
}

func Test_Chromaprint_005(t *testing.T) {
	ctx := NewChromaprint(ALGORITHM_DEFAULT)
	if ctx == nil {
		t.Error("Unexpected nil return from NewChromaprint")
	}
	defer ctx.Free()
	rate := 44100
	ch := 1
	size := rate * 5 * ch // Int16 samples for 5 seconds
	if err := ctx.Start(44100, 2); err != nil {
		t.Error(err)
	}
	buf := make([]byte, size*2)
	for i := 0; i < 5; i++ {
		t.Log("Feeding 5 seconds of silence...")
		if err := ctx.Write(buf, size); err != nil {
			t.Error(err)
		}
	}
	if err := ctx.Finish(); err != nil {
		t.Error(err)
	}
	if fp, err := ctx.GetFingerprint(); err != nil {
		t.Error(err)
	} else {
		t.Log("Ctx=", ctx)
		t.Log("Fingerprint=", fp)
	}

}
