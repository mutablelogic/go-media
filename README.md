
# go-media

This module provides an interface for media services, including:

  * Bindings in golang for [ffmpeg 5.1](https://ffmpeg.org/);
  * Opening media files, devices and network sockets for reading 
    and writing;
  * Retrieving metadata and artwork from audio and video media;
  * Re-multiplexing media files from one format to another;
  * Fingerprinting audio files to identify music.

## Current Status

This module is currently in development and subject to change. If there are any specific features
you are interested in, please see below "Contributing & Distribution" below.

## Requirements

In order to build the examples, you'll need the library and header files for [ffmpeg 5.1](https://ffmpeg.org/download.html) installed. The `chromaprint` library is also required for fingerprinting audio files.

On Macintosh with [homebrew](http://bew.sh/), for example:

```bash
brew install ffmpeg chromaprint make
```

There are some examples in the `cmd` folder of the main repository on how to use
the package. The various make targets are:

  * `make all` will perform tests, build all examples and the backend API;
  * `make test` will perform tests;
  * `make cmd` will build example command-line tools into the `build` folder;
  * `make clean` will remove all build artifacts.

For example,

```bash
git clone git@github.com:djthorpe/go-media.git
cd go-media
make
```

## Examples

There are two example [Command Line applications](https://github.com/mutablelogic/go-media/tree/master/cmd):

  * `extractartwork` can be used to walk through a directory and extract artwork from media
    files and save the artwork into files;
  * `transcode` can be used to copy, re-mux and re-sample media files from one format to another.

You can compile both applications with `make cmd`which places the binaries into the `build` folder. 
Use the `-help` option on either application to see the options.


## The Media Transcoding API

The API is split into two parts:

  * `sys/ffmpeg51` provides the implementation of the lower-level function calls
    to ffmpeg. The documentation is [here](https://pkg.go.dev/github.com/mutablelogic/go-media/sys/ffmpeg51)
  * `pkg/media` provides the higher-level API for opening media files, reading,
    transcoding, resampling and writing media files. The interfaces and documentation
    are best read here:
      * [Audio](https://github.com/mutablelogic/go-media/blob/master/audio.go)
      * [Video](https://github.com/mutablelogic/go-media/blob/master/video.go)
      * [Media](https://github.com/mutablelogic/go-media/blob/master/media.go)
      * And [here](https://pkg.go.dev/github.com/mutablelogic/go-media/)

## Audio Fingerprinting

TODO

## Contributing & Distribution

__This module is currently in development and subject to change.__

Please do file feature requests and bugs [here](https://github.com/mutablelogic/go-media/issues).
The license is Apache 2 so feel free to redistribute. Redistributions in either source
code or binary form must reproduce the copyright notice, and please link back to this
repository for more information:

> Copyright (c) 2021, David Thorpe, All rights reserved.

## References

  * https://ffmpeg.org/doxygen/5.1/index.html

