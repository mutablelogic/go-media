package main

import (
	"flag"
	"fmt"
	"io"
	"log"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func main() {
	in := flag.String("in", "", "input file")
	out := flag.String("out", "", "output file")
	flag.Parse()

	// Check in and out
	if *in == "" || *out == "" {
		log.Fatal("-in and -out files must be specified")
	}

	// Allocate a packet
	pkt := ff.AVCodec_av_packet_alloc()
	if pkt == nil {
		log.Fatal("failed to allocate packet")
	}
	defer ff.AVCodec_av_packet_free(pkt)

	// Open input file
	input, err := ff.AVFormat_open_url(*in, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ff.AVFormat_close_input(input)

	// Find stream information
	if err := ff.AVFormat_find_stream_info(input, nil); err != nil {
		log.Fatal(err)
	}

	// Dump the input format
	ff.AVFormat_dump_format(input, 0, *in)

	// Open the output file
	output, err := ff.AVFormat_create_file(*out, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ff.AVFormat_close_writer(output)

	// Stream mapping
	stream_map := make([]int, input.NumStreams())
	stream_index := 0
	for i := range stream_map {
		in_stream := input.Stream(i)
		in_codec_par := in_stream.CodecPar()

		// Ignore if not audio, video or subtitle
		if in_codec_par.CodecType() != ff.AVMEDIA_TYPE_AUDIO && in_codec_par.CodecType() != ff.AVMEDIA_TYPE_VIDEO {
			stream_map[i] = -1
			continue
		}

		// Create a new stream
		stream_map[i] = stream_index
		stream_index = stream_index + 1

		// Create a new output stream
		out_stream := ff.AVFormat_new_stream(output, nil)
		if out_stream == nil {
			log.Fatal("failed to create new stream")
		}

		// Copy the codec parameters
		if err := ff.AVCodec_parameters_copy(out_stream.CodecPar(), in_codec_par); err != nil {
			log.Fatal(err)
		}

		out_stream.CodecPar().SetCodecTag(0)
	}

	// Dump the output format
	ff.AVFormat_dump_format(output, 0, *out)

	// Open the output file
	if !output.Flags().Is(ff.AVFMT_NOFILE) {
		if ctx, err := ff.AVFormat_avio_open(*out, ff.AVIO_FLAG_WRITE); err != nil {
			log.Fatal(err)
		} else {
			output.SetPb(ctx)
		}
	}

	// Write the header
	if err := ff.AVFormat_write_header(output, nil); err != nil {
		log.Fatal(err)
	}

	// Write the frames
	for {
		if err := ff.AVFormat_read_frame(input, pkt); err != nil {
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
		}
		in_stream := input.Stream(pkt.StreamIndex())
		if out_stream_index := stream_map[pkt.StreamIndex()]; out_stream_index < 0 {
			continue
		} else {
			log_packet(input, pkt, "in")
			out_stream := output.Stream(out_stream_index)

			/* copy packet */
			ff.AVCodec_av_packet_rescale_ts(pkt, in_stream.TimeBase(), out_stream.TimeBase())
			pkt.SetPos(-1)
			log_packet(output, pkt, "out")

			if err := ff.AVFormat_interleaved_write_frame(output, pkt); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Write the trailer
	if err := ff.AVFormat_write_trailer(output); err != nil {
		log.Fatal(err)
	}
}

func log_packet(ctx *ff.AVFormatContext, pkt *ff.AVPacket, tag string) {
	stream_index := pkt.StreamIndex()
	tb := ctx.Stream(stream_index).TimeBase()
	fmt.Printf("%4s: pts:%s pts_time:%s dts:%s dts_time:%s duration:%s duration_time:%s stream_index:%d\n",
		tag,
		ff.AVUtil_ts2str(pkt.Pts()), ff.AVUtil_ts2timestr(pkt.Pts(), &tb),
		ff.AVUtil_ts2str(pkt.Dts()), ff.AVUtil_ts2timestr(pkt.Dts(), &tb),
		ff.AVUtil_ts2str(pkt.Duration()), ff.AVUtil_ts2timestr(pkt.Duration(), &tb),
		stream_index,
	)
}
