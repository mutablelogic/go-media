
# go-media

This module provides an interface for media services, including:

* Bindings in golang for [ffmpeg 6](https://ffmpeg.org/);
* Opening media files, devices and network sockets for reading and writing;
* Retrieving metadata and artwork from audio and video media;
* Re-multiplexing media files from one format to another;
* Fingerprinting audio files to identify music.

## Current Status

This module is currently in development and subject to change. If there are any specific features
you are interested in, please see below "Contributing & Distribution" below.

## Requirements

In order to build the examples, you'll need the library and header files for [ffmpeg 6](https://ffmpeg.org/download.html) installed. 
The `chromaprint` library is also required for fingerprinting audio files. On Macintosh with [homebrew](http://bew.sh/), for example:

```bash
brew install ffmpeg@6 chromaprint make
```

There are some examples in the `cmd` folder of the main repository on how to use
the package. The various make targets are:

* `make all` will perform tests, build all examples and the backend API;
* `make test` will perform tests;
* `make cmd` will build example command-line tools into the `build` folder;
* `make clean` will remove all build artifacts.

There are also some targets to build a docker image:

* `DOCKER_REGISTRY=docker.io/user make docker` will build a docker image;
* `DOCKER_REGISTRY=docker.io/user make docker-push` will push the docker image to the registry.

For example,

```bash
git clone git@github.com:djthorpe/go-media.git
cd go-media
DOCKER_REGISTRY=ghcr.io/mutablelogic make docker
```

## Examples

## Media Transcoding

## Audio Fingerprinting

You can programmatically fingerprint audio files, compare fingerprints and identify music using the following packages:

* `sys/chromaprint` provides the implementation of the lower-level function calls
  to chromaprint. The documentation is [here](https://pkg.go.dev/github.com/mutablelogic/go-media/sys/chromaprint)
* `pkg/chromaprint` provides the higher-level API for fingerprinting and identifying music. The documentation
  is [here](https://pkg.go.dev/github.com/mutablelogic/go-media/pkg/chromaprint).

You'll need an API key in order to use the [AcoustID](https://acoustid.org/) service. You can get a key
[here](https://acoustid.org/login).

## Contributing & Distribution

__This module is currently in development and subject to change.__

Please do file feature requests and bugs [here](https://github.com/mutablelogic/go-media/issues).
The license is Apache 2 so feel free to redistribute. Redistributions in either source
code or binary form must reproduce the copyright notice, and please link back to this
repository for more information:

> Copyright (c) 2021-2024 David Thorpe, All rights reserved.

## References

* https://ffmpeg.org/doxygen/6.1/index.html
