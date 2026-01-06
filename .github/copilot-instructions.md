# go-media AI Agent Instructions

## Architecture Overview

This is a Go module providing FFmpeg bindings with three architectural layers:

1. **sys/ffmpeg80/** - CGO bindings to FFmpeg 8.0 C libraries (low-level, mirrors FFmpeg API)
2. **pkg/ffmpeg/** - High-level Go API with Reader, Decoder, Encoder, Resampler abstractions
3. **pkg/ffmpeg/task/** - Task Manager pattern exposing common operations (probe, remux, decode) via schema-based requests

Additional components:
- **pkg/sdl/** - SDL2-based video/audio player (Window, Audio, Player, FrameLoop abstractions)
- **pkg/chromaprint/** - Audio fingerprinting using Chromaprint library
- **cmd/gomedia/** - CLI tool using kong for argument parsing

## Critical Build Workflow

**First build ALWAYS requires full make** to compile FFmpeg and Chromaprint static libraries:
```bash
make              # Builds FFmpeg 8.0.1, Chromaprint 1.5.1, then gomedia binary
```

**Subsequent builds** use fast incremental compilation:
```bash
make cmd/gomedia  # Uses cached FFmpeg/Chromaprint, rebuilds Go only
```

The Makefile sets critical CGO environment variables:
```bash
PKG_CONFIG_PATH="${PWD}/build/install/lib/pkgconfig"
CGO_LDFLAGS="-lstdc++ -Wl,-no_warn_duplicate_libraries"
```

Never use `go build` directly without these vars or you'll get pkg-config errors.

## FFmpeg Integration Patterns

### Frame Buffer Management
**Critical**: FFmpeg reuses frame buffers across decode operations. When queueing frames for SDL or async processing, ALWAYS copy frames first:

```go
// decoder.go, playback.go pattern
copy, err := frame.Copy()  // Allocates new buffer
if err != nil {
    return err
}
frameCh <- copy  // Safe to pass to another goroutine
```

### Stream Decoder Initialization
When creating decoders, handle unsupported codecs gracefully (e.g., `dvb_teletext` subtitles may not be compiled in):

```go
// decoder.go pattern
codec := ff.AVCodec_find_decoder(codecID)
if codec == nil {
    return nil, errors.Join(ErrCodecNotFound, errors.New("codec: "+codecID.Name()))
}

// In mapStreams(), skip unsupported codecs:
if errors.Is(err, ErrCodecNotFound) {
    continue  // Skip this stream, don't fail entire decode
}
```

### Reader Input Options
Support per-format demuxer tuning via `InputFormat` and `InputOpts` fields:

```go
// schema/base.go Request struct
InputFormat string   `kong:"name='input-format',help='Force input format'"`
InputOpts   []string `kong:"name='input-opt',help='Input format options (key=value)'"`

// ffmpeg/reader.go usage
opt := ffmpeg.WithInput(req.InputFormat, req.InputOpts...)
reader, err := ffmpeg.Open(req.Input, opt)
```

Auto-detection heuristic for MPEG-TS files in `cmd/gomedia/main.go`:
```go
if strings.HasSuffix(strings.ToLower(req.Input), ".ts") {
    req.InputFormat = "mpegts"
    req.InputOpts = []string{
        "probesize=5000000",
        "analyzeduration=10000000",
        "fflags=+genpts",
        "discardcorrupt=1",
    }
}
```

## SDL Player Architecture

**SDL MUST run on the main thread on macOS**. The pattern uses:

1. **Context** - Initializes SDL, manages event loop via goroutine registration
2. **FrameLoop** - Runs on main thread, dispatches frames from channel via user events
3. **Player** - Manages Window and Audio, handles format conversion (swscale for video, planar→interleaved for audio)

Key pattern in `cmd/gomedia/main.go`:
```go
player := sdl.NewPlayer(ctx)
frameLoop := sdl.NewFrameLoop(ctx, player.PlayFrame, 100, 
    sdl.WithFrameDelayFunc(player.VideoDelay))  // PTS-based timing

// Decode feeds frames to FrameWriter adapter
err := task.Decode(req, frameLoop.NewFrameWriter())
```

### Video Timing
Use PTS-based delay, NOT fixed 33ms intervals:
```go
// pkg/sdl/player.go VideoDelay()
delta := float64(frame.Pts()-p.lastPTS) * ff.AVUtil_rational_q2d(frame.TimeBase())
return time.Duration(clamp(delta, 0, 0.25) * float64(time.Second))
```

### Format Conversion
Player auto-converts video formats via cached resampler:
- YUV420P, RGB24 → native SDL texture update
- All other formats → swscale to YUV420P, then Update()

## Testing Patterns

### Debugging FFmpeg Operations
Enable detailed logging via `GOMEDIA_SDL_DEBUG=1` environment variable for SDL operations.

For FFmpeg reader/decoder issues, check:
```go
// pkg/ffmpeg/logging.go provides callback registration
ff.AVUtil_log_set_callback(logCallback)
ff.AVUtil_log_set_level(ff.AV_LOG_VERBOSE)
```

### Test File Requirements
- Valid test file: `etc/test/sample.mp4` (h264/aac, 1280x720, 5.3s duration)
- MPEG-TS files may have unsupported subtitle codecs (dvb_teletext) - this is expected, decoder should skip them

## Common Gotchas

1. **Stream JSON Marshaling**: AVStream.MarshalJSON delegates to C struct which outputs garbage. Use manual field extraction with AVCodecPar(), TimeBase(), etc. (see `pkg/ffmpeg/schema/stream.go`)

2. **AVStream API Naming**: Methods are `Id()` not `ID()`, use `AVCodecID.Name()` not `AVCodec_get_name()`

3. **AV_NOPTS_VALUE**: Cast to `int64` for comparisons: `pts != int64(ff.AV_NOPTS_VALUE)`

4. **Texture Recreation**: SDL textures must be recreated when format or dimensions change - see `pkg/sdl/window.go ensureTexture()`

5. **Audio Channel Layouts**: Modern FFmpeg uses AVChannelLayout struct, not legacy channel mask integers

## File Conventions

- **schema/** packages define request/response types with kong CLI tags
- **task/** operations return `(Response, error)` or `error` only
- FFmpeg C types prefixed with `AV` (AVCodec, AVFormatContext, AVFrame)
- Go wrappers omit prefix (Reader, Decoder, Frame wraps AVFrame)
