package main

import (
	"flag"
	"log"
	"os"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

const (
	SRC_WIDTH    = 320
	SRC_HEIGHT   = 240
	SRC_PIX_FMT  = ff.AV_PIX_FMT_YUV420P
	DEST_PIX_FMT = ff.AV_PIX_FMT_RGB24
)

func main() {
	out := flag.String("out", "", "output file")
	size := flag.String("size", "320x240", "output frame size")
	flag.Parse()

	// Check out and size
	if *out == "" {
		log.Fatal("-out argument must be specified")
	}
	width, height, err := ff.AVUtil_parse_video_size(*size)
	if err != nil {
		log.Fatal(err)
	}

	// Create destination
	dest, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer dest.Close()

	// Create scaling context
	ctx := ff.SWScale_get_context(SRC_WIDTH, SRC_HEIGHT, SRC_PIX_FMT, width, height, DEST_PIX_FMT, ff.SWS_BILINEAR, nil, nil, nil)
	if ctx == nil {
		log.Fatal("failed to allocate swscale context")
	}
	defer ff.SWScale_free_context(ctx)

	// Allocate source and destination image buffers
	src_data, src_stride, _, err := ff.AVUtil_image_alloc(SRC_WIDTH, SRC_HEIGHT, SRC_PIX_FMT, 16)
	if err != nil {
		log.Fatal(err)
	}
	defer ff.AVUtil_image_free(src_data)

	dest_data, dest_stride, dest_bufsize, err := ff.AVUtil_image_alloc(width, height, DEST_PIX_FMT, 1)
	if err != nil {
		log.Fatal(err)
	}
	defer ff.AVUtil_image_free(dest_data)

	for i := 0; i < 1000; i++ {
		// Generate synthetic video
		fill_yuv_image(src_data, src_stride, SRC_WIDTH, SRC_HEIGHT, i)

		// Convert to destination format
		// TODO: Currently getting bad src image pointers here
		ff.SWScale_scale(ctx, src_data, src_stride, 0, SRC_HEIGHT, dest_data, dest_stride)

		// Write scaled image to file
		if _, err := dest.Write(ff.AVUtil_image_bytes(dest_data, dest_bufsize)); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Scaling succeeded. Play the output file with the command:\n  ffplay -f rawvideo -pixel_format %s -video_size %dx%d %s\n", ff.AVUtil_get_pix_fmt_name(DEST_PIX_FMT), width, height, *out)
}

func fill_yuv_image(data [][]byte, stride []int, width, height int, frame_index int) {
	/* Y */
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			data[0][y*stride[0]+x] = byte(x + y + frame_index*3)
		}
	}

	/* Cb and Cr */
	for y := 0; y < height>>1; y++ {
		for x := 0; x < width>>1; x++ {
			data[1][y*stride[1]+x] = byte(128 + y + frame_index*2)
			data[2][y*stride[2]+x] = byte(64 + x + frame_index*5)
		}
	}
}

/*
 * Copyright (c) 2001 Fabrice Bellard
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
