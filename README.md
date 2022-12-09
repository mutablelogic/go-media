
# go-media

This module provides an interface for media services, including:

  * Bindings in golang for [ffmpeg 5.1](https://ffmpeg.org/);
  * Opening media files for reading and writing;
  * Retrieving metadata and artwork from audio and video media;
  * Re-multiplexing media files from one format to another;
  * Resampling raw audio from one format to another;
  * Serve a backend API for access to media services.

## Requirements

  * Library and header files for [ffmpeg 5.1](https://ffmpeg.org/download.html);

## Building

This module does not include a full
copy of __ffmpeg__ as part of the build process, but expects `pkgconfig`
files `libavcodec.pc`, `libavdevice.pc`, `libavfilter.pc`, `libavformat.pc`,
`libavresample.pc` and `libavutil.pc` to be present (and an existing set of header
files and libraries to be available to link against, of course).

You may need two environment variables set in order to locate the correct installation of 
`ffmpeg`:

  * `PKG_CONFIG_PATH` is used for locating the pkgconfig files;
  * `DYLD_LIBRARY_PATH` is used for locating a dynamic library when testing and/or running
    if linked dynamically.

On Macintosh with homebrew, for example:

```bash
[bash] brew install ffmpeg chromaprint
[bash] git clone git@github.com:djthorpe/go-media.git
[bash] cd go-media
[bash] PKG_CONFIG_PATH="/usr/local/lib/pkgconfig" make
```

On Debian Linux you shouldn't need to locate the correct path to the sqlite3 library, since
only one copy is installed:

```bash
[bash] sudo apt install libavcodec-dev libavdevice-dev libavfilter-dev \
       libavformat-dev libavresample-dev libavutil-dev libchromaprint-dev
[bash] git clone git@github.com:djthorpe/go-media.git
[bash] cd go-media
[bash] make
```

There are some examples in the `cmd` folder of the main repository on how to use
the package. The various make targets are:

  * `make all` will perform tests, build all examples and the backend API;
  * `make test` will perform tests;
  * `make cmd` will build example command-line tools into the `build` folder;
  * `make server plugins` will install the backend server and required plugins in the `build` folder;
  * `make clean` will remove all build artifacts.

## Contributing & Distribution

__This module is currently in development and subject to change.__

Please do file feature requests and bugs [here](https://github.com/mutablelogic/go-media/issues).
The license is Apache 2 so feel free to redistribute. Redistributions in either source
code or binary form must reproduce the copyright notice, and please link back to this
repository for more information:

> Copyright (c) 2021, David Thorpe, All rights reserved.

## References

  * https://ffmpeg.org/doxygen/4.1/index.html

