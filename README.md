
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

In order to build the examples, you'll need the library and header files 
for [FFmpeg 6](https://ffmpeg.org/download.html) installed.The `chromaprint` library is also 
required for fingerprinting audio files and SDL2 for the video player.

### MacOS

On Macintosh with [homebrew](http://bew.sh/), for example:

```bash
brew install ffmpeg@6 chromaprint make
brew link ffmpeg@6
```

### Debian

If you're using Debian you may not be able to get the ffmpeg 6 unless you first of all add the debi-multimedia repository. 
You can do this by adding the following line to your `/etc/apt/sources.list` file:

```bash
# Run commands as privileged user
echo "deb https://www.deb-multimedia.org $(lsb_release -sc) main" >> /etc/apt/sources.list
apt update -y -oAcquire::AllowInsecureRepositories=true
apt install -y --force-yes deb-multimedia-keyring
```

Then you can proceed to install the ffmpeg 6 and the other dependencies:

```bash
# Run commands as privileged user
apt install -y libavcodec-dev libavdevice-dev libavfilter-dev libavutil-dev libswscale-dev libswresample-dev
apt install -y libchromaprint-dev
apt install -y libsdl2-dev
```

### Docker Container

TODO

## Examples

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
> Copyright (c) 2021-2024 David Thorpe, All rights reserved.

This software links to shared libraries of [FFmpeg](http://ffmpeg.org/) licensed under
the [LGPLv2.1](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html).

## References

* https://ffmpeg.org/doxygen/6.1/index.html
* https://pkg.go.dev/github.com/mutablelogic/go-media
