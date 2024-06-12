# ffmpeg examples

This directory contains examples of how to use ffmpeg based on the
examples [here](https://ffmpeg.org/doxygen/6.1/examples.html) but
using the low-level golang bindings.

* [remux](remux) - Remuxing - libavformat/libavcodec demuxing and muxing API usage example. Remux streams from one container format to another. Data is copied from the input to the output without transcoding.
* [scale_video](scale_video) - libswscale API usage example. Generate a synthetic video signal and use libswscale to perform rescaling.
