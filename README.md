
# go-media

This module provides an interface for media services, mostly based on bindings
for [FFmpeg](https://ffmpeg.org/). It is designed to be used in a pipeline
for processing media files, and is not a standalone application.

You'd want to use this module if you want to integrate media processing into an
existing pipeline, not necessarily to build a standalone application, where
you can already use a command like [FFmpeg](https://ffmpeg.org/) or 
[GStreamer](https://gstreamer.freedesktop.org/).

## Current Status

This module is currently in development and subject to change. If there are any specific features
you are interested in, please see below "Contributing & Distribution" below.

## What do you want to do?

Here is some examples of how you might want to use this module:

| Use Case | Examples        |
|----------|-----------------|
| Use low-level bindings in golang for [FFmpeg 7.1](https://ffmpeg.org/) | [here]() |
| Opening media files, devices and network sockets for reading and writing | [here]() |
| Retrieving metadata, artwork or thumbnails from audio and video media |  [here]() |
| Re-multiplexing media files from one format to another |  [here]() |
| Encoding and decoding audio, video and subtitle streams |  [here]() |
| Resampling audio and resizing video streams |  [here]() |
| Applying filters and effects to audio and video streams |  [here]() |
| Fingerprinting audio files to identify music |  [here]() |
| Creating an audio or video player | [here]() |

## Requirements

There are two ways to satisfy the dependencies on FFmpeg:

1. The module is based on [FFmpeg 7.1](https://ffmpeg.org/) and requires you to have installed the libraries
   and headers for FFmpeg. You can install the libraries using your package manager.
2. The module can download the source code for FFmpeg and build static libraries and headers
   for you. This is done using the `make` command.

Either way, in order to integrate the module into your golang code, you need to have satisfied these
dependencies and use a specific set of flags to compile your code.

### Building FFmpeg

To build FFmpeg, you need to have a compiler, nasm, pkg-config and make.

#### Debian/Ubuntu

```bash
# Required
apt install \
  build-essential cmake nasm curl

# Optional
apt install \
  libfreetype-dev libmp3lame-dev libopus-dev libvorbis-dev libvpx-dev \
  libx264-dev libx265-dev libnuma-dev

# Make ffmpeg
git clone github.com/mutablelogic/go-media
cd go-media
make ffmpeg
```

#### Fedora

TODO

```bash
# Fedora
dnf install freetype-devel lame-devel opus-devel libvorbis-devel libvpx-devel x264-devel x265-devel numactl-devel
git clone github.com/mutablelogic/go-media
cd go-media
make ffmpeg
```


#### MacOS Homebrew

TODO

```bash
# Homebrew
brew install freetype lame opus libvorbis libvpx x264 x265
git clone github.com/mutablelogic/go-media
cd go-media
make ffmpeg
```

This will place the static libraries in the `build/install` folder which you can refer to when compiling your
golang code. 

## Linking to FFmpeg

For example, here's a typical compile or run command on a Mac:

```bash
PKG_CONFIG_PATH="${PWD}/build/install/lib/pkgconfig" \
  LD_LIBRARY_PATH="/opt/homebrew/lib" \
  CGO_LDFLAGS_ALLOW="-(W|D).*" \
  go build -o build/media ./cmd/media
```
### Demultiplexing

```go
package main

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

### Decoding - Video

This example shows you how to decode video frames from a media file into images, and
encode those images to JPEG format.

```go
package main

import (
  "context"
  "fmt"
  "image/jpeg"
  "io"
  "log"
  "os"
  "path/filepath"

  // Packages
  ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"

  // Namespace imports
  . "github.com/mutablelogic/go-media"
)

func main() {
  // Open a media file for reading. The format of the file is guessed.
  input, err := ffmpeg.Open(os.Args[1])
  if err != nil {
    log.Fatal(err)
  }

  // Make a map function which can be used to decode the streams and set
  // the parameters we want from the decode. The audio and video streams
  // are resampled and resized to fit the parameters we pass back the decoder.
  mapfunc := func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
    if par.Type() == VIDEO {
      // Convert frame to yuv420p to frame size and rate as source
      return ffmpeg.VideoPar("yuv420p", par.WidthHeight(), par.FrameRate()), nil
    }
    // Ignore other streams
    return nil, nil
  }

  // Make a folder where we're going to store the thumbnails
  tmp, err := os.MkdirTemp("", "decode")
  if err != nil {
    log.Fatal(err)
  }

  // Decode the streams and receive the video frame
  // If the map function is nil, the frames are copied. In this example,
  // we get a yuv420p frame at the same size as the original.
  n := 0
  err = input.Decode(context.Background(), mapfunc, func(stream int, frame *ffmpeg.Frame) error {
    // Write the frame to a file
    w, err := os.Create(filepath.Join(tmp, fmt.Sprintf("frame-%d-%d.jpg", stream, n)))
    if err != nil {
      return err
    }
    defer w.Close()

    // Convert to an image and encode a JPEG
    if image, err := frame.Image(); err != nil {
      return err
    } else if err := jpeg.Encode(w, image, nil); err != nil {
      return err
    } else {
      log.Println("Wrote:", w.Name())
    }

    // End after 10 frames
    n++
    if n >= 10 {
      return io.EOF
    }
    return nil
  })
  if err != nil {
    log.Fatal(err)
  }
}
```

### Encoding - Audio and Video

This example shows you how to encode video and audio frames into a media file.
It creates a testcard signal overlayed with a timestamp, and a 1KHz tone at -5dB

```go
package main

import (
  "fmt"
  "io"
  "log"
  "os"

  // Packages
  ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
  generator "github.com/mutablelogic/go-media/pkg/generator"
)

// This example encodes an audio an video stream to a file
func main() {
  // Create a new file with an audio and video stream
  file, err := ffmpeg.Create(os.Args[1],
    ffmpeg.OptStream(1, ffmpeg.VideoPar("yuv420p", "1024x720", 30)),
    ffmpeg.OptStream(2, ffmpeg.AudioPar("fltp", "mono", 22050)),
  )
  if err != nil {
    log.Fatal(err)
  }
  defer file.Close()

  // Make an video generator which can generate frames with the same
  // parameters as the video stream
  video, err := generator.NewEBU(file.Stream(1).Par())
  if err != nil {
    log.Fatal(err)
  }
  defer video.Close()

  // Make an audio generator which can generate a 1KHz tone
  // at -5dB with the same parameters as the audio stream
  audio, err := generator.NewSine(1000, -5, file.Stream(2).Par())
  if err != nil {
    log.Fatal(err)
  }
  defer audio.Close()

  // Write 90 seconds, passing video and audio frames to the encoder
  // and returning io.EOF when the duration is reached
  duration := float64(90)
  err = file.Encode(func(stream int) (*ffmpeg.Frame, error) {
    var frame *ffmpeg.Frame
    switch stream {
    case 1:
      frame = video.Frame()
    case 2:
      frame = audio.Frame()
    }
    if frame != nil && frame.Ts() < duration {
      return frame, nil
    }
    return nil, io.EOF
  }, nil)
  if err != nil {
    log.Fatal(err)
  }
}
```

### Multiplexing

TODO

### Retrieving Metadata and Artwork from a media file

Here is an example of opening a media file and retrieving metadata and artwork.
You have to read the artwork separately from the metadata.

```go
package main

import (
  "log"
  "os"

  // Packages
  ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

func main() {
  // Open a media file for reading. The format of the file is guessed.
  reader, err := ffmpeg.Open(os.Args[1])
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
  // which needs to be converted to an image. 
  for _, artwork := range reader.Metadata(ffmpeg.MetaArtwork) {
    mimetype := artwork.Value()
    if mimetype != "" {
      // Retrieve the data using the metadata.Bytes() method
      log.Print("We got some artwork of mimetype ", mimetype)
    }
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

> __go-media__\
> [https://github.com/mutablelogic/go-media/](https://github.com/mutablelogic/go-media/)\
> Copyright (c) 2021-2025 David Thorpe, All rights reserved.

This software links to shared libraries of [FFmpeg](http://ffmpeg.org/) licensed under
the [LGPLv2.1](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html).

## References

* <https://ffmpeg.org/doxygen/7.0/index.html>
* <https://pkg.go.dev/github.com/mutablelogic/go-media>
