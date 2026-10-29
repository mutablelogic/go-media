# go-media AI Agent Instructions

## Architecture Overview

This Go module provides CGO bindings to several media C libraries, layered consistently:

1. **sys/\*** - low-level CGO bindings, mirroring each C library's API 1:1
   - `sys/ffmpeg80/` - FFmpeg bindings, selected via `SYS_VERSION` in the Makefile
   - `sys/libheif/`, `sys/libraw/`, `sys/libexif/`, `sys/libfvad/`, `sys/chromaprint/`, `sys/dvb/`
2. **High-level packages** - idiomatic Go APIs over `sys/ffmpeg80`
   - `reader/` - demuxing and decoding from a file, URL, device or `io.Reader` (`Reader`, `Decoder`)
   - `writer/` - encoding and muxing (`Writer`, `Encoder`), and `Resampler`, which converts audio and video frames to a target format
   - `frame/` - `AudioFrame`, `VideoFrame` and `SubtitleFrame`, wrapping `AVFrame`/`AVSubtitle`
   - `profile/` and `task/` - encoding profiles, and asynchronous tasks such as encoding, each with `schema/`, `manager/`, `httphandler/` and `httpclient/` packages
   - `segmenter/` - segments audio at silence, a fixed duration or given times, and classifies segments as speech, music, silence or other sound, using `reader/`; see its README
3. **gomedia/** - the application layer that the CLI/server binary is built from
   - `gomedia/manager/` - orchestrates `reader/`, `writer/`, profiles and tasks (probing, encoding, capabilities)
   - `gomedia/cmd/` - `kong`-based CLI command definitions, wrapped by `cmd/gomedia`
4. **cmd/** - thin `main()`s built with `go-server`'s `cmd.Main(...)`: `cmd/gomedia` (CLI, client and server) and `cmd/segmenter` (local `segment`, `stream` and `classify` commands)

Older code, including the previous high-level FFmpeg API (`pkg/ffmpeg`) and the SDL2 player (`pkg/sdl`), is archived under `_old/`, which Go ignores. Don't assume any of it is built or tested.

## Critical Build Workflow

**First build ALWAYS requires the full `make`** - it builds several C dependencies as static libraries before compiling Go:
```bash
make              # Builds ffmpeg, chromaprint, libexif, libraw, libheif, libfvad, then all cmd/* binaries
```

Source versions are pinned in the Makefile (check there for current values - they change):
```
FFMPEG_VERSION, CHROMAPRINT_VERSION, LIBEXIF_VERSION, LIBRAW_VERSION, LIBHEIF_VERSION
```

**Subsequent builds** can target a single command once the C deps are already installed under `build/install`:
```bash
make cmd/gomedia  # Rebuilds Go only, IF ffmpeg/chromaprint/libexif/libraw/libheif/libfvad are already built
```

The Makefile sets CGO env vars for every build/test target (`CGO_ENV`), and **these differ by OS**:
```bash
# common to both
PKG_CONFIG_PATH="$(realpath build/install)/lib/pkgconfig"
CGO_LDFLAGS_ALLOW="-(W|D).*"
# darwin adds -Wl,-no_warn_duplicate_libraries to CGO_LDFLAGS; linux does not
CGO_LDFLAGS="-lstdc++ [-Wl,-no_warn_duplicate_libraries on darwin]"
```
Never run `go build`/`go test` directly without these vars (see Makefile's `CGO_ENV`) or you'll get pkg-config/link errors. If you must run `go test` outside `make`, set `PKG_CONFIG_PATH` to `$(realpath build/install)/lib/pkgconfig` at minimum.

### libheif codec selection is dependency-driven, not hardcoded
`libheif-dep` in the Makefile passes `-DENABLE_PLUGIN_LOADING=OFF` (codecs are compiled in, not dynamically loaded) and sets each `-DWITH_<CODEC>=ON/OFF` flag by checking `pkg-config --exists <lib>` on the *build machine* (`libde265`, `x265`, `aom`, `dav1d`, `libjpeg`, ffmpeg libs, etc.). If a codec's `-dev` package isn't installed when `make libheif` configures, that codec is silently compiled out - libheif will then fail to decode files needing it with `HEIF_ERROR_DECODER_PLUGIN_ERROR` / `HEIF_SUBERROR_UNSPECIFIED`, which looks like a runtime plugin-loading problem but is actually a build-time configuration gap. Since libheif itself builds `-DBUILD_SHARED_LIBS=0` (static), there's no `.so`/`LD_LIBRARY_PATH`/plugin-path angle to chase here - if HEIF decode tests fail, check which `-dev` packages were present on the machine that ran `make libheif`, not runtime linking.

## FFmpeg Integration Patterns

### Frame Buffer Management
**Critical**: FFmpeg reuses frame buffers across decode operations. A `reader.Decoder` closes each frame after its `FrameFn` returns, and `writer.Resampler` reuses its output frame, so copy any samples or pixels you need to keep before the callback returns.

### Stream Decoder Initialization
`reader.Decoder` only decodes the streams registered with `Add`, and fails to open if a registered stream has no decoder (`reader/decoder.go`), so check a stream's codec parameters before adding it.

### Reader Input Options
Per-format demuxer tuning, and devices, are set with `reader.WithInput`:

```go
r, err := reader.Open(url, reader.WithInput("s16le", "sample_rate=16000", "channels=1"))
r, err := reader.Open(":0", reader.WithInput("avfoundation"))  // the first audio device on macOS
```

There is currently no automatic per-extension heuristic (e.g. auto-tuning MPEG-TS probing) - callers must pass the input format and options explicitly if a demuxer needs help.

## Testing Patterns

### Debugging
- FFmpeg's own logging is controlled with `AVUtil_log_set_level` and `AVUtil_log_set_callback` in `sys/ffmpeg80`.

### Test Fixtures
- `etc/test/sample.mp4` - h264/aac, 1280x720, ~5.3s
- `etc/test/photo.HEIC` - used by the `sys/libheif` tests
- RAW fixture(s) under `etc/test/` for the `sys/libraw` tests (e.g. an Olympus `.ORF`)
- There are currently no MPEG-TS fixtures or `dvb_teletext` handling in this tree - don't assume that gotcha still applies.

## Common Gotchas

1. **Stream JSON Marshaling**: `AVStream.MarshalJSON` delegates to the C struct and produces garbage. Use manual field extraction via `Id()`, `CodecPar()`, `TimeBase()`, etc. (see `profile/schema/stream.go`).
2. **AVStream API naming**: methods are `Id()` not `ID()`; use `AVCodecID.Name()`, not `AVCodec_get_name()`.
3. **AV_NOPTS_VALUE**: cast to `int64` for comparisons: `pts != int64(ff.AV_NOPTS_VALUE)`.
4. **Audio Channel Layouts**: modern FFmpeg uses the `AVChannelLayout` struct, not legacy channel-mask integers.
5. **libraw timestamps are host-timezone-poisoned**: libraw computes `imgother.timestamp` via C's `mktime()` on the EXIF wall-clock string, which has no real offset - `mktime()` interprets it *as local time on whatever machine parsed the file*. `time.Unix(ts, 0)` (in `time.Local`) is the correct inverse and recovers the original wall-clock digits on any host, but converting to another zone (e.g. calling `.UTC()` on it) shifts those digits by the parsing host's own UTC offset - which varies by machine/CI runner and is not the photo's real capture offset. Recover the wall-clock fields via `time.Local` and *relabel* them as UTC (no arithmetic conversion), so results are deterministic across hosts, as go-metadata's RAW handler does.
6. **HEIF "Decoder plugin generated an error: Unspecified"**: almost always means libheif was built with a codec (e.g. `libde265` for HEVC/HEIC) compiled out because its `-dev` package wasn't present when `make libheif` ran pkg-config detection - see "libheif codec selection" above. It's not a runtime plugin-path or shared-library-resolution issue; libheif is built statically and plugin loading is disabled.

## File Conventions

- `schema/` packages (such as `profile/schema/` and `task/schema/`) define request/response types
- FFmpeg C types are prefixed with `AV` (`AVCodec`, `AVFormatContext`, `AVFrame`); Go wrappers omit the prefix (`reader.Reader`, `reader.Decoder`, `frame.AudioFrame` wraps `AVFrame`)
- Metadata keys use a `namespace:name` convention (e.g. `exif:DateTimeOriginal`, `tiff:Make`, `artwork:thumbnail`) - see `gomedia.Metadata.Key()`
