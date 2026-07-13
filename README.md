# go-media

[![Test](https://github.com/mutablelogic/go-media/actions/workflows/on_pull_request_merge.yaml/badge.svg)](https://github.com/mutablelogic/go-media/actions/workflows/on_pull_request_merge.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mutablelogic/go-media.svg)](https://pkg.go.dev/github.com/mutablelogic/go-media)

`gomedia` is a CLI, client and server tool for managing media files - audio, video and image - extracting metadata, artwork and thumbnails, identifying music, and remultiplexing and transcoding audio and video.

## Motivation

The goal isn't to replace `ffmpeg`, but to provide a programmatic interface for processing
and extracting information from audio, video and image files. `go-media` supports all the
audio and video formats FFmpeg supports, plus HEIF/AVIF and RAW camera images, including
extraction of embedded EXIF and XMP metadata.

## Current Status

This module is **in development** - the API may still change. Implemented so far:

- FFmpeg-backed audio/video: codecs, formats, filters, pixel/sample formats, audio channel
  layouts, demuxing/muxing/remuxing, decoding/encoding, filtering and resampling, hardware
  acceleration (transcoding via the CLI is still TBC)
- HEIF/AVIF image decoding (`pkg/heif`), registered with Go's standard `image` package
- RAW camera image decoding across many manufacturers (`pkg/raw`), also registered with `image`
- EXIF (`pkg/exif`) and XMP (`pkg/xmp`) metadata reading
- A format-agnostic metadata extraction registry (`metadata/`) spanning image, audio, video
  and application (e.g. Photoshop) content types
- Audio fingerprinting and identification via Chromaprint/AcoustID (`pkg/chromaprint`)

## Requirements

Building from source compiles several C/C++ dependencies (FFmpeg, Chromaprint, libexif,
LibRaw, libheif) as static libraries before the Go module itself. You'll need a C/C++
toolchain, `pkg-config`, `cmake` and `nasm`, plus the codec libraries you want support for -
FFmpeg's own configure step and libheif's CMake step both auto-detect optional codecs via
`pkg-config`, so anything not installed is simply compiled out rather than causing a build
failure.

**macOS (Homebrew):**

```bash
# Basic dependencies (required)
brew install pkg-config cmake nasm curl freetype lame opus libvorbis libvpx x264 x265

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
make              # Builds ffmpeg, chromaprint, libexif, libraw, libheif, then all cmd/* binaries
```

Static libraries and pkg-config files land in `build/install`. Once the C dependencies are
built, you can rebuild just the Go side:

```bash
make cmd/gomedia   # Rebuild the gomedia binary only, reusing the already-built C libraries
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
make test-ffmpeg      # pkg/ffmpeg/...
make test-chromaprint # pkg/segmenter, pkg/chromaprint
make test-exif        # sys/libexif, pkg/exif
make test-raw         # sys/libraw, pkg/raw
make test-heif        # sys/libheif, pkg/heif
make test-metadata    # metadata/...
make test-gomedia     # gomedia/...
```

## Usage

### Command-Line Tool

```bash
# Extract metadata from a file (all namespaces, or one via --namespace)
gomedia metadata <file>
gomedia metadata --namespace exif <file>

# Extract embedded artwork/thumbnails
gomedia artwork <file>

# Probe a media file's container and streams
gomedia probe <file>

# List capabilities
gomedia codecs
gomedia filters
gomedia formats
gomedia pixel-formats
gomedia sample-formats
gomedia audio-channels

# Segment audio (fixed-size and/or silence-based)
gomedia audio-segment <file> --out ./segments

# Audio fingerprinting and AcoustID lookup (built with the chromaprint tag; requires an API key)
export CHROMAPRINT_KEY=<your-key>
gomedia audio-fingerprint <file>
gomedia audio-lookup <fingerprint> <duration>
```

Run `gomedia --help` (or `gomedia <command> --help`) for the full, current set of flags -
this list reflects the commands defined in `gomedia/cmd` and may grow.

### Go API - Metadata Extraction

```go
package main

import (
	"context"
	"fmt"
	"os"

	metadata "github.com/mutablelogic/go-media/metadata"
	_ "github.com/mutablelogic/go-media/metadata/image" // registers HEIF/RAW/JPEG/... handlers
)

func main() {
	f, err := os.Open("photo.HEIC")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	contentType, _, err := metadata.ContentType(f)
	if err != nil {
		panic(err)
	}

	meta, err := metadata.GetMetadata(context.Background(), f, contentType, "")
	if err != nil {
		panic(err) // treat as a warning: meta may still contain results from other handlers
	}
	for _, m := range meta {
		fmt.Printf("%s = %s\n", m.Key(), m.Value())
	}
}
```

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

For direct FFmpeg access, use `sys/ffmpeg80` (the default; `sys/ffmpeg71` and `sys/ffmpeg61`
are also available, selected via the Makefile's `SYS_VERSION`):

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
sys/ffmpeg80/, sys/ffmpeg71/, sys/ffmpeg61/  # Low-level CGO FFmpeg bindings
sys/libheif/, sys/libraw/, sys/libexif/      # Low-level CGO bindings for image metadata/codecs
sys/chromaprint/, sys/dvb/                   # Other low-level bindings

pkg/ffmpeg/          # High-level FFmpeg API (Reader, Decoder, Encoder, Resampler, Frame)
pkg/heif/            # HEIF/AVIF decoding, registered with the stdlib image package
pkg/raw/             # RAW camera image decoding, registered with the stdlib image package
pkg/exif/            # EXIF metadata reading
pkg/xmp/             # XMP document read/write
pkg/sdl/             # SDL2 video/audio player (library only; not wired into the gomedia CLI)
pkg/chromaprint/     # Audio fingerprinting

metadata/            # Format-agnostic metadata extraction registry
  image/, audio/, video/, application/  # Per-kind handlers, registered via metadata.AddHandler

gomedia/
  schema/            # Request/response types for the CLI/API surface
  manager/           # Orchestrates pkg/ffmpeg, metadata/, pkg/chromaprint, pkg/xmp
  cmd/                # kong-based CLI command definitions

cmd/gomedia/          # main() entrypoint, wraps gomedia/cmd via go-server's cmd.Main
```

## Docker

Build a Docker image with all dependencies:

```bash
make docker                              # tags ghcr.io/mutablelogic/gomedia:<version>-<os>-<arch>
DOCKER_REPO=docker.io/user/gomedia make docker
```

For GPU-accelerated encoding/decoding with Vulkan, see [Docker GPU Setup](docs/docker-gpu.md)
for instructions on passing through GPU devices from the host.

## License & Distribution

This software is licensed under the **Apache License 2.0**.

> **go-media**  
> [https://github.com/mutablelogic/go-media/](https://github.com/mutablelogic/go-media/)  
> Copyright (c) 2021-2026 David Thorpe, All rights reserved.

Under the Apache License 2.0, you are free to use, modify, and distribute this software for any purpose, including commercial applications. You may statically or dynamically link to this library without affecting your own code's license. **Attribution requirement:** When redistributing this software or derivative works (in source or binary form), you must include the copyright notice above and a copy of the [LICENSE](LICENSE) file. If you modify the code, you must state your changes. See the [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0) for complete terms.

**Important:** When distributing binaries that include this software, you must also comply with the licenses of the linked libraries (FFmpeg LGPL and Chromaprint MIT) as described below.

### FFmpeg LGPL Notice

This software statically links to [FFmpeg](http://ffmpeg.org/) libraries, which are licensed under the
[GNU Lesser General Public License (LGPL) v2.1](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html).

**LGPL Compliance:** Under the LGPL, you may:

- Use this software for commercial or non-commercial purposes
- Distribute this software in its compiled form
- Modify the go-media source code under Apache 2.0

**Requirements when distributing binaries:**

1. Include this notice and the FFmpeg LGPL license
2. Provide access to the FFmpeg source code used (available at `build/ffmpeg-8.0.3/` after building - check the Makefile's `FFMPEG_VERSION` for the exact version)
3. Allow users to relink the application with modified FFmpeg libraries

The FFmpeg source code is automatically downloaded during the build process. See `Makefile` for details.

### Contributing

Please file feature requests and bugs at [github.com/mutablelogic/go-media/issues](https://github.com/mutablelogic/go-media/issues). Pull Requests are welcome, after discussion of the proposed changes.

## References

- [FFmpeg Documentation](https://ffmpeg.org/documentation.html)
- [libheif](https://github.com/strukturag/libheif)
- [LibRaw](https://www.libraw.org/)
- [libexif](https://libexif.github.io/)
- [Go Package Documentation](https://pkg.go.dev/github.com/mutablelogic/go-media)
- [AcoustID API](https://acoustid.org/webservice)
