package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

const (
	AUDIO_INBUF_SIZE    = 20480
	AUDIO_REFILL_THRESH = 4096
)

func main() {
	in := flag.String("in", "", "input file")
	codec_name := flag.String("codec", "mp3", "codec to use")
	out := flag.String("out", "", "output file")
	flag.Parse()

	// Check in and out
	if *in == "" || *out == "" {
		log.Fatal("-in and -out files must be specified")
	}

	// Find audio decoder
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

	// open codec
	if err := ff.AVCodec_open(ctx, codec, nil); err != nil {
		log.Fatal(err)
	}

	// open file for reading
	r, err := os.Open(*in)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	// open file for writing
	w, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	// Create a packet
	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		log.Fatal("Could not allocate packet")
	}
	defer ff.AVCodec_packet_free(packet)

	// Decoded frame
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		log.Fatal("Could not allocate frame")
	}
	defer ff.AVUtil_frame_free(frame)

	// Decode until EOF
	inbuf := make([]byte, AUDIO_INBUF_SIZE+ff.AV_INPUT_BUFFER_PADDING_SIZE)
	data_size, err := r.Read(inbuf)
	if err != nil {
		log.Fatal(err)
	}
	for {
		fmt.Println("data_size", data_size)
		if data_size == 0 {
			break
		}

		// Parse the data
		size := ff.AVCodec_parser_parse(parser, ctx, packet, inbuf[:data_size], ff.AV_NOPTS_VALUE, ff.AV_NOPTS_VALUE, 0)
		if size < 0 {
			log.Fatal("Error while parsing")
		}
		inbuf = inbuf[size:]
		data_size -= size

		// Decode the data
		if packet.Size() > 0 {
			if err := decode(w, ctx, packet, frame); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Flush the decoder
	ff.AVCodec_packet_empty(packet)
	if err := decode(w, ctx, packet, frame); err != nil {
		log.Fatal(err)
	}
}

func decode(w io.Writer, ctx *ff.AVCodecContext, packet *ff.AVPacket, frame *ff.AVFrame) error {
	// send the packet with the compressed data to the decoder
	if err := ff.AVCodec_send_packet(ctx, packet); err != nil {
		return err
	}

	// Read all the output frames (in general there may be any number of them)
	for {
		log.Println("  receive_packet")
		if err := ff.AVCodec_receive_frame(ctx, frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			log.Println("AVCodec_receive_frame error", err)
			return err
		}

		data_size := ff.AVUtil_get_bytes_per_sample(ctx.SampleFormat())
		if data_size < 0 {
			log.Fatal("Failed to calculate data size")
		}

		for i := 0; i < frame.NumSamples(); i++ {
			for ch := 0; ch < ctx.ChannelLayout().NumChannels(); ch++ {
				buf := frame.Bytes(ch)
				_, err := w.Write(buf[i*data_size : data_size])
			}
		}
	}
}
