# go-media

[![Test](https://github.com/mutablelogic/go-media/actions/workflows/test.yaml/badge.svg)](https://github.com/mutablelogic/go-media/actions/workflows/test.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mutablelogic/go-media.svg)](https://pkg.go.dev/github.com/mutablelogic/go-media)

Go bindings for FFmpeg, Chromaprint, libexif, libheif, LibRaw and libfvad, with higher-level
packages for reading, decoding, encoding, filtering and remultiplexing audio and video, and
for segmenting and classifying audio. It includes `gomedia`, a CLI, client and server for
probing and encoding media files, `segmenter`, a CLI for cutting audio and streams into
segments, and classifying them as speech, music, silence or other sound, and `thumbnailer`,
a CLI for the scenes and thumbnails of a video.

## Motivation

The goal isn't to replace `ffmpeg`, but to provide a programmatic interface for processing
audio and video files from Go. `go-media` supports all the audio and video formats FFmpeg
supports, and its bindings for libheif, LibRaw, libexif and Chromaprint are the basis for
reading HEIF/AVIF and RAW camera images, EXIF metadata, and audio fingerprints in
[go-metadata](https://github.com/mutablelogic/go-metadata). The segmenter prepares audio
for speech recognition and analysis, such as splitting a podcast into parts at its pauses
or near the times its topics change.

## Current Status

This module is **in development** - the API may still change. Implemented so far:

- FFmpeg-backed audio/video: codecs, formats, filters, pixel/sample formats, audio channel
  layouts, demuxing/muxing/remuxing, decoding/encoding, filtering and resampling, hardware
  acceleration (transcoding via the CLI is still TBC)
- Low-level bindings for Chromaprint, libexif, libheif, LibRaw and libfvad (`sys/`)
- Segmenting audio at silence or near given times, and classifying segments as speech,
  music, silence or other sound (`segmenter`, see [segmenter/README.md](segmenter/README.md))

Metadata extraction, including EXIF, XMP, HEIF and RAW images, TMDB and AcoustID lookups,
has moved to [go-metadata](https://github.com/mutablelogic/go-metadata), which uses these
bindings.

## Requirements

Building from source compiles several C/C++ dependencies (FFmpeg, Chromaprint, libexif,
LibRaw, libheif, libfvad) as static libraries before the Go module itself. You'll need a C/C++
toolchain, `pkg-config`, `cmake` and `nasm`, plus the codec libraries you want support for -
FFmpeg's own configure step and libheif's CMake step both auto-detect optional codecs via
`pkg-config`, so anything not installed is simply compiled out rather than causing a build
failure.

**macOS (Homebrew):**

```bash
# Basic dependencies (required)
brew install pkg-config cmake nasm curl freetype lame opus libvorbis libvpx x264 x265 zvbi

# Recommended, for HEIF/AVIF decoding (libde265, aom, dav1d)
brew install libde265 aom dav1d jpeg

# Optional: homebrew-ffmpeg tap for more FFmpeg codecs
brew tap homebrew-ffmpeg/ffmpeg
brew install homebrew-ffmpeg/ffmpeg/ffmpeg \
  --with-fdk-aac --with-libbluray --with-libsoxr --with-libvidstab \
  --with-libvmaf --with-openh264 --with-openjpeg --with-rav1e \
  --with-srt --with-svt-av1 --with-webp --with-xvid --with-zimg
```

**Debian/Ubuntu:**

```bash
apt install pkg-config cmake nasm curl build-essential \
  libfreetype-dev libmp3lame-dev libopus-dev libvorbis-dev libvpx-dev \
  libx264-dev libx265-dev libnuma-dev libzvbi-dev \
  libde265-dev libaom-dev libdav1d-dev
```

**Fedora:**

```bash
dnf install pkg-config cmake nasm curl freetype-devel lame-devel opus-devel \
  libvorbis-devel libvpx-devel x264-devel x265-devel numactl-devel zvbi-devel \
  libde265-devel libaom-devel dav1d-devel
```

**Vulkan Hardware Support (Linux):**

```bash
# Debian/Ubuntu
apt install libvulkan-dev

# Fedora
dnf install vulkan-loader-devel
```

## Building

```bash
git clone https://github.com/mutablelogic/go-media
cd go-media
make              # Builds ffmpeg, chromaprint, libexif, libraw, libheif, libfvad, then all cmd/* binaries
```

Static libraries and pkg-config files land in `build/install`. Once the C dependencies are
built, you can rebuild just the Go side:

```bash
make cmd/gomedia     # Rebuild the gomedia binary only, reusing the already-built C libraries
make cmd/segmenter   # Rebuild the segmenter binary only
```

### Building Manually

If you need to invoke `go build`/`go test` yourself rather than through `make`, set the same
CGO environment variables the Makefile uses:

```bash
export PKG_CONFIG_PATH="${PWD}/build/install/lib/pkgconfig"
export CGO_LDFLAGS_ALLOW="-(W|D).*"
export CGO_LDFLAGS="-lstdc++ -Wl,-no_warn_duplicate_libraries"  # drop the -Wl,... flag on Linux
go build -o build/gomedia ./cmd/gomedia
```

### Testing

```bash
make test           # Build all C deps and run the full Go test suite
make test-sys        # sys/<ffmpeg version> bindings only
make test-ffmpeg      # sys/<ffmpeg version>, reader, writer, frame
make test-chromaprint # sys/chromaprint
make test-exif        # sys/libexif
make test-raw         # sys/libraw
make test-heif        # sys/libheif
make test-libfvad     # sys/libfvad, segmenter
make test-musan       # Classify the MUSAN corpus with the segmenter (downloads 11GB)
make test-gomedia     # task/..., profile/..., gomedia/...
```

## Usage

### Command-Line Tool

```bash
# Run the server
gomedia run

# Probe a media file, or a URL or device, for its container format, streams and metadata
gomedia probe-media <file>
gomedia probe-source <url>

# Encode a media file into a new format
gomedia encode <file>

# List capabilities
gomedia codecs
gomedia formats
gomedia pixelformats
gomedia sampleformats
gomedia channellayouts
gomedia devices

# Manage audio profiles, and tasks
gomedia audio-profile-create
gomedia tasks
gomedia task-events
```

Run `gomedia --help` (or `gomedia <command> --help`) for the full, current set of flags -
this list reflects the commands defined in `gomedia/cmd` and may grow.

### Segmenter

`segmenter` cuts audio into segments at silence, at a fixed duration, or near times you give
it, and classifies each segment as speech, music, silence or other sound. It runs locally,
without a server:

```bash
# Print the segments of a file, and write each one to a folder as an mp3 file
segmenter segment --format mp3 podcast.mp3 parts

# Cut near the times topics change, preferring a pause within a second of each
segmenter segment --segments 3m10s,7m45s,12m20s podcast.mp3 parts

# List the microphones, then show the level and segments of one as you speak
segmenter stream
segmenter stream device://avfoundation/:0

# Classify folders of labelled audio, and summarise how each label was classified
segmenter classify --limit 50 build/musan/speech build/musan/music build/musan/noise
```

The [segmenter package](segmenter/README.md) describes the options, how segments are cut
and classified, and how accurately, and how to use it from Go.

### Thumbnailer

`thumbnailer` finds the scenes of a video, and representative thumbnails of it, from frames
sampled at the video's keyframes, or by seeking. Each frame is scored on its brightness,
contrast, sharpness and colourfulness, and thumbnails are the best frames of scenes which
don't look alike. It writes them as JPEG files:

```bash
# Six thumbnails from different scenes, 640 pixels wide
thumbnailer thumbs --count 6 video.mp4 thumbs

# The one thumbnail which best represents the video
thumbnailer poster --width 320 video.mp4 poster.jpg

# The scenes, and the best frame of each
thumbnailer scenes video.mp4 scenes
```

From Go, `thumbnailer.Open(path)`, or `thumbnailer.NewReader(r)` for a reader which can be
seeked, such as a file, returns a `Thumbnailer`, whose `Analyse`, `Thumbnails`, `Poster` and
`Frame` methods return the scenes and frames.

### Go API - Probing and Media Management

```go
package main

import (
	"context"
	"fmt"
	"os"

	manager "github.com/mutablelogic/go-media/gomedia/manager"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
)

func main() {
	m, err := manager.New(context.Background())
	if err != nil {
		panic(err)
	}

	f, err := os.Open("video.mp4")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	resp, err := m.Probe(context.Background(), schema.ProbeRequest{Reader: f})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Format: %s\n", resp.Format)
	fmt.Printf("Duration: %.3fs\n", resp.Duration)
	for _, stream := range resp.Streams {
		fmt.Printf("Stream %d\n", stream.Id())
	}
}
```

### Low-Level FFmpeg Bindings

For direct FFmpeg access, use `sys/ffmpeg80`, selected via the Makefile's `SYS_VERSION`:

```go
import (
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

filter := ff.AVFilter_get_by_name("scale")
graph := ff.AVFilterGraph_alloc()
// ... etc
```

## Project Structure

```text
sys/ffmpeg80/                                # Low-level CGO FFmpeg bindings
sys/libheif/, sys/libraw/, sys/libexif/      # Low-level CGO bindings for image metadata/codecs
sys/chromaprint/, sys/dvb/, sys/libfvad/     # Other low-level bindings

reader/, writer/, frame/   # Reading and decoding, encoding and writing, and frames
segmenter/           # Segmenting and classifying audio, using reader/ (see its README)
thumbnailer/         # Finding the scenes and thumbnails of a video, using reader/

gomedia/
  manager/           # Orchestrates reader, writer, profiles and tasks
  cmd/                # kong-based CLI command definitions

cmd/gomedia/          # main() entrypoint, wraps gomedia/cmd via go-server's cmd.Main
cmd/segmenter/        # Segments and classifies audio and streams locally (segment, stream, classify)
cmd/thumbnailer/      # Writes the thumbnails and scenes of a video as JPEG files (thumbs, poster, scenes)
```

## Docker

Build a Docker image of `gomedia` and `segmenter` with all dependencies:

```bash
make docker                              # tags ghcr.io/mutablelogic/gomedia:<version>-linux-<arch>
DOCKER_REPO=docker.io/user/gomedia make docker
```

### Library Images

Each release also publishes images of the libraries go-media links with, for amd64 and arm64,
so a project which uses go-media can build on them rather than building FFmpeg and the other
libraries itself. There are two images for each FFmpeg license: `gpl` (with x264, x265 and the
other encoders, as the Makefile builds by default) and `lgpl` (decoders, without the libraries
FFmpeg only allows in a GPL build):

| Image | Contents |
| --- | --- |
| `ghcr.io/mutablelogic/go-media:<version>-<license>-dev` | Debian with the libraries in `/opt/go-media`, their headers and pkg-config files, and the tools to build with them, but no Go |
| `ghcr.io/mutablelogic/go-media:<version>-<license>-runtime` | Debian slim, with only the packages a program linked with the libraries needs to run |

The runtime images are built again each week, for Debian's security updates. A project builds
with its own version of Go on the dev image, and runs on the runtime image:

```dockerfile
ARG GO_MEDIA_VERSION=v1.12.0

FROM golang:1.27-trixie AS go
FROM ghcr.io/mutablelogic/go-media:${GO_MEDIA_VERSION}-lgpl-dev AS builder
COPY --from=go /usr/local/go /usr/local/go
ENV PATH=/usr/local/go/bin:${PATH}
WORKDIR /usr/src/app
COPY . .
RUN go build -o /usr/local/bin/app ./cmd/app

FROM ghcr.io/mutablelogic/go-media:${GO_MEDIA_VERSION}-lgpl-runtime
COPY --from=builder /usr/local/bin/app /usr/local/bin/app
ENTRYPOINT [ "/usr/local/bin/app" ]
```

The dev image sets `PKG_CONFIG_PATH` to `/opt/go-media/lib/pkgconfig`, and the cgo flags
the libraries need. The Debian packages each image installs are listed by
[etc/docker/packages.sh](etc/docker/packages.sh). To build the images locally, for this
machine's architecture:

```bash
make docker-dev FFMPEG_LICENSE=lgpl       # tags ghcr.io/mutablelogic/go-media:<version>-lgpl-dev
make docker-runtime FFMPEG_LICENSE=lgpl   # tags ghcr.io/mutablelogic/go-media:<version>-lgpl-runtime
```

### Library Archives

To build outside a container, such as in CI, each release also has the libraries as archives,
with their headers, pkg-config files and stamps:

| Archive | Built on | What they link with |
| --- | --- | --- |
| `go-media-<version>-<license>-linux-amd64.tar.gz`, `...-linux-arm64.tar.gz` | Debian trixie, from the dev image | The Debian packages in `debian-packages.txt` |
| `go-media-<version>-<license>-darwin-arm64.tar.gz` | macOS with Homebrew | The Homebrew formulae in `Brewfile` |
| `go-media-<version>-<license>-windows-amd64.zip` | MSYS2's MinGW64 | MinGW64's toolchain, `libde265`, `aom` and `dav1d` |

Unpack one into `build/install`, so `make` doesn't build the libraries again, or anywhere else,
as the pkg-config files are relative to where they are. Then install what they link with, and
build with them:

```bash
mkdir -p build/install
tar -xzf go-media-1.12.0-lgpl-darwin-arm64.tar.gz -C build/install
brew bundle --file build/install/Brewfile    # or on Debian: apt install $(cat build/install/debian-packages.txt)

export PKG_CONFIG_PATH="$PWD/build/install/lib/pkgconfig"
export CGO_LDFLAGS_ALLOW="-(W|D).*" CGO_LDFLAGS="-lstdc++"
go build ./...
```

On Linux, the libraries need a glibc as new as Debian trixie's; on another distribution, the
dev image is the better choice. `make libraries-archive` packages the libraries the same way, on
any platform.

For GPU-accelerated encoding/decoding with Vulkan, see [Docker GPU Setup](docs/docker-gpu.md)
for instructions on passing through GPU devices from the host.

## License & Distribution

This software is licensed under the **Apache License 2.0**.

> **go-media**  
> [https://github.com/mutablelogic/go-media/](https://github.com/mutablelogic/go-media/)  
> Copyright (c) 2021-2026 The go-media authors, All rights reserved.

Under the Apache License 2.0, you are free to use, modify, and distribute this software for any purpose, including commercial applications. You may statically or dynamically link to this library without affecting your own code's license. **Attribution requirement:** When redistributing this software or derivative works (in source or binary form), you must include the copyright notice above, and a copy of the [LICENSE](LICENSE) and [NOTICE](NOTICE) files. If you modify the code, you must state your changes. See the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0) for complete terms.

**Important:** When distributing binaries that include this software, you must also comply with the licenses of the libraries it links to, as described below. Their copyright notices and licenses are in the [NOTICE](NOTICE) file.

| Library | License |
| --- | --- |
| [FFmpeg](https://ffmpeg.org/legal.html) | GPL 3.0 or later, as the Makefile builds it (see below); LGPL 2.1 or later by default |
| [libexif](https://libexif.github.io/) | LGPL 2.1 or later |
| [libheif](https://github.com/strukturag/libheif) | LGPL 3.0 |
| [LibRaw](https://www.libraw.org/) | LGPL 2.1 or CDDL 1.0, with bundled code under BSD and MIT licenses |
| [Chromaprint](https://acoustid.org/chromaprint) | LGPL 2.1, as it includes parts of FFmpeg; its own code is MIT |
| [libfvad](https://github.com/dpirch/libfvad) | BSD 3-clause, with an additional grant of patent rights from Google |

### FFmpeg as built by the Makefile

The Makefile configures FFmpeg with `--enable-gpl` and `--enable-version3`, for x264, x265 and
OpenSSL 3, which FFmpeg's configure reports as **"GPL version 3 or later"** (it's
`FFMPEG_LICENSE` in `build/ffmpeg-<version>/config.h`). A binary which includes it is licensed
under the [GNU GPL v3](https://www.gnu.org/licenses/gpl-3.0.html), so distributing it means
providing its complete source code under the GPL v3. go-media's own Apache 2.0 code is
compatible with that.

`make FFMPEG_LICENSE=lgpl` builds FFmpeg without `--enable-gpl`, and without the libraries
FFmpeg only allows in a GPL build, such as x264 and x265, which FFmpeg reports as
**"LGPL version 3 or later"**. A binary which includes it is licensed under the
[GNU LGPL v3](https://www.gnu.org/licenses/lgpl-3.0.html), as the `lgpl` library images are.

`make FFMPEG_NONFREE=1` also enables the nonfree components, fdk-aac and libtls, when they're
found. FFmpeg reports such a build as **"nonfree and unredistributable"**, so a binary built
with it is for your own use only.

### LGPL Notice

This software also statically links to libexif, libheif, LibRaw and Chromaprint,
which are licensed under the LGPL: [v2.1](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html), except
libheif, which is [v3](https://www.gnu.org/licenses/lgpl-3.0.html) (LibRaw is also available under the CDDL).

**LGPL Compliance:** Under the LGPL, you may:

- Use this software for commercial or non-commercial purposes
- Distribute this software in its compiled form
- Modify the go-media source code under Apache 2.0

**Requirements when distributing binaries:**

1. Include the [NOTICE](NOTICE) file, the LGPL licenses, and a copy of the
   [GNU GPL v3](https://www.gnu.org/licenses/gpl-3.0.html), which LGPL v3 (libheif) requires
   alongside it
2. Provide access to the source code of the libraries used (available in `build/` after building, such as `build/ffmpeg-8.1.3/` - check the Makefile's `*_VERSION` variables for the exact versions), and, as FFmpeg is GPL, of the whole binary
3. Allow users to relink the application with modified versions of the libraries

The source code of the libraries is automatically downloaded during the build process. See `Makefile` for details.

### Contributing

Please file feature requests and bugs at [github.com/mutablelogic/go-media/issues](https://github.com/mutablelogic/go-media/issues). Pull Requests are welcome, after discussion of the proposed changes.

## References

- [FFmpeg Documentation](https://ffmpeg.org/documentation.html)
- [libheif](https://github.com/strukturag/libheif)
- [LibRaw](https://www.libraw.org/)
- [libexif](https://libexif.github.io/)
- [libfvad](https://github.com/dpirch/libfvad), the WebRTC voice activity detector
- [MUSAN](https://www.openslr.org/17/), the corpus of speech, music and noise the segmenter is tuned with
- [Go Package Documentation](https://pkg.go.dev/github.com/mutablelogic/go-media)
- [AcoustID API](https://acoustid.org/webservice)
