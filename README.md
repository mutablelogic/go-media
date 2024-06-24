
# go-media

This module provides an interface for media services, including:

* Bindings in golang for [FFmpeg 6.1](https://ffmpeg.org/);
* Opening media files, devices and network sockets for reading and writing;
* Retrieving metadata and artwork from audio and video media;
* Re-multiplexing media files from one format to another;
* Fingerprinting audio files to identify music.

## Current Status

This module is currently in development and subject to change. If there are any specific features
you are interested in, please see below "Contributing & Distribution" below.

## Requirements

In order to build the examples, you'll need the library and header files for [FFmpeg 6](https://ffmpeg.org/download.html) installed.The `chromaprint` library is also required for fingerprinting audio files. On Macintosh with [homebrew](http://bew.sh/), for example:

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

There are a variety of types of object needed as part of media processing.
All examples require a `Manager` to be created, which is used to enumerate all supported formats
and open media files and byte streams.

* `Manager` is the main entry point for the package. It is used to open media files and byte streams,
  and enumerate supported formats, codecs, pixel formats, etc.
* `Media` is a hardware device, file or byte stream. It contains metadata, artwork, and streams.
* `Decoder` is used to demultiplex media streams. Create a decoder and enumerate the streams which
  you'd like to demultiplex. Provide the audio and video parameters if you want to resample or
  reformat the streams.
* `Encoder` is used to multiplex media streams. Create an encoder and send the output of the
  decoder to reencode the streams.

### Demultiplexing

```go
import (
  media "github.com/mutablelogic/go-media"
)

func main() {
  manager, err := media.NewManager()
  if err != nil {
    log.Fatal(err)
  }

  // Open a media file for reading. The format of the file is guessed.
  // Alteratively, you can pass a format as the second argument. Further optional
  // arguments can be used to set the format options.
  file, err := manager.Open(os.Args[1], nil)
  if err != nil {
    log.Fatal(err)
  }
  defer file.Close()

  // Choose which streams to demultiplex - pass the stream parameters
  // to the decoder. If you don't want to resample or reformat the streams,
  // then you can pass nil as the function and all streams will be demultiplexed.
  decoder, err := file.Decoder(func (stream media.Stream) (media.Parameters, error) {
    return stream.Parameters(), nil
  }
  if err != nil {
    log.Fatal(err)
  }

  // Demuliplex the stream and receive the packets. If you don't want to
  // process the packets yourself, then you can pass nil as the function
  if err := decoder.Demux(context.Background(), func(_ media.Packet) error {
    // Each packet is specific to a stream. It can be processed here
    // to receive audio or video frames, then resize or resample them,
    // for example. Alternatively, you can pass the packet to an encoder
    // to remultiplex the streams without processing them.
    return nil
  }); err != nil {
    log.Fatal(err)  
  })
}
```

### Decoding - Video Frames

This example shows you how to decode video frames from a media file into images.

```go
import (
  media "github.com/mutablelogic/go-media"
)

func main() {
  manager, err := media.NewManager()
  if err != nil {
    log.Fatal(err)
  }

  media, err := manager.Open("etc/test/sample.mp4", nil)
  if err != nil {
    log.Fatal(err)
  }
  defer media.Close()

  // Create a decoder for the media file. Only video streams are decoded
  decoder, err := media.Decoder(func(stream Stream) (Parameters, error) {
    if stream.Type() == VIDEO {
      // Copy video
      return stream.Parameters(), nil
    } else {
      // Ignore other stream types
      return nil, nil
    }
  })
  if err != nil {
    log.Fatal(err)
  }

  // The frame function is called for each frame in the stream
  framefn := func(frame Frame) error {
    image, err := frame.Image()
    if err != nil {
      return err
    }
    // TODO: Do something with the image here....
    return nil
  }

  // decode frames from the stream
  if err := decoder.Decode(context.Background(), framefn); err != nil {
    log.Fatal(err)
  }
}
```

### Encoding

TODO

### Multiplexing

TODO

### Retrieving Metadata and Artwork from a media file

Here is an example of opening a media file and retrieving metadata and artwork.

```go
package main

import (
  "log"
  "os"

  media "github.com/mutablelogic/go-media"
  file "github.com/mutablelogic/go-media/pkg/file"
)

func main() {
  manager, err := media.NewManager()
  if err != nil {
    log.Fatal(err)
  }

  // Open a media file for reading. The format of the file is guessed.
  // Alteratively, you can pass a format as the second argument. Further optional
  // arguments can be used to set the format options.
  reader, err := manager.Open(os.Args[1], nil)
  if err != nil {
    log.Fatal(err)
  }
  defer reader.Close()

  // Retrieve all the metadata from the file, and display it. If you pass
  // keys to the Metadata function, then only entries with those keys will be
  // returned.
  for _, metadata := range reader.Metadata() {
    log.Print(metadata.Key(), " => ", metadata.Value())
  }

  // Retrieve artwork by using the MetaArtwork key. The value is of type []byte.
  // which needs to be converted to an image. There is a utility method to
  // detect the image type.
  for _, artwork := range reader.Metadata(media.MetaArtwork) {
    mimetype, ext, err := file.MimeType(artwork.Value().([]byte))
    if err != nil {
      log.Fatal(err)
    }
    log.Print("got artwork", mimetype, ext)
  }
}
```

### Audio Fingerprinting

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

> go-media\
> https://github.com/mutablelogic/go-media/\
> Copyright (c) 2021-2024 David Thorpe, All rights reserved.

This software links to shared libraries of [FFmpeg](http://ffmpeg.org/) licensed under
the [LGPLv2.1](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html).

## References

* https://ffmpeg.org/doxygen/6.1/index.html
* https://pkg.go.dev/github.com/mutablelogic/go-media
