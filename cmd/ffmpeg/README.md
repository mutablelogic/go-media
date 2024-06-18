# ffmpeg examples

This directory contains examples of how to use ffmpeg based on the
examples [here](https://ffmpeg.org/doxygen/6.1/examples.html) but
using the low-level golang bindings.

* [decode_audio](decode_audio) - libavcodec decoding audio API usage example.
    Decode data from an MP2 input file and generate a raw audio file to be played with ffplay.
* [decode_video](decode_video) - libavcodec decoding video API usage example.
    Read from an MPEG1 video file, decode frames, and generate PGM images as output.
* [demux_decode](demux_decode) - ibavformat and libavcodec demuxing and decoding API usage example.
    Show how to use the libavformat and libavcodec API to demux and decode audio
    and video data. Write the output as raw audio and video files to be played by ffplay.
* [encode_audio](encode_audio) - libavcodec encoding audio API usage example.
    Generate a synthetic audio signal and encode it to an output MP2 file.
* [encode_video](encode_video) - libavcodec encoding video API usage example.
    Generate synthetic video data and encode it to an output file.
* [mux](mux) - Muxing - libavformat/libavcodec muxing API usage example - NOT COMPLETED
    Generate a synthetic audio signal and mux it into a container format.
* [remux](remux) - Remuxing - libavformat/libavcodec demuxing and muxing API usage example.
    Remux streams from one container format to another. Data is copied from the input to the output
    without transcoding.
* [resample_audio](resample_audio) - libswresample API usage example.
    Generate a synthetic audio signal, and Use libswresample API to perform audio resampling. The output
    is written to a raw audio file to be played with ffplay.
* [scale_video](scale_video) - libswscale API usage example.
    Generate a synthetic video signal and use libswscale to perform rescaling.
* [show_metadata](show_metadata) - libavformat metadata extraction API usage example.
    Show metadata from an input file.

## Running the examples

To run the examples, use `make cmd` in the root of the repository. This will build the examples into the `build` folder.
You can use a `-help` flag to see the options for each example.
