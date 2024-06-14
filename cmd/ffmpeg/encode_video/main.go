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

func main() {
	out := flag.String("out", "", "output file")
	codec_name := flag.String("codec", "mpeg1video", "codec to use")
	size := flag.String("size", "352x288", "video size")
	flag.Parse()

	// Check out and size
	if *out == "" {
		log.Fatal("-out argument must be specified")
	}
	width, height, err := ff.AVUtil_parse_video_size(*size)
	if err != nil {
		log.Fatal(err)
	}

	/* find the mpeg1video encoder */
	codec := ff.AVCodec_find_encoder_by_name(*codec_name)
	if codec == nil {
		log.Fatal("Codec not found")
	}

	// Allocate a codec
	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		log.Fatal("Could not allocate video codec context")
	}
	defer ff.AVCodec_free_context(ctx)

	// Set codec parameters
	ctx.SetBitRate(400000)
	ctx.SetWidth(width) // resolution must be a multiple of two
	ctx.SetHeight(height)
	ctx.SetTimeBase(ff.AVUtil_rational(1, 25))
	ctx.SetFramerate(ff.AVUtil_rational(25, 1))

	// Emit one intra frame every ten frames. Check frame pict_type before passing frame
	// to encoder, if frame->pict_type is AV_PICTURE_TYPE_I then gop_size is ignored and
	// the output of encoder will always be I frame irrespective to gop_size
	ctx.SetGopSize(10)
	ctx.SetMaxBFrames(1)
	ctx.SetPixFmt(ff.AV_PIX_FMT_YUV420P)
	if codec.ID() == ff.AV_CODEC_ID_H264 {
		if err := ctx.SetPrivDataKV("preset", "slow"); err != nil {
			log.Fatal(err)
		}
	}

	// Open the codec
	if err := ff.AVCodec_open(ctx, codec, nil); err != nil {
		log.Fatal(err)
	}

	// Create the file
	w, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	// Packet for holding encoded data
	pkt := ff.AVCodec_packet_alloc()
	if pkt == nil {
		log.Fatal("Could not allocate packet")
	}
	defer ff.AVCodec_packet_free(pkt)

	// Frame containing input raw audio
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		log.Fatal("Could not allocate video frame")
	}
	defer ff.AVUtil_frame_free(frame)

	// Set the frame parameters
	frame.SetPixFmt(ctx.PixFmt())
	frame.SetWidth(ctx.Width())
	frame.SetHeight(ctx.Height())

	// Allocate the data buffers
	if err := ff.AVUtil_frame_get_buffer(frame, 0); err != nil {
		log.Fatal(err)
	}

	// Encode 5 seconds of video
	for i := 0; i < 25*5; i++ {
		// Make sure the frame data is writable.
		// On the first round, the frame is fresh from av_frame_get_buffer()
		// and therefore we know it is writable.
		// But on the next rounds, encode() will have called
		// avcodec_send_frame(), and the codec may have kept a reference to
		// the frame in its internal structures, that makes the frame
		// unwritable.
		// av_frame_make_writable() checks that and allocates a new buffer
		// for the frame only if necessary.
		if err := ff.AVUtil_frame_make_writable(frame); err != nil {
			log.Fatal(err)
		}

		// Prepare a dummy image. In real code, this is where you would have your own logic for
		// filling the frame. FFmpeg does not care what you put in the frame.
		fill_yuv_image(frame, i)

		// Set timestamp in the frame
		frame.SetPts(int64(i))

		// Encode the image
		if err := encode(w, ctx, frame, pkt); err != nil {
			log.Fatal(err)
		}
	}

	// Flush the encoder
	log.Println("flush")
	if err := encode(w, ctx, nil, pkt); err != nil {
		log.Fatal(err)
	}

	// Add sequence end code to have a real MPEG file. It makes only sense because this tiny examples writes packets
	// directly. This is called "elementary stream" and only works for some codecs. To create a valid file, you
	// usually need to write packets into a proper file format or protocol; see mux.c.
	if codec.ID() == ff.AV_CODEC_ID_MPEG1VIDEO || codec.ID() == ff.AV_CODEC_ID_MPEG2VIDEO {
		w.Write([]byte{0, 0, 1, 0xb7})
	}
}

func encode(w io.Writer, ctx *ff.AVCodecContext, frame *ff.AVFrame, pkt *ff.AVPacket) error {
	// Send the frame for encoding, if the frame is nil then flush instead
	if frame != nil {
		log.Println("  send frame pts", frame.Pts())
	}
	if err := ff.AVCodec_send_frame(ctx, frame); err != nil {
		log.Println("Error sending frame", err)
		return err
	}

	// Read all the available output packets (in general there may be any number of them)
	for {
		log.Println("  receive_packet")
		if err := ff.AVCodec_receive_packet(ctx, pkt); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			log.Println("AVCodec_receive_packet error", err)
			return err
		}
		// Write the packet to the output file
		//log.Println("  write_packet: ", pkt)
		if _, err := w.Write(pkt.Bytes()); err != nil {
			return err
		}
		// Release packet data
		ff.AVCodec_packet_unref(pkt)
	}
}

func fill_yuv_image(frame *ff.AVFrame, frame_index int) {
	width := frame.Width()
	height := frame.Height()

	/* Y */
	ydata := frame.Uint8(0)
	ystride := frame.Linesize(0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ydata[y*ystride+x] = uint8(x + y + frame_index*3)
		}
	}

	/* Cb and Cr */
	cbdata := frame.Uint8(1)
	cbstride := frame.Linesize(1)
	crdata := frame.Uint8(2)
	crstride := frame.Linesize(2)
	fmt.Println("cbstride", cbstride, "crstride", crstride)

	for y := 0; y < height>>1; y++ {
		for x := 0; x < width>>1; x++ {
			cbdata[y*cbstride+x] = uint8(128 + y + frame_index*2)
			crdata[y*crstride+x] = uint8(64 + x + frame_index*5)
		}
	}
}
