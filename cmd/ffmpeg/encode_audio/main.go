package main

import (
	"flag"
	"log"
	"slices"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func main() {
	out := flag.String("out", "", "output file")
	flag.Parse()

	// Check out and size
	if *out == "" {
		log.Fatal("-out argument must be specified")
	}

	// find the MP2 encoder
	codec := ff.AVCodec_find_encoder(ff.AV_CODEC_ID_MP2)
	if codec == nil {
		log.Fatal("Codec not found")
	}

	// Allocate a codec
	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		log.Fatal("Could not allocate audio codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Set codec parameters
	ctx.SetBitRate(64000)
	ctx.SetSampleFormat(ff.AV_SAMPLE_FMT_S16)
	if !check_sample_fmt(codec, ctx.SampleFormat()) {
		log.Fatalf("Encoder does not support sample format %v", ctx.SampleFormat())
	}

	// select other audio parameters supported by the encoder
	//	ctx.SetSampleRate(select_sample_rate(codec))
}

// check that a given sample format is supported by the encoder
func check_sample_fmt(codec *ff.AVCodec, sample_fmt ff.AVSampleFormat) bool {
	return slices.Contains(codec.SampleFormats(), sample_fmt)
}

// just pick the highest supported samplerate
/*
func select_sample_rate(codec *AVCodec) {
    const int *p;
    int best_samplerate = 0;

    if (!codec->supported_samplerates)
        return 44100;

    p = codec->supported_samplerates;
    while (*p) {
        if (!best_samplerate || abs(44100 - *p) < abs(44100 - best_samplerate))
            best_samplerate = *p;
        p++;
    }
    return best_samplerate;
}
*/
