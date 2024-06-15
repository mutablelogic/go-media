package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"syscall"
	"unsafe"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

const (
	AUDIO_INBUF_SIZE    = 20480
	AUDIO_REFILL_THRESH = 4096
)

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

func main() {
	in := flag.String("in", "", "input file")
	codec_name := flag.String("codec", "mp3", "input codec to use")
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

	log.Print("in file=", r.Name())
	log.Print("  codec=", codec)

	// open file for writing
	w, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	log.Print("out file=", w.Name())

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
		log.Print("in data_size=", data_size)
		if data_size == 0 {
			break
		}

		// Parse the data
		size := ff.AVCodec_parser_parse(parser, ctx, packet, inbuf, ff.AV_NOPTS_VALUE, ff.AV_NOPTS_VALUE, 0)
		if size < 0 {
			log.Fatal("Error while parsing")
		}
		log.Print("parsed bytes=", size)
		log.Print("packet to decode size=", packet.Size())

		// Adjust the input buffer beyond the parsed data
		inbuf = inbuf[size:]
		data_size -= size

		// Decode the data
		if packet.Size() > 0 {
			if err := decode(w, ctx, packet, frame); err != nil {
				log.Fatal(err)
			}
		}

		// TODO
		/*
					if (data_size < AUDIO_REFILL_THRESH) {
			            memmove(inbuf, data, data_size);
			            data = inbuf;
			            len = fread(data + data_size, 1,
			                        AUDIO_INBUF_SIZE - data_size, f);
			            if (len > 0)
			                data_size += len;
			        }*/
	}

	// Flush the decoder
	ff.AVCodec_packet_unref(packet)
	if err := decode(w, ctx, packet, frame); err != nil {
		log.Fatal(err)
	}

	// Print output pcm infomations, because there have no metadata of pcm
	sfmt := ctx.SampleFormat()
	if ff.AVUtil_sample_fmt_is_planar(sfmt) {
		packed := ff.AVUtil_get_sample_fmt_name(sfmt)
		log.Printf("Warning: the sample format the decoder produced is planar (%s). This example will output the first channel only.\n", packed)
		sfmt = ff.AVUtil_get_packed_sample_fmt(ctx.SampleFormat())
	}

	n_channels := ctx.ChannelLayout().NumChannels()
	fmt, err := get_format_from_sample_fmt(sfmt)
	if err != nil {
		log.Fatal(err)
	}

	// Print output pcm infomations
	log.Printf("Play the output audio file with the command:\n  ffplay -f %s -ac %d -ar %d %s\n", fmt, n_channels, ctx.SampleRate(), *out)
}

func decode(w io.Writer, ctx *ff.AVCodecContext, packet *ff.AVPacket, frame *ff.AVFrame) error {
	// bytes per sample
	bytes_per_sample := ff.AVUtil_get_bytes_per_sample(ctx.SampleFormat())
	if bytes_per_sample < 0 {
		return errors.New("failed to calculate bytes per sample")
	}

	// send the packet with the compressed data to the decoder
	log.Print("decode packet bytes=", packet.Size())
	if err := ff.AVCodec_send_packet(ctx, packet); err != nil {
		return err
	}

	// Read all the output frames (in general there may be any number of them)
	for {
		log.Println("  receive_frame")
		if err := ff.AVCodec_receive_frame(ctx, frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			log.Println("AVCodec_receive_frame error", err)
			return err
		}

		log.Println("  write frame bytes=", frame.NumSamples()*bytes_per_sample*ctx.ChannelLayout().NumChannels())
		for i := 0; i < frame.NumSamples(); i++ {
			for ch := 0; ch < ctx.ChannelLayout().NumChannels(); ch++ {
				buf := frame.Uint8(ch)
				_, err := w.Write(buf[i*bytes_per_sample : (i+1)*bytes_per_sample])
				if err != nil {
					return err
				}
			}
		}
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
