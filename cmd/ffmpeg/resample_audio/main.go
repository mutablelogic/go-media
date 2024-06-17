package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"log"
	"math"
	"os"
	"unsafe"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
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

	// Source needs to be non-planar
	src_ch_layout := ff.AV_CHANNEL_LAYOUT_MONO
	src_format := ff.AV_SAMPLE_FMT_S16
	src_nb_samples := 1024
	src_rate := 44100

	// Destination needs to be non-planar
	dest_ch_layout := ff.AV_CHANNEL_LAYOUT_MONO
	dest_format := ff.AV_SAMPLE_FMT_U8
	dest_rate := 48000

	if err := ff.SWResample_set_opts(ctx,
		dest_ch_layout, dest_format, dest_rate,
		src_ch_layout, src_format, src_rate,
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

	dest_nb_samples := int(float64(src_nb_samples) * float64(dest_rate) / float64(src_rate))
	max_dest_nb_samples := dest_nb_samples
	dest, err := ff.AVUtil_samples_alloc(dest_nb_samples, dest_ch_layout.NumChannels(), dest_format, false)
	if err != nil {
		log.Fatal(err)
	}

	t := float64(0)
	for t < 10 {
		ff.AVUtil_samples_set_silence(src, 0, src_nb_samples)

		// Generate synthetic audio
		t = fill_samples(src, src_rate, t)

		// Calculate destination number of samples
		dest_nb_samples = int(ff.SWResample_get_delay(ctx, int64(src_rate))) + int(float64(src_nb_samples)*float64(dest_rate)/float64(src_rate))
		if dest_nb_samples > max_dest_nb_samples {
			ff.AVUtil_samples_free(dest)
			dest, err = ff.AVUtil_samples_alloc(dest_nb_samples, dest_ch_layout.NumChannels(), dest_format, true)
			if err != nil {
				log.Fatal(err)
			}
			max_dest_nb_samples = dest_nb_samples
		}

		// convert to destination format
		n, err := ff.SWResample_convert(ctx, dest, dest_nb_samples, src, src_nb_samples)
		if err != nil {
			log.Fatal(err)
		}

		// Calulate the number of samples converted
		_, dest_planesize, err := ff.AVUtil_samples_get_buffer_size(n, dest_ch_layout.NumChannels(), dest_format, true)
		if err != nil {
			log.Fatal(err)
		}

		// We only write the first plane - non-planar format
		if _, err := w.Write(dest.Bytes(0)[:dest_planesize]); err != nil {
			log.Fatal(err)
		}
	}

	ff.AVUtil_samples_free(dest)

	if fmt, err := get_format_from_sample_fmt(dest_format); err != nil {
		log.Fatal(err)
	} else if desc, err := ff.AVUtil_channel_layout_describe(&dest_ch_layout); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Resampling succeeded. Play the output file with the command:")
		log.Printf("  ffplay -f %s -channel_layout %s -ar %d %s\n", fmt, desc, dest_rate, *out)
	}
}

/**
 * Fill buffer with nb_samples, generated starting from time t.
 */
func fill_samples(data *ff.AVSamples, sample_rate int, t float64) float64 {
	tincr := 1.0 / float64(sample_rate)
	buf := data.Int16(0)
	c := 2.0 * math.Pi * 440.0

	// Generate sin tone with 440Hz frequency and duplicated channels
	for i := 0; i < data.NumSamples(); i += data.NumChannels() {
		sample := math.Sin(c * t)
		for j := 0; j < data.NumChannels(); j++ {
			buf[i+j] = int16(sample * 0.5 * math.MaxInt16)
		}
		t = t + tincr
	}
	return t
}

// NativeEndian is the ByteOrder of the current system.
var NativeEndian binary.ByteOrder

func init() {
	// Examine the memory layout of an int16 to determine system
	// endianness.
	var one int16 = 1
	b := (*byte)(unsafe.Pointer(&one))
	if *b == 0 {
		NativeEndian = binary.BigEndian
	} else {
		NativeEndian = binary.LittleEndian
	}
}

func get_format_from_sample_fmt(sample_fmt ff.AVSampleFormat) (string, error) {
	type sample_fmt_entry struct {
		sample_fmt ff.AVSampleFormat
		fmt_be     string
		fmt_le     string
	}
	sample_fmt_entries := []sample_fmt_entry{
		{ff.AV_SAMPLE_FMT_U8, "u8", "u8"},
		{ff.AV_SAMPLE_FMT_S16, "s16be", "s16le"},
		{ff.AV_SAMPLE_FMT_S32, "s32be", "s32le"},
		{ff.AV_SAMPLE_FMT_FLT, "f32be", "f32le"},
		{ff.AV_SAMPLE_FMT_DBL, "f64be", "f64le"},
	}

	for _, entry := range sample_fmt_entries {
		if sample_fmt == entry.sample_fmt {
			if NativeEndian == binary.LittleEndian {
				return entry.fmt_le, nil
			} else {
				return entry.fmt_be, nil
			}
		}
	}
	return "", errors.New("sample format is not supported as output format")
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
