# Device URLs

Audio and video devices, such as cameras and microphones, are named with a
`device://` URL wherever go-media takes a URL: `gomedia probe-source` and
`segmenter stream` accept them, alongside files and network streams.

## URL Format

```text
device://<format>/<address>
```

- **format** is the name of the FFmpeg input device, such as `avfoundation`,
  `v4l2`, `alsa` or `pulse`.
- **address** is how that device names what to capture, which depends on the
  format. It's everything after the `/` which follows the format, so a path
  which starts with `/` gives `//`, such as `device://v4l2//dev/video0`.

Options for the device, such as a frame rate or video size, aren't part of the
URL, as anything after a `?` is ignored. Pass them separately instead, as
`key=value` pairs, such as with `gomedia probe-source --opts`.

Files and network streams are given as they are:

```text
/path/to/file.mp4
https://example.com/stream.m3u8
rtmp://server/live/stream
rtsp://camera/stream
```

## Listing Devices

`segmenter stream`, with no URL, lists the audio input devices, such as
microphones, with their URLs:

```text
$ segmenter stream
URL                       DEVICE              DEFAULT
device://avfoundation/:0  ZoomAudioDevice     yes
device://avfoundation/:1  Logitech StreamCam
device://avfoundation/:2  EarPods Microphone
```

`gomedia devices` lists all the input and output devices, audio and video, from
a running `gomedia` server:

```bash
gomedia devices
gomedia devices --format avfoundation --is-input
```

## Platforms

### macOS (AVFoundation)

AVFoundation numbers video devices and audio devices separately, from zero.
The address is `<video>:<audio>`, either of which can be `none`, or left out
for no video:

```text
device://avfoundation/0:0      # The first camera and the first microphone
device://avfoundation/0:none   # The first camera, with no audio
device://avfoundation/:1       # The second microphone, with no video
```

Common options include `framerate` (such as `30`), `video_size` (such as
`1280x720`), `pixel_format` (such as `uyvy422` or `nv12`) and
`audio_buffer_size`. For example:

```bash
gomedia probe-source "device://avfoundation/0:none" --opts framerate=30,video_size=1280x720
```

macOS asks for permission to use the camera or microphone the first time an
application, such as your terminal, uses one.

### Linux (Video4Linux2)

```text
device://v4l2//dev/video0
```

Common options include `framerate`, `video_size` and `input_format`.

### Linux (ALSA)

```text
device://alsa/default
device://alsa/hw:1,0           # Card 1, device 0
```

### Linux (PulseAudio)

```text
device://pulse/default
device://pulse/alsa_input.usb-0000:00:1f.3.analog-stereo
```

## From Go

A device is opened with the `reader` package by giving its format as the
input format, and its address as the URL. FFmpeg's devices are registered
when the `sys/ffmpeg80` package is loaded, so they can be found by name:

```go
import (
    reader "github.com/mutablelogic/go-media/reader"
)

// The first camera, at 30 frames per second
r, err := reader.Open("0:none", reader.WithInput("avfoundation", "framerate=30"))
if err != nil {
    return err
}
defer r.Close()
```

The segmenter takes the same reader options, so it can segment a microphone
as it records:

```go
seg, err := segmenter.Open(":0", 16000,
    segmenter.WithReaderOpt(reader.WithInput("avfoundation")),
    segmenter.WithMode(segmenter.ModeVADAggressive, 0, 0.01),
)
```

`device://` URLs themselves are parsed by the commands which take them, in
`gomedia/task` (for `gomedia probe-source`) and `cmd/segmenter` (for
`segmenter stream`), which turn them into an input format and an address like
this.
