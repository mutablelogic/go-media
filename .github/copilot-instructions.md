# go-media AI Agent Instructions

## Architecture Overview

This Go module provides CGO bindings to several media C libraries, layered consistently:

1. **sys/\*** - low-level CGO bindings, mirroring each C library's API 1:1
   - `sys/ffmpeg80/` (default), plus `sys/ffmpeg71/`, `sys/ffmpeg61/` - FFmpeg bindings, selected via `SYS_VERSION` in the Makefile
   - `sys/libheif/`, `sys/libraw/`, `sys/libexif/`, `sys/chromaprint/`, `sys/dvb/`
2. **pkg/\*** - high-level, idiomatic Go APIs per library
   - `pkg/ffmpeg/` - Reader, Decoder, Encoder, Resampler, Frame abstractions
   - `pkg/heif/`, `pkg/raw/`, `pkg/exif/` - `Open`/`Read`/`Parse` returning a closable value; RAW and HEIF also implement `image.Image` and register with the stdlib `image` package
   - `pkg/xmp/` - XMP document read/write, used to serialize extracted metadata
   - `pkg/sdl/` - SDL2 video/audio player (Context, FrameLoop, Player). **Not currently wired into the `gomedia` CLI** - it's an independent, tested library; treat it as such rather than assuming there's a `play` command.
   - `pkg/chromaprint/` - audio fingerprinting
3. **metadata/** - format-agnostic metadata extraction, independent of `pkg/ffmpeg`
   - `metadata/metadata.go` defines `AddHandler(re *regexp.Regexp, fn HandlerFunc, namespaces ...string)`; each format package (`metadata/image`, `metadata/audio`, `metadata/video`, `metadata/application`) registers itself in an `init()` against a content-type regex and the metadata namespaces it can produce (`tiff`, `exif`, `image`, `dc`, `artwork`, ...)
   - `metadata.GetMetadata(ctx, r, contentType, filter)` runs every matching handler concurrently and merges results; a returned error is the *joined* errors of failing handlers and should be treated as a warning, not a reason to discard metadata that did come back
   - `filter` follows a `namespace:name` convention, e.g. `"exif:"` (whole namespace), `"DateTimeOriginal"` (bare name, any namespace), `"exif:DateTimeOriginal"` (exact), `"artwork:thumbnail"` (embedded preview image)
4. **gomedia/** - the application layer that the CLI/server binary is built from
   - `gomedia/schema/` - request/response types with `kong` CLI tags (the public-facing API surface; distinct from the lower-level `pkg/ffmpeg/schema/` types used internally)
   - `gomedia/manager/` - `manager.Media` orchestrates `pkg/ffmpeg`, `metadata/`, `pkg/chromaprint`, `pkg/xmp` behind the `gomedia/schema` types (`Probe`, `GetMetadata`, `SegmentAudio`, `ListCodecs`, etc.)
   - `gomedia/cmd/` - `kong`-based CLI command definitions (`CLICommands` embeds `MetadataCLICommands`, `CapabilitiesCLICommands`, `EncodingCLICommands`); each `Run` method calls `c.WithManager(ctx, func(m *manager.Media) error {...})`
5. **cmd/gomedia/** - thin `main()` that wraps `gomedia/cmd.CLI{}` via `go-server`'s `cmd.Main(...)`. There is no bespoke input-format-sniffing or SDL-wiring logic here anymore - look in `gomedia/manager` and `gomedia/cmd` instead.

An older Task Manager pattern under `pkg/ffmpeg/task/` and an SDL-playback `main.go` are referenced by some stale docs/README content but no longer exist on this branch (that code now lives archived under `_old/`, if present). Don't assume `pkg/ffmpeg/task` exists - the equivalent responsibility now lives in `gomedia/manager`.

## Critical Build Workflow

**First build ALWAYS requires the full `make`** - it builds several C dependencies as static libraries before compiling Go:
```bash
make              # Builds ffmpeg, chromaprint, libexif, libraw, libheif, then all cmd/* binaries
```

Source versions are pinned in the Makefile (check there for current values - they change):
```
FFMPEG_VERSION, CHROMAPRINT_VERSION, LIBEXIF_VERSION, LIBRAW_VERSION, LIBHEIF_VERSION
```

**Subsequent builds** can target a single command once the C deps are already installed under `build/install`:
```bash
make cmd/gomedia  # Rebuilds Go only, IF ffmpeg/chromaprint/libexif/libraw/libheif are already built
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
**Critical**: FFmpeg reuses frame buffers across decode operations. When queueing frames for another goroutine (e.g. SDL playback), ALWAYS copy first:

```go
// pkg/sdl/playback.go, FrameWriter.WriteFrame
copy, err := frame.Copy()  // pkg/ffmpeg/frame.go - allocates a new buffer
if err != nil {
    return err
}
frameCh <- copy  // Safe to pass to another goroutine
```

### Stream Decoder Initialization
Handle unsupported codecs gracefully rather than failing the whole decode (`pkg/ffmpeg/decoder.go`):

```go
codec := ff.AVCodec_find_decoder(codecID)
if codec == nil {
    return nil, errors.Join(ErrCodecNotFound, errors.New("codec: "+codecID.Name()))
}

// mapStreams() skips streams whose decoder init failed with ErrCodecNotFound
// rather than aborting the whole file
if errors.Is(err, ErrCodecNotFound) {
    continue
}
```

### Reader Input Options
Per-format demuxer tuning flows from the CLI/manager layer down into `pkg/ffmpeg`:

```go
// gomedia/schema/probe.go (CLI-facing) - InputFormat/InputOpts also exist on
// pkg/ffmpeg/schema.Request for internal callers
InputFormat string   `kong:"name='input-format',help='Force input format'"`
InputOpts   []string `kong:"name='input-opt',help='Input format options (key=value)'"`

// pkg/ffmpeg/opts.go
reader, err := ffmpeg.NewReader(r, ffmpeg.WithInput(req.InputFormat, req.InputOpts...))
```

There is currently no automatic per-extension heuristic (e.g. auto-tuning MPEG-TS probing) in `cmd/gomedia` - callers must pass `InputFormat`/`InputOpts` explicitly if a demuxer needs help.

## SDL Player Architecture (library-only; not wired into the CLI)

1. **Context** - initializes SDL, manages the event loop via goroutine registration
2. **FrameLoop** - runs on the main thread, dispatches frames from a channel via user events: `sdl.NewFrameLoop(ctx, handlerFunc, bufferSize, opts...)`
3. **Player** - created via `ctx.NewPlayer()` (a method on `*Context`, not a package-level constructor); manages Window and Audio, handles format conversion

**SDL MUST run on the main thread on macOS.**

### Video Timing
`pkg/sdl/player.go`'s `VideoDelay` is PTS-based, not a fixed interval - it tracks `lastPTS`/a frame timer and compares against `frame.Ts()` (skipping non-video frames via `frame.Type() != media.VIDEO`). Don't reintroduce a fixed 33ms sleep.

### Format Conversion
Player auto-converts video formats via a cached resampler: YUV420P/RGB24 go straight to a native SDL texture update; everything else goes through swscale to YUV420P first.

## Testing Patterns

### Debugging
- `GOMEDIA_SDL_DEBUG=1` enables detailed SDL operation logging (`pkg/sdl/player.go`).
- `pkg/ffmpeg/logging.go` exposes `SetLogging(verbose bool, fn LogFn)` - use that instead of calling `AVUtil_log_set_callback`/`AVUtil_log_set_level` directly.

### Test Fixtures
- `etc/test/sample.mp4` - h264/aac, 1280x720, ~5.3s
- `etc/test/photo.HEIC` - used by `pkg/heif`/`metadata/image` HEIF tests
- RAW fixture(s) under `etc/test/` for `pkg/raw`/`metadata/image` RAW tests (e.g. an Olympus `.ORF`)
- There are currently no MPEG-TS fixtures or `dvb_teletext` handling in this tree - don't assume that gotcha still applies.

## Common Gotchas

1. **Stream JSON Marshaling**: `AVStream.MarshalJSON` delegates to the C struct and produces garbage. Use manual field extraction via `Id()`, `CodecPar()`, `TimeBase()`, etc. (see `pkg/ffmpeg/schema/stream.go`).
2. **AVStream API naming**: methods are `Id()` not `ID()`; use `AVCodecID.Name()`, not `AVCodec_get_name()`.
3. **AV_NOPTS_VALUE**: cast to `int64` for comparisons: `pts != int64(ff.AV_NOPTS_VALUE)`.
4. **Texture Recreation**: SDL textures must be recreated when format/dimensions change - see `pkg/sdl/window.go`'s `ensureTexture()`.
5. **Audio Channel Layouts**: modern FFmpeg uses the `AVChannelLayout` struct, not legacy channel-mask integers.
6. **libraw timestamps are host-timezone-poisoned**: libraw computes `imgother.timestamp` via C's `mktime()` on the EXIF wall-clock string, which has no real offset - `mktime()` interprets it *as local time on whatever machine parsed the file*. `time.Unix(ts, 0)` (in `time.Local`) is the correct inverse and recovers the original wall-clock digits on any host, but converting to another zone (e.g. calling `.UTC()` on it) shifts those digits by the parsing host's own UTC offset - which varies by machine/CI runner and is not the photo's real capture offset. See `pkg/raw/meta.go`'s `exif:DateTimeOriginal` handling: it recovers the wall-clock fields via `time.Local` and *relabels* them as UTC (no arithmetic conversion), so results are deterministic across hosts.
7. **HEIF "Decoder plugin generated an error: Unspecified"**: almost always means libheif was built with a codec (e.g. `libde265` for HEVC/HEIC) compiled out because its `-dev` package wasn't present when `make libheif` ran pkg-config detection - see "libheif codec selection" above. It's not a runtime plugin-path or shared-library-resolution issue; libheif is built statically and plugin loading is disabled.

## File Conventions

- `schema/` packages (both `gomedia/schema/` and `pkg/ffmpeg/schema/`) define request/response types with `kong` CLI tags
- `metadata/<kind>/` packages register `metadata.AddHandler` in `init()` rather than exposing a public extraction API directly
- FFmpeg C types are prefixed with `AV` (`AVCodec`, `AVFormatContext`, `AVFrame`); Go wrappers omit the prefix (`Reader`, `Decoder`, `Frame` wraps `AVFrame`)
- Metadata keys use a `namespace:name` convention (e.g. `exif:DateTimeOriginal`, `tiff:Make`, `artwork:thumbnail`) - see `gomedia.Metadata.Key()`
