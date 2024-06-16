package main

import (
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////


// Prepare a 16 bit dummy audio frame of 'frame_size' samples and 'nb_channels' channels
func get_audio_frame(stream *Stream) *AVFrame {
	 AVFrame *frame = stream.tmp_frame

	 int j, i, v;
	 int16_t *q = (int16_t*)frame->data[0];
  
	 /* check if we want to generate more frames */
	 if (av_compare_ts(ost->next_pts, ost->enc->time_base,
					   STREAM_DURATION, (AVRational){ 1, 1 }) > 0)
		 return NULL;
  
	 for (j = 0; j <frame->nb_samples; j++) {
		 v = (int)(sin(ost->t) * 10000);
		 for (i = 0; i < ost->enc->ch_layout.nb_channels; i++)
			 *q++ = v;
		 ost->t     += ost->tincr;
		 ost->tincr += ost->tincr2;
	 }
  
	 frame->pts = ost->next_pts;
	 ost->next_pts  += frame->nb_samples;
  
	 return frame;
 }

/*
 * encode one audio frame and send it to the muxer
 * return true when encoding is finished
 */
func write_audio_frame(ctx *ff.AVFormatContext, stream *Stream) bool {
	frame := get_audio_frame(stream)
	if frame != nil {
		// convert samples from native format to destination codec format, using the resampler
		// compute destination number of samples
		delay := ff.SWResample_get_delay(stream.SWRContext(), stream.Encoder().SampleRate()) + frame.NumSamples()
		dst_nb_samples := ff.AVUtil_rescale_rnd(delay,ctx.SampleRate(),ctx.SampleRate(),ff.AV_ROUND_UP);
  
		 // When we pass a frame to the encoder, it may keep a reference to it internally; make sure we do not overwrite it here
		 ff.AVUtil_frame_make_writable(stream.Frame())
  
		 // Convert to destination format
		 ff.SWResample_convert_frame(stream.swr_ctx,frame,stream.frame)

			 return false
		 }
		 frame = ost->frame;
  
		 frame->pts = av_rescale_q(ost->samples_count, (AVRational){1, c->sample_rate}, c->time_base);
		 ost->samples_count += dst_nb_samples;
	}
	return write_frame(ctx, stream, frame)
}

func write_video_frame(ctx *ff.AVFormatContext, stream *Stream) bool {
	return true
}




 static int write_audio_frame(AVFormatContext *oc, OutputStream *ost)
 {
	 AVCodecContext *c;
	 AVFrame *frame;
	 int ret;
	 int dst_nb_samples;
  
	 c = ost->enc;
  
	 frame = get_audio_frame(ost);
  
	 if (frame) {
		 /* convert samples from native format to destination codec format, using the resampler */
		 /* compute destination number of samples */
		 dst_nb_samples = av_rescale_rnd(swr_get_delay(ost->swr_ctx, c->sample_rate) + frame->nb_samples,
										 c->sample_rate, c->sample_rate, AV_ROUND_UP);
		 av_assert0(dst_nb_samples == frame->nb_samples);
  
		 /* when we pass a frame to the encoder, it may keep a reference to it
		  * internally;
		  * make sure we do not overwrite it here
		  */
		 ret = av_frame_make_writable(ost->frame);
		 if (ret < 0)
			 exit(1);
  
		 /* convert to destination format */
		 ret = swr_convert(ost->swr_ctx,
						   ost->frame->data, dst_nb_samples,
						   (const uint8_t **)frame->data, frame->nb_samples);
		 if (ret < 0) {
			 fprintf(stderr, "Error while converting\n");
			 exit(1);
		 }
		 frame = ost->frame;
  
		 frame->pts = av_rescale_q(ost->samples_count, (AVRational){1, c->sample_rate}, c->time_base);
		 ost->samples_count += dst_nb_samples;
	 }
  
	 return write_frame(oc, c, ost->st, frame, ost->tmp_pkt);
 }