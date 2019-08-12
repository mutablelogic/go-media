package ffmpeg_test

import (
	"testing"

	// Frameworks
	ffmpeg "github.com/djthorpe/gopi-media/ffmpeg"
)

////////////////////////////////////////////////////////////////////////////////
// TEST ENUMS

func Test_avformat_000(t *testing.T) {
	t.Log("Test_avformat_000")
}

/*
func Test_avformat_001(t *testing.T) {
	ffmpeg.AVInit()
	if ctx, err := ffmpeg.NewAVIOContext("avutil.go", ffmpeg.AVIO_FLAG_READ); err != nil {
		t.Fatal(err)
	} else if err := ctx.Close(); err != nil {
		t.Fatal(err)
	}
}

*/

func Test_avformat_002(t *testing.T) {
	if ctx := ffmpeg.NewAVFormatContext(); ctx == nil {
		t.Fatal("NewAVFormatContext failed")
	} else {
		ctx.Free()
	}
}

func Test_avformat_003(t *testing.T) {
	if ctx := ffmpeg.NewAVFormatContext(); ctx == nil {
		t.Fatal("NewAVFormatContext failed")
	} else if err := ctx.OpenInput("../etc/sample.mp4", nil); err != nil {
		ctx.Free()
		t.Error(err)
	} else {
		ctx.CloseInput()
	}
}

func Test_avformat_004(t *testing.T) {
	if ctx := ffmpeg.NewAVFormatContext(); ctx == nil {
		t.Fatal("NewAVFormatContext failed")
	} else if err := ctx.OpenInput("../etc/sample.mp4", nil); err != nil {
		ctx.Free()
		t.Error(err)
	} else {
		t.Log(ctx.Metadata())
		ctx.CloseInput()
	}
}

func Test_avformat_005(t *testing.T) {
	if ctx := ffmpeg.NewAVFormatContext(); ctx == nil {
		t.Fatal("NewAVFormatContext failed")
	} else if err := ctx.OpenInput("../etc/sample.mp4", nil); err != nil {
		ctx.Free()
		t.Error(err)
	} else {
		t.Log(ctx.Filename())
		ctx.CloseInput()
	}
}

func Test_avformat_006(t *testing.T) {
	ffmpeg.AVFormatInit()
	ffmpeg.AVFormatInit()
	ffmpeg.AVFormatDeinit()
	ffmpeg.AVFormatDeinit()
}

func Test_avformat_007(t *testing.T) {
	if iformats := ffmpeg.EnumerateInputFormats(); len(iformats) == 0 {
		t.Error("EnumerateInputFormats expected a return value")
	} else {
		for _, iformat := range iformats {
			t.Log(iformat)
		}
	}
}

func Test_avformat_008(t *testing.T) {
	if oformats := ffmpeg.EnumerateOutputFormats(); len(oformats) == 0 {
		t.Error("EnumerateOutputFormats expected a return value")
	} else {
		for _, oformat := range oformats {
			t.Log(oformat)
		}
	}
}
