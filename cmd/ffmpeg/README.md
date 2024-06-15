# ffmpeg examples

This directory contains examples of how to use ffmpeg based on the
examples [here](https://ffmpeg.org/doxygen/6.1/examples.html) but
using the low-level golang bindings.

* [decode_audio](decode_audio) - libavcodec decoding audio API usage example.
    Decode data from an MP2 input file and generate a raw audio file to be played with ffplay.
* [decode_video](decode_video) - libavcodec decoding video API usage example.
    Read from an MPEG1 video file, decode frames, and generate PGM images as output.
* [encode_audio](encode_audio) - libavcodec encoding audio API usage example.
    Generate a synthetic audio signal and encode it to an output MP2 file.
* [encode_video](encode_video) - libavcodec encoding video API usage example.
    Generate synthetic video data and encode it to an output file.
* [remux](remux) - Remuxing - libavformat/libavcodec demuxing and muxing API usage example.
    Remux streams from one container format to another. Data is copied from the input to the output without transcoding.
* [scale_video](scale_video) - libswscale API usage example.
    Generate a synthetic video signal and use libswscale to perform rescaling.
* [show_metadata](show_metadata) - libavformat metadata extraction API usage example.
    Show metadata from an input file.

## Running the examples

To run the examples, use `make cmd` in the root of the repository. This will build the examples into the `build` folder.
You can use a `-help` flag to see the options for each example.
