
# go-media

This module provides an interface for media services, including:

  * Bindings in golang for [ffmpeg](https://ffmpeg.org/);
  * Opening media files for reading and writing;
  * Retrieving metadata and artwork from audio and video media;
  * Re-multiplexing media files from one format to another;
  * Serve a backend API for access to media services.

Presently the module is in development and the API is subject to change.

| If you want to...                    |  Folder         | Documentation |
|--------------------------------------|-----------------|---------------|
| Use the lower-level ffmpeg bindings similar to the [C API](https://ffmpeg.org/doxygen/trunk/) | [sys/ffmpeg](https://github.com/mutablelogic/go-media/tree/master/sys/ffmpeg) | [README.md](https://github.com/mutablelogic/go-media/blob/master/sys/ffmpeg/README.md) |
| Use the high-level media manager package for reading, writing, multiplexing and transcoding| [pkg/media](https://github.com/mutablelogic/go-media/tree/master/pkg/media) | [README.md](https://github.com/mutablelogic/go-media/blob/master/pkg/media/README.md) |
| Implement or use a REST API for media files | [plugin/media](https://github.com/mutablelogic/go-media/tree/master/plugin/media) | [README.md](https://github.com/mutablelogic/go-media/blob/master/plugin/media/README.md) |
| See example command-line tools | [cmd](https://github.com/mutablelogic/go-media/tree/master/cmd) | [README.md](https://github.com/mutablelogic/go-media/blob/master/cmd/README.md) |

## Requirements

  * Library and header files for [ffmpeg](https://ffmpeg.org/download.html);
  * Library and header files for [chromaprint](https://github.com/acoustid/chromaprint);
  * [go1.17](https://golang.org/dl/) or later;
  * Tested on Debian Linux (32- and 64- bit) on ARM and macOS on x64
    architectures.

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

