
# go-media

This module provides Go bindings and utilities for FFmpeg, including:

* Low-level CGO bindings for [FFmpeg 8.0](https://ffmpeg.org/) (in `sys/ffmpeg80`)
* High-level Go API for media operations (in `pkg/ffmpeg`)
* Task manager for common media operations (in `pkg/ffmpeg/task`)
* Audio fingerprinting with Chromaprint/AcoustID (in `pkg/chromaprint`)
* Command-line tool `gomedia` for media inspection and manipulation
* HTTP server for media services

## Current Status

This module is in development (meaning the API might change).
FFmpeg 8.0 bindings are complete with support for:

* Codecs, formats, filters, pixel formats, sample formats
* Audio channel layouts
* Demuxing, muxing, remuxing
* Audio and video decoding/encoding (transcoding is TBC)
* Filtering and resampling
* Metadata extraction
* Hardware acceleration support

## Requirements

### Building FFmpeg

The module includes scripts to build FFmpeg with common codecs. Install dependencies first:

**macOS (Homebrew):**

```bash
# Basic dependencies (required)
brew install pkg-config freetype lame opus libvorbis libvpx x264 x265

# Optional: Install homebrew-ffmpeg tap for more codecs
brew tap homebrew-ffmpeg/ffmpeg
brew install homebrew-ffmpeg/ffmpeg/ffmpeg \
  --with-fdk-aac --with-libbluray --with-libsoxr --with-libvidstab \
  --with-libvmaf --with-openh264 --with-openjpeg --with-rav1e \
  --with-srt --with-svt-av1 --with-webp --with-xvid --with-zimg
```

**Debian/Ubuntu:**

```bash
apt install pkg-config libfreetype-dev libmp3lame-dev libopus-dev \
  libvorbis-dev libvpx-dev libx264-dev libx265-dev libnuma-dev libzvbi-dev
```

**Fedora:**

```bash
dnf install pkg-config freetype-devel lame-devel opus-devel \
  libvorbis-devel libvpx-devel x264-devel x265-devel numactl-devel zvbi-devel
```

**Vulkan Hardware Support (Linux):**

For Vulkan-based hardware acceleration on Linux, install:

```bash
# Debian/Ubuntu
apt install libvulkan-dev

# Fedora
dnf install vulkan-loader-devel
```

Then build FFmpeg:

```bash
git clone https://github.com/mutablelogic/go-media
cd go-media
make ffmpeg chromaprint
```

This creates static libraries in `build/install` with the necessary pkg-config files.

### Building the Go Module

The module uses CGO and requires the FFmpeg libraries. The Makefile handles the necessary
environment variables:

```bash
make              # Build the gomedia command-line tool
make test         # Run all tests
make test-sys     # Run system/FFmpeg binding tests only
```

To build manually:

```bash
export PKG_CONFIG_PATH="${PWD}/build/install/lib/pkgconfig"
export CGO_LDFLAGS_ALLOW="-(W|D).*"
export CGO_LDFLAGS="-lstdc++ -Wl,-no_warn_duplicate_libraries"
go build -o build/gomedia ./cmd/gomedia
```

## Usage

### Command-Line Tool

The `gomedia` tool provides various media operations:

```bash
# List available codecs
gomedia list-codecs

# List available filters  
gomedia list-filters

# List supported formats
gomedia list-formats

# List pixel/sample formats
gomedia list-pixel-formats
gomedia list-sample-formats

# Probe a media file
gomedia probe <file>

# Remux a file (change container without re-encoding)
gomedia remux --input <input> --output <output>

# Audio fingerprinting and lookup (requires AcoustID API key)
export CHROMAPRINT_KEY=<your-key>
gomedia audio-lookup <file>

# Run HTTP server
gomedia server run
```

### Go API - Task Manager

The task manager provides a high-level API for media operations:

```go
package main

import (
 "context"
 "fmt"
 
 task "github.com/mutablelogic/go-media/pkg/ffmpeg/task"
 schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

func main() {
 // Create a task manager
 manager, err := task.NewManager()
 if err != nil {
  panic(err)
 }

 // List all video codecs
 codecs, err := manager.ListCodecs(context.Background(), &schema.ListCodecRequest{
  Type: "video",
 })
 if err != nil {
  panic(err)
 }
 
 for _, codec := range codecs {
  fmt.Printf("%s: %s\n", codec.Name, codec.LongName)
 }

 // Probe a media file
 info, err := manager.Probe(context.Background(), &schema.ProbeRequest{
  Input: "video.mp4",
 })
 if err != nil {
  panic(err)
 }
 
 fmt.Printf("Format: %s\n", info.Format)
 fmt.Printf("Duration: %v\n", info.Duration)
 for _, stream := range info.Streams {
  fmt.Printf("Stream %d: %s\n", stream.Index, stream.Type)
 }
}
```

### Available Task Manager Methods

The task manager (`pkg/ffmpeg/task.Manager`) provides these methods:

**Query Operations:**

* `ListCodecs(ctx, *ListCodecRequest) (ListCodecResponse, error)` - List available codecs
* `ListFilters(ctx, *ListFilterRequest) (ListFilterResponse, error)` - List available filters
* `ListFormats(ctx, *ListFormatRequest) (ListFormatResponse, error)` - List formats and devices
* `ListPixelFormats(ctx, *ListPixelFormatRequest) (ListPixelFormatResponse, error)` - List pixel formats
* `ListSampleFormats(ctx, *ListSampleFormatRequest) (ListSampleFormatResponse, error)` - List sample formats
* `ListAudioChannelLayouts(ctx, *ListAudioChannelLayoutRequest) (ListAudioChannelLayoutResponse, error)` - List audio layouts

**Media Operations:**

* `Probe(ctx, *ProbeRequest) (*ProbeResponse, error)` - Inspect media files
* `Remux(ctx, *RemuxRequest) error` - Remux media without re-encoding
* `AudioFingerprint(ctx, *AudioFingerprintRequest) (*AudioFingerprintResponse, error)` - Generate fingerprints and lookup

### Low-Level FFmpeg Bindings

For direct FFmpeg access, use the `sys/ffmpeg80` package:

```go
import (
 ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

// Low-level FFmpeg operations
filter := ff.AVFilter_get_by_name("scale")
graph := ff.AVFilterGraph_alloc()
// ... etc
```

### HTTP Server

The module includes an HTTP server exposing the task manager via REST API:

```bash
# Start server
gomedia server run --url http://localhost:8080/api

# Query endpoints
curl http://localhost:8080/api/codec
curl http://localhost:8080/api/filter?name=scale
curl http://localhost:8080/api/format?type=muxer
```

### Audio Fingerprinting

```go
import (
 chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
)

// Requires CHROMAPRINT_KEY environment variable or explicit API key
client, err := chromaprint.NewClient(apiKey)
if err != nil {
 panic(err)
}

// Generate fingerprint and lookup
result, err := client.Lookup(context.Background(), "audio.mp3")
if err != nil {
 panic(err)
}

fmt.Printf("Title: %s\n", result.Title)
fmt.Printf("Artist: %s\n", result.Artist)
```

## Docker

Build a Docker image with all dependencies:

```bash
DOCKER_REGISTRY=docker.io/user make docker
```

## Project Structure

```
sys/ffmpeg80/          # Low-level CGO FFmpeg bindings
pkg/ffmpeg/            # High-level Go API
  task/                # Task manager for common operations
  schema/              # Request/response schemas
  httphandler/         # HTTP handlers
pkg/chromaprint/       # Audio fingerprinting
cmd/gomedia/           # Command-line tool
```

## Contributing & Distribution

Please file feature requests and bugs at [github.com/mutablelogic/go-media/issues](https://github.com/mutablelogic/go-media/issues).

Licensed under Apache 2.0. Redistributions must include copyright notice.

> **go-media**  
> [https://github.com/mutablelogic/go-media/](https://github.com/mutablelogic/go-media/)  
> Copyright (c) 2021-2026 David Thorpe, All rights reserved.

This software links to [FFmpeg](http://ffmpeg.org/) libraries licensed under the
[LGPLv2.1](http://www.gnu.org/licenses/old-licenses/lgpl-2.1.html).

## References

* [FFmpeg 8.0 Documentation](https://ffmpeg.org/doxygen/8.0/index.html)
* [Go Package Documentation](https://pkg.go.dev/github.com/mutablelogic/go-media)
* [AcoustID API](https://acoustid.org/webservice)
