/**
 * libswresample API use example.
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	// Package imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
)

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s input output\n", filepath.Base(flag.CommandLine.Name()))
		os.Exit(1)
	}

	// create resampler context
	ctx := ffmpeg.SWR_alloc()
	if ctx == nil {
		fmt.Fprintln(os.Stderr, "Could not allocate resampler context")
		os.Exit(1)
	}
	defer ctx.SWR_free()

	// Set parameters
	src_ch_layout := ffmpeg.AV_CHANNEL_LAYOUT_STEREO
	src_rate := int64(48000)
	src_sample_fmt := ffmpeg.AV_SAMPLE_FMT_DBL
	ctx.AVUtil_av_opt_set_chlayout("in_chlayout", &src_ch_layout)
	ctx.AVUtil_av_opt_set_int("in_sample_rate", src_rate)
	ctx.AVUtil_av_opt_set_sample_fmt("in_sample_fmt", src_sample_fmt)

	dst_ch_layout := ffmpeg.AV_CHANNEL_LAYOUT_SURROUND
	dst_rate := int64(44100)
	dst_sample_fmt := ffmpeg.AV_SAMPLE_FMT_S16
	ctx.AVUtil_av_opt_set_chlayout("out_chlayout", &dst_ch_layout)
	ctx.AVUtil_av_opt_set_int("out_sample_rate", dst_rate)
	ctx.AVUtil_av_opt_set_sample_fmt("out_sample_fmt", dst_sample_fmt)

	// initialize the resampling context
	if err := ctx.SWR_init(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

/*

   // allocate source and destination samples buffers

   src_nb_channels = src_ch_layout.nb_channels;
   ret = av_samples_alloc_array_and_samples(&src_data, &src_linesize, src_nb_channels,
											src_nb_samples, src_sample_fmt, 0);
   if (ret < 0) {
	   fprintf(stderr, "Could not allocate source samples\n");
	   goto end;
   }

   // compute the number of converted samples: buffering is avoided
//ensuring that the output buffer will contain at least all the
	//converted input samples
   max_dst_nb_samples = dst_nb_samples =
	   av_rescale_rnd(src_nb_samples, dst_rate, src_rate, AV_ROUND_UP);

   // buffer is going to be directly written to a rawaudio file, no alignment
   dst_nb_channels = dst_ch_layout.nb_channels;
   ret = av_samples_alloc_array_and_samples(&dst_data, &dst_linesize, dst_nb_channels,
											dst_nb_samples, dst_sample_fmt, 0);
   if (ret < 0) {
	   fprintf(stderr, "Could not allocate destination samples\n");
	   goto end;
   }

   t = 0;
   do {
	   // generate synthetic audio
	   fill_samples((double *)src_data[0], src_nb_samples, src_nb_channels, src_rate, &t);

	   // compute destination number of samples
	   dst_nb_samples = av_rescale_rnd(swr_get_delay(swr_ctx, src_rate) +
									   src_nb_samples, dst_rate, src_rate, AV_ROUND_UP);
	   if (dst_nb_samples > max_dst_nb_samples) {
		   av_freep(&dst_data[0]);
		   ret = av_samples_alloc(dst_data, &dst_linesize, dst_nb_channels,
								  dst_nb_samples, dst_sample_fmt, 1);
		   if (ret < 0)
			   break;
		   max_dst_nb_samples = dst_nb_samples;
	   }

	   // convert to destination format
	   ret = swr_convert(swr_ctx, dst_data, dst_nb_samples, (const uint8_t **)src_data, src_nb_samples);
	   if (ret < 0) {
		   fprintf(stderr, "Error while converting\n");
		   goto end;
	   }
	   dst_bufsize = av_samples_get_buffer_size(&dst_linesize, dst_nb_channels,
												ret, dst_sample_fmt, 1);
	   if (dst_bufsize < 0) {
		   fprintf(stderr, "Could not get sample buffer size\n");
		   goto end;
	   }
	   printf("t:%f in:%d out:%d\n", t, src_nb_samples, ret);
	   fwrite(dst_data[0], 1, dst_bufsize, dst_file);
   } while (t < 10);

   if ((ret = get_format_from_sample_fmt(&fmt, dst_sample_fmt)) < 0)
	   goto end;
   av_channel_layout_describe(&dst_ch_layout, buf, sizeof(buf));
   fprintf(stderr, "Resampling succeeded. Play the output file with the command:\n"
		   "ffplay -f %s -channel_layout %s -channels %d -ar %d %s\n",
		   fmt, buf, dst_nb_channels, dst_rate, dst_filename);

end:
   fclose(dst_file);

   if (src_data)
	   av_freep(&src_data[0]);
   av_freep(&src_data);

   if (dst_data)
	   av_freep(&dst_data[0]);
   av_freep(&dst_data);

   swr_free(&swr_ctx);
   return ret < 0;
}
*/
