package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
	// Packages
)

func main() {
	out := flag.String("out", "", "output file")
	flag.Parse()

	// Check flags
	if *out == "" {
		log.Fatal("-out flag must be specified")
	}

	// Create a resampler context
	ctx := ff.SWResample_alloc()
	if ctx == nil {
		log.Fatal("could not allocate resampler context")
	}
	defer ff.SWResample_free(ctx)

	// Set common parameters for resampling
	src_ch_layout := ff.AV_CHANNEL_LAYOUT_STEREO
	src_format := ff.AV_SAMPLE_FMT_DBL
	src_nb_samples := 1024
	src_rate := 44100
	dest_ch_layout := ff.AV_CHANNEL_LAYOUT_SURROUND
	dest_format := ff.AV_SAMPLE_FMT_S16
	dest_rate := 48000

	if err := ff.SWResample_set_opts(ctx,
		src_ch_layout, src_format, src_rate,
		dest_ch_layout, dest_format, dest_rate,
	); err != nil {
		log.Fatal(err)
	}

	// initialize the resampling context
	if err := ff.SWResample_init(ctx); err != nil {
		log.Fatal(err)
	}

	// Allocate source and destination samples buffers
	src, err := ff.AVUtil_samples_alloc(src_nb_samples, src_ch_layout.NumChannels(), src_format, false)
	if err != nil {
		log.Fatal(err)
	}
	defer ff.AVUtil_samples_free(src)

	// Open destination file
	w, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	dest_nb_samples := src_nb_samples * dest_rate / src_rate
	max_dest_nb_samples := dest_nb_samples
	dest, err := ff.AVUtil_samples_alloc(dest_nb_samples, dest_ch_layout.NumChannels(), dest_format, false)
	if err != nil {
		log.Fatal(err)
	}

	for t := 0; t < 10; t++ {
		// Generate input data
		// TODO
		ff.AVUtil_samples_set_silence(src, 0, src_nb_samples)

		// Calculate destination number of samples
		dest_nb_samples = int(ff.SWResample_get_delay(ctx, int64(src_rate))) + src_nb_samples*dest_rate/src_rate
		if dest_nb_samples > max_dest_nb_samples {
			ff.AVUtil_samples_free(dest)
			dest, err = ff.AVUtil_samples_alloc(dest_nb_samples, dest_ch_layout.NumChannels(), dest_format, true)
			if err != nil {
				log.Fatal(err)
			}
			max_dest_nb_samples = dest_nb_samples
		}

		// convert to destination format
		n, err := ff.SWResample_convert(ctx, dest, src, dest_nb_samples, src_nb_samples)
		if err != nil {
			log.Fatal(err)
		}

		// Calulate the number of samples converted
		dest_bufsize, _, err := ff.AVUtil_samples_get_buffer_size(n, dest_ch_layout.NumChannels(), dest_format, true)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("TODO: write samples to file (buffer size: ", dest_bufsize, ")")
		for i := 0; i < dest.NumPlanes(); i++ {
			fmt.Println("  Plane", i, ":", len(dest.Bytes(i)))
		}
	}

	ff.AVUtil_samples_free(dest)
}

/*
 * Copyright (c) 2012 Stefano Sabatini
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
