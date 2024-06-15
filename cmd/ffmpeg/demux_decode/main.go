package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
	"unsafe"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

// Create a structure to hold the image
type Image struct {
	Data   [][]byte
	Stride []int
	Size   int
}

func NewImage(width, height int, fmt ff.AVPixelFormat, align int) (*Image, error) {
	// Allocate image
	data, stride, size, err := ff.AVUtil_image_alloc(width, height, fmt, align)
	if err != nil {
		return nil, err
	} else {
		return &Image{Data: data, Stride: stride, Size: size}, nil
	}
}

func (i *Image) Free() {
	ff.AVUtil_image_free(i.Data)
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

func main() {
	in := flag.String("in", "", "input file")
	aout := flag.String("audio-out", "", "raw audio output file")
	vout := flag.String("video-out", "", "raw video output file")
	flag.Parse()

	// Check in and out
	if *in == "" || (*aout == "" && *vout == "") {
		log.Fatal("-in and at least one of -audio-out and -video-out flags must be specified")
	}

	// Ppen input file, and allocate format context
	ctx, err := ff.AVFormat_open_url(*in, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ff.AVFormat_close_input(ctx)

	// Find stream information
	if err := ff.AVFormat_find_stream_info(ctx, nil); err != nil {
		log.Fatal(err)
	}

	// Dump the input format
	ff.AVFormat_dump_format(ctx, 0, *in)

	var audio_decoder_ctx, video_decoder_ctx *ff.AVCodecContext
	var wa, wv io.Writer
	var image *Image

	audio_stream, video_stream := -1, -1

	// Get decoder for audio, and create an output file
	if *aout != "" {
		if stream, ctx, err := open_codec_context(ctx, ff.AVMEDIA_TYPE_AUDIO); err != nil {
			log.Fatal(err)
		} else {
			audio_decoder_ctx = ctx
			audio_stream = stream
		}
		defer ff.AVCodec_free_context(audio_decoder_ctx)

		// Create output file
		if w, err := os.Create(*aout); err != nil {
			log.Fatal(err)
		} else {
			wa = w
		}
	}

	// Get decoder for video, and create an output file
	if *vout != "" {
		if stream, ctx, err := open_codec_context(ctx, ff.AVMEDIA_TYPE_VIDEO); err != nil {
			log.Fatal(err)
		} else {
			video_decoder_ctx = ctx
			video_stream = stream
		}
		defer ff.AVCodec_free_context(video_decoder_ctx)

		// Create output file
		if w, err := os.Create(*vout); err != nil {
			log.Fatal(err)
		} else {
			wv = w
		}

		// Allocate an image for the video frame
		if frame, err := NewImage(video_decoder_ctx.Width(), video_decoder_ctx.Height(), video_decoder_ctx.PixFmt(), 1); err != nil {
			log.Fatal(err)
		} else {
			image = frame
		}
		defer image.Free()
	}

	// Allocate a frame
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		log.Fatal("Could not allocate frame")
	}
	defer ff.AVUtil_frame_free(frame)

	// Allocate a packet
	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		log.Fatal("failed to allocate packet")
	}
	defer ff.AVCodec_packet_free(packet)

	// Read frames
	for {
		err := ff.AVFormat_read_frame(ctx, packet)
		if errors.Is(err, io.EOF) {
			break
		}
		// check if the packet belongs to a stream we are interested in, otherwise skip it
		if packet.StreamIndex() == audio_stream {
			if err := decode_packet(wa, wv, audio_decoder_ctx, packet, frame, image); err != nil {
				log.Fatal(err)
			}
		} else if packet.StreamIndex() == video_stream {
			if err := decode_packet(wa, wv, video_decoder_ctx, packet, frame, image); err != nil {
				log.Fatal(err)
			}
		}

		// Unreference the packet
		ff.AVCodec_packet_unref(packet)
	}

	// Flush the decoders
	if audio_decoder_ctx != nil {
		if err := decode_packet(wa, wv, audio_decoder_ctx, nil, frame, nil); err != nil {
			log.Fatal(err)
		}
	}
	if video_decoder_ctx != nil {
		if err := decode_packet(wa, wv, video_decoder_ctx, nil, frame, image); err != nil {
			log.Fatal(err)
		}
	}

	// Output command for playing the video
	if *vout != "" {
		log.Print("Play the output video file with the command:")
		log.Printf("  ffplay -f rawvideo -pixel_format %s -video_size %dx%d %s", ff.AVUtil_get_pix_fmt_name(video_decoder_ctx.PixFmt()), video_decoder_ctx.Width(), video_decoder_ctx.Height(), *vout)
	}

	// Output command for playing the audio
	if *aout != "" {
		fmt := audio_decoder_ctx.SampleFormat()
		num_channels := audio_decoder_ctx.ChannelLayout().NumChannels()
		if ff.AVUtil_sample_fmt_is_planar(audio_decoder_ctx.SampleFormat()) {
			fmt = ff.AVUtil_get_packed_sample_fmt(fmt)
			num_channels = 1
			log.Print("Warning: the sample format the decoder produced is planar. This example will output the first channel only.")
		}
		if fmt, err := get_format_from_sample_fmt(fmt); err != nil {
			log.Fatal(err)
		} else {
			log.Print("Play the output audio file with the command:")
			log.Printf("  ffplay -f %s -ar %d -ac %d -i %s", fmt, audio_decoder_ctx.SampleRate(), num_channels, *aout)
		}
	}
}

func open_codec_context(ctx *ff.AVFormatContext, media_type ff.AVMediaType) (int, *ff.AVCodecContext, error) {
	stream_num, codec, err := ff.AVFormat_find_best_stream(ctx, media_type, -1, -1)
	if err != nil {
		return -1, nil, err
	}

	// Find the decoder for the stream
	decoder := ff.AVCodec_find_decoder(codec.ID())
	if decoder == nil {
		return -1, nil, fmt.Errorf("failed to find decoder for codec %s", codec.Name())
	}

	// Allocate a codec context for the decoder
	dec_ctx := ff.AVCodec_alloc_context(decoder)
	if dec_ctx == nil {
		return -1, nil, fmt.Errorf("failed to allocate codec context for codec %s", codec.Name())
	}

	// Copy codec parameters from input stream to output codec context
	stream := ctx.Stream(stream_num)
	if err := ff.AVCodec_parameters_to_context(dec_ctx, stream.CodecPar()); err != nil {
		ff.AVCodec_free_context(dec_ctx)
		return -1, nil, fmt.Errorf("failed to copy codec parameters to decoder context for codec %s", codec.Name())
	}

	// Init the decoder
	if err := ff.AVCodec_open(dec_ctx, decoder, nil); err != nil {
		ff.AVCodec_free_context(dec_ctx)
		return -1, nil, err
	}

	// Return success
	return stream_num, dec_ctx, nil
}

func decode_packet(wa, wv io.Writer, ctx *ff.AVCodecContext, packet *ff.AVPacket, frame *ff.AVFrame, image *Image) error {
	// submit the packet to the decoder
	if err := ff.AVCodec_send_packet(ctx, packet); err != nil {
		return err
	}

	// get all the available frames from the decoder
	for {
		if err := ff.AVCodec_receive_frame(ctx, frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			// Finished decoding packet or EOF
			return nil
		} else if err != nil {
			return err
		}

		// write the frame data to output file
		if ctx.Codec().Type() == ff.AVMEDIA_TYPE_AUDIO {
			if err := write_audio_frame(wa, frame); err != nil {
				return err
			}
		} else if ctx.Codec().Type() == ff.AVMEDIA_TYPE_VIDEO {
			if err := write_video_frame(wv, frame, image); err != nil {
				return err
			}
		}
	}
}

func write_audio_frame(w io.Writer, frame *ff.AVFrame) error {
	/* Write the raw audio data samples of the first plane. This works
	 * fine for packed formats (e.g. AV_SAMPLE_FMT_S16). However,
	 * most audio decoders output planar audio, which uses a separate
	 * plane of audio samples for each channel (e.g. AV_SAMPLE_FMT_S16P).
	 * In other words, this code will write only the first audio channel
	 * in these cases.
	 * You should use libswresample or libavfilter to convert the frame
	 * to packed data. */
	log.Printf("audio_frame format:%s nb_samples:%d pts:%s", ff.AVUtil_get_sample_fmt_name(frame.SampleFormat()), frame.NumSamples(), ff.AVUtil_ts2str(frame.Pts()))

	n := frame.NumSamples() * ff.AVUtil_get_bytes_per_sample(frame.SampleFormat())
	data := frame.Uint8(0)
	if _, err := w.Write(data[:n]); err != nil {
		return err
	}

	// Return success
	return nil
}

func write_video_frame(w io.Writer, frame *ff.AVFrame, image *Image) error {
	// copy decoded frame to destination buffer: this is required since rawvideo expects non aligned data
	log.Printf("video_frame format:%s size:%dx%d pts:%s", ff.AVUtil_get_pix_fmt_name(frame.PixFmt()), frame.Width(), frame.Height(), ff.AVUtil_ts2str(frame.Pts()))

	src_data, src_stride := frame.Data()
	ff.AVUtil_image_copy(image.Data, image.Stride, src_data, src_stride, frame.PixFmt(), frame.Width(), frame.Height())

	// Write each plane
	for i := 0; i < len(image.Data); i++ {
		if _, err := w.Write(image.Data[i]); err != nil {
			return err
		}
	}

	// Return success
	return nil
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
