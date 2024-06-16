package main

import (
	"flag"
	"fmt"
	"log"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func main() {
	out := flag.String("out", "", "output file")
	flag.Parse()

	// Check in and out
	if *out == "" {
		log.Fatal("-out flag must be specified")
	}

	// Allocate the output media context
	ctx, err := ff.AVFormat_create_file(*out, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ff.AVFormat_close_writer(ctx)

	// Add the audio and video streams using the default format codecs and initialize the codecs.
	var video, audio *Stream
	if codec := ctx.Output().VideoCodec(); codec != ff.AV_CODEC_ID_NONE {
		if stream, err := NewStream(ctx, codec); err != nil {
			log.Fatalf("could not add video stream: %v", err)
		} else {
			video = stream
		}
		defer video.Close()
	}
	if codec := ctx.Output().AudioCodec(); codec != ff.AV_CODEC_ID_NONE {
		if stream, err := NewStream(ctx, codec); err != nil {
			log.Fatalf("could not add audio stream: %v", err)
		} else {
			audio = stream
		}
		defer audio.Close()
	}

	// Now that all the parameters are set, we can open the audio
	// and video codecs and allocate the necessary encode buffers.
	if video != nil {
		// TODO: AVDictionary of options
		if err := video.Open(nil); err != nil {
			log.Fatalf("could not open video codec: %v", err)
		}
	}
	if audio != nil {
		// TODO: AVDictionary of options
		if err := audio.Open(nil); err != nil {
			log.Fatalf("could not open audio codec: %v", err)
		}
	}

	fmt.Println(ctx)

	// Dump the output format
	ff.AVFormat_dump_format(ctx, 0, *out)

	// Open the output file, if needed
	if !ctx.Flags().Is(ff.AVFMT_NOFILE) {
		w, err := ff.AVFormat_avio_open(*out, ff.AVIO_FLAG_WRITE)
		if err != nil {
			log.Fatalf("could not open output file: %v", err)
		} else {
			ctx.SetPb(w)
		}
		defer ff.AVFormat_avio_close(w)
	}

	// Write the stream header, if any
	// TODO: AVDictionary of options
	if err := ff.AVFormat_write_header(ctx, nil); err != nil {
		log.Fatalf("could not write header: %v", err)
	}

	// TODO Write data
	encode_audio, encode_video := true, true
	for encode_audio || encode_video {
		// Choose video if both are available, and video is earlier than audio
		if (encode_video && !encode_audio) || (encode_video && ff.AVUtil_compare_ts(video.next_pts, video.Encoder.TimeBase(), audio.next_pts, audio.Encoder.TimeBase()) <= 0) {
			encode_video = !write_video_frame(ctx, video)
		} else {
			encode_audio = !write_audio_frame(ctx, audio)
		}
	}

	// Write the trailer
	if err := ff.AVFormat_write_trailer(ctx); err != nil {
		log.Fatalf("could not write trailer: %v", err)
	}
}

/*
 * Copyright (c) 2003 Fabrice Bellard
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */
