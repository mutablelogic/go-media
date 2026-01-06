# Device URL Scheme

## Overview

The gomedia task layer supports a unified URL scheme for accessing media from files, network streams, and hardware devices. This document describes the device URL scheme.

## URL Format

### File Paths

Regular file paths work as expected:

```
/absolute/path/to/file.mp4
relative/path/to/file.mp3
file.mov
```

### Network Streams

Network protocols are recognized by their scheme:

```
http://example.com/stream.m3u8
https://example.com/video.mp4
rtmp://server/live/stream
rtsp://camera/stream
```

### Device URLs

Hardware devices use the `device://` scheme:

```
device://format/device-path[?options]
```

Components:

- **format**: The FFmpeg input format name (e.g., `avfoundation`, `v4l2`, `alsa`, `pulse`)
- **device-path**: Format-specific device identifier
- **options**: Optional query parameters for device configuration

## Platform-Specific Devices

### macOS (AVFoundation)

AVFoundation provides access to cameras and microphones on macOS.

**List available devices:**

```bash
./build/gomedia list-formats --name avfoundation
```

**Basic usage:**

```
device://avfoundation/video-index:audio-index
```

**Examples:**

```bash
# Video and audio from default devices
./build/gomedia probe "device://avfoundation/0:0"

# Video only (no audio)
./build/gomedia probe "device://avfoundation/0:none"

# Audio only (no video)
./build/gomedia probe "device://avfoundation/none:0"

# With framerate option
./build/gomedia probe "device://avfoundation/0:0?framerate=30"

# With pixel format and framerate
./build/gomedia probe "device://avfoundation/0:0?framerate=30&pixel_format=uyvy422"

# With video size
./build/gomedia probe "device://avfoundation/0:0?video_size=1280x720&framerate=30"
```

**Common AVFoundation options:**

- `framerate`: Frame rate (e.g., `30`, `60`)
- `pixel_format`: Pixel format (e.g., `uyvy422`, `yuyv422`, `nv12`, `0rgb`, `bgr0`)
- `video_size`: Video dimensions (e.g., `1280x720`, `1920x1080`)
- `audio_buffer_size`: Audio buffer size in frames
- `list_devices`: Set to `true` to list devices (used internally)

### Linux (Video4Linux2)

V4L2 provides access to video capture devices on Linux.

**Examples:**

```bash
# Default video device
device://v4l2//dev/video0

# With options
device://v4l2//dev/video0?framerate=30&video_size=640x480
```

### Linux (ALSA)

ALSA provides audio capture on Linux.

**Examples:**

```bash
# Default audio device
device://alsa/hw:0

# Specific card
device://alsa/hw:1,0
```

### Linux (PulseAudio)

PulseAudio provides audio capture on Linux.

**Examples:**

```bash
# Default source
device://pulse/default

# Named source
device://pulse/alsa_input.usb-0000:00:1f.3.analog-stereo
```

## Implementation Details

### Format Lookup

Device formats are looked up using FFmpeg's device iteration APIs rather than `av_find_input_format()`, since some device-only formats (like avfoundation) are not available through the standard format lookup.

The implementation searches:

1. Video input devices (`AVDevice_input_video_device_first/next`)
2. Audio input devices (`AVDevice_input_audio_device_first/next`)

### Code Usage

```go
import (
    "github.com/mutablelogic/go-media/pkg/ffmpeg/task"
)

// Parse a device URL
parsed, err := task.ParseMediaURL("device://avfoundation/0:0?framerate=30")
if err != nil {
    // Handle error
}

// Open a reader from parsed URL
reader, err := task.OpenReader(parsed)
if err != nil {
    // Handle error
}
defer reader.Close()

// Or use the convenience function
reader, err := task.OpenReaderFromURL("device://avfoundation/0:0?framerate=30")
if err != nil {
    // Handle error
}
defer reader.Close()
```

## Testing

Run the URL parsing tests:

```bash
go test ./pkg/ffmpeg/task -run TestParseMediaURL -v
```

Test device probing:

```bash
# List available devices
./build/gomedia list-formats --name avfoundation

# Probe a device
./build/gomedia probe "device://avfoundation/0:0?framerate=30&pixel_format=uyvy422"
```
