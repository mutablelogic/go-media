package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

const (
	INBUF_SIZE = 4096
)

func main() {
	in := flag.String("in", "", "input file")
	codec_name := flag.String("codec", "mpeg", "input codec to use")
	out := flag.String("out", "", "output file")
	flag.Parse()

	// Check in and out
	if *in == "" || *out == "" {
		log.Fatal("-in and -out files must be specified")
	}

	// Find video decoder
	codec := ff.AVCodec_find_decoder_by_name(*codec_name)
	if codec == nil {
		log.Fatal("Codec not found")
	}
	parser := ff.AVCodec_parser_init(codec.ID())
	if parser == nil {
		log.Fatal("Parser not found")
	}
	defer ff.AVCodec_parser_close(parser)

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		log.Fatal("Could not allocate audio codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// For some codecs, such as msmpeg4 and mpeg4, width and height MUST be initialized
	// there because this information is not available in the bitstream.

	// open it
	if err := ff.AVCodec_open(ctx, codec, nil); err != nil {
		log.Fatal(err)
	}

	// Read file
	r, err := os.Open(*in)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// Decode packets and frames
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		log.Fatal("Could not allocate video frame")
	}
	defer ff.AVUtil_frame_free(frame)

	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		log.Fatal("Could not allocate packet")
	}
	defer ff.AVCodec_packet_free(packet)

	inbuf := make([]byte, INBUF_SIZE+ff.AV_INPUT_BUFFER_PADDING_SIZE)
	data := inbuf
FOR_LOOP:
	for {
		var eof bool
		data_size, err := r.Read(inbuf)
		if err == io.EOF || data_size == 0 {
			eof = true
		} else if err != nil {
			log.Fatal(err)
		}

		// Use the parser to split the data into frames
		data = inbuf[:data_size]
		for data_size > 0 {
			fmt.Println("parsing input data ", data_size)
			size := ff.AVCodec_parser_parse(parser, ctx, packet, data, ff.AV_NOPTS_VALUE, ff.AV_NOPTS_VALUE, 0)
			if size < 0 {
				log.Fatal("Error while parsing")
			} else {
				fmt.Println("  parsed data ", size)
				fmt.Println("  packet ", packet.Size())
			}

			// Adjust the input buffer beyond the parsed data
			data = data[size:]
			data_size -= size

			// Decode the packet
			if packet.Size() > 0 {
				if err := decode(ctx, frame, packet, *out); err != nil {
					log.Fatal(err)
				}
			}
		}
		if eof {
			break FOR_LOOP
		}
	}
}

func decode(ctx *ff.AVCodecContext, frame *ff.AVFrame, packet *ff.AVPacket, out string) error {
	if err := ff.AVCodec_send_packet(ctx, packet); err != nil {
		return err
	}

	for {
		if err := ff.AVCodec_receive_frame(ctx, frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}

		// The picture is allocated by the decoder. no need to free it
		filename := filepath.Join(out, fmt.Sprintf("frame-%d.pgm", ctx.FrameNum()))
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()

		// Save frame
		if err := pgm_save(w, frame); err != nil {
			return err
		}
	}
}

func pgm_save(w io.Writer, frame *ff.AVFrame) error {
	width := frame.Width()
	height := frame.Height()

	// Write the header
	if _, err := fmt.Fprintf(w, "P5\n%d %d\n%d\n", width, height, 255); err != nil {
		return err
	}

	// Write the data
	stride := frame.Linesize(0)
	log.Print("stride ", stride)
	log.Print("width ", width)
	log.Print("height ", height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pix := frame.Uint8(0)[y*stride+x]
			if _, err := w.Write([]byte{pix}); err != nil {
				return err
			}
		}
	}
	return nil
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
