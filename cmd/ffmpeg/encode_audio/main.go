package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"math"
	"os"
	"slices"
	"syscall"

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
	ctx.SetSampleRate(select_sample_rate(codec))
	if err := ctx.SetChannelLayout(ff.AV_CHANNEL_LAYOUT_MONO); err != nil {
		log.Fatal(err)
	}

	// Check
	if !check_sample_fmt(codec, ctx.SampleFormat()) {
		log.Fatalf("Encoder does not support sample format %v", ctx.SampleFormat())
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
		log.Fatal("Could not allocate audio frame")
	}
	defer ff.AVUtil_frame_free(frame)

	// Set the frame parameters
	frame.SetNumSamples(ctx.FrameSize())
	frame.SetSampleFormat(ctx.SampleFormat())
	if err := frame.SetChannelLayout(ctx.ChannelLayout()); err != nil {
		log.Fatal(err)
	}

	// Allocate the data buffers
	if err := ff.AVUtil_frame_get_buffer(frame, 0); err != nil {
		log.Fatal(err)
	}

	// Encode a single tone sound
	t := float64(0)
	tincr := 2 * math.Pi * 440.0 / float64(ctx.SampleRate())
	num_channels := ctx.ChannelLayout().NumChannels()

	for i := 0; i < 200; i++ {
		log.Println("frame", i)

		// Make sure the frame is writable -- makes a copy if the encoder kept a reference internally
		if err := ff.AVUtil_frame_make_writable(frame); err != nil {
			log.Fatal(err)
		}

		// Set samples in the frame
		samples := frame.Int16(0)
		for j := 0; j < ctx.FrameSize(); j++ {
			// Set sample on first channel
			samples[j*num_channels] = (int16)(math.Sin(t) * 10000)

			// Copy to other channels
			for k := 1; k < num_channels; k++ {
				samples[j+k] = samples[j]
			}

			// Increment the time
			t += tincr
		}

		// Encode the frame
		if err := encode(w, ctx, frame, pkt); err != nil {
			log.Fatal(err)
		}
	}

	// Flush the encoder
	log.Println("flush")
	if err := encode(w, ctx, nil, pkt); err != nil {
		log.Fatal(err)
	}
}

// Check that a given sample format is supported by the encoder
func check_sample_fmt(codec *ff.AVCodec, sample_fmt ff.AVSampleFormat) bool {
	return slices.Contains(codec.SampleFormats(), sample_fmt)
}

// Pick the highest supported samplerate
func select_sample_rate(codec *ff.AVCodec) int {
	samplerates := codec.SupportedSamplerates()
	if len(samplerates) == 0 {
		return 44100
	}
	best_samplerate := 0
	for _, rate := range samplerates {
		if rate > best_samplerate {
			best_samplerate = rate
		}
	}
	return best_samplerate
}

func encode(w io.Writer, ctx *ff.AVCodecContext, frame *ff.AVFrame, pkt *ff.AVPacket) error {
	// Send the frame for encoding, if the frame is nil then flush instead
	log.Println("  send frame")
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
		if _, err := w.Write(pkt.Bytes()); err != nil {
			return err
		}
		// Release packet data
		ff.AVCodec_packet_unref(pkt)
	}
}
