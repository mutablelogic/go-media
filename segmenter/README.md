# Segmenter

The segmenter reads media files, and segments them into sections, each containing a contiguous block of speech, music, noise or silence. It can cut at a fixed duration, at silence, or near times you give it, and it classifies
each segment.

It's designed for preparing audio for speech recognition and analysis: for
example, splitting a podcast into parts at its pauses, sending each part to a
transcriber, and skipping the music and silence between them.

```go
import "github.com/mutablelogic/go-media/segmenter"

seg, err := segmenter.Open("podcast.mp3", 16000,
 segmenter.WithSegments(30*time.Second),
 segmenter.WithMode(segmenter.ModeVADAggressive, 0, 0),
)
if err != nil {
 return err
}
defer seg.Close()

err = seg.Decode(ctx, func(s *segmenter.Segment) error {
 fmt.Println(s.Start, s.Duration, s.Class, s.Stats.Speech)
 return nil
})
```

`Open` takes a file path or URL, and `NewReader` takes an `io.Reader`. The
callback is called for each segment, in order, with segments which follow
on from each other with no gaps. A segment's samples are only valid during
the callback, so copy them to keep them. Return `io.EOF` from the callback to
stop early.

## Building

The segmenter uses cgo, and links FFmpeg and [libfvad](https://github.com/dpirch/libfvad)
as static libraries, which the Makefile downloads and builds into
`build/install`, with their pkg-config files. See the [requirements](../README.md#requirements)
for the tools you need to build them. Then:

```bash
make ffmpeg libfvad   # Build the libraries the segmenter needs
make test-libfvad     # Build them if needed, and run the segmenter tests
make cmd/segmenter    # Build the segmenter command, to build/segmenter
```

To run `go build` or `go test` yourself, set the same environment variables
the Makefile does, so cgo finds the libraries:

```bash
export PKG_CONFIG_PATH="${PWD}/build/install/lib/pkgconfig"
export CGO_LDFLAGS_ALLOW="-(W|D).*"
export CGO_LDFLAGS="-lstdc++ -Wl,-no_warn_duplicate_libraries"  # drop the -Wl,... flag on Linux
go test ./segmenter
```

In another module which uses the segmenter, set `PKG_CONFIG_PATH` to the
`lib/pkgconfig` folder of an installation with FFmpeg's libraries and
`libfvad.pc`, such as this one's `build/install`.

## Options

| Option | Description |
| --- | --- |
| `WithSegments(d)` | Output a segment when it reaches the duration `d` (minimum 100ms) |
| `WithSegments(t1, t2, ...)` | Cut segments near these times from the start, at least two seconds apart, preferring a silence within a second of each, so there's one segment for each part between them |
| `WithMode(mode, silence, threshold)` | How silence is detected, and whether segments are classified (see below) |
| `WithStream(index)` | The audio stream to decode. Without it, the audio stream with the lowest index is decoded |
| `WithReaderOpt(opt)` | An option for the reader, such as `reader.WithInput("s16le", "sample_rate=16000", "channels=1")` for raw audio, or `reader.WithInput("avfoundation")` for a device |

## Modes

`WithMode` sets how silence is detected. Segments are cut where silence has
lasted at least the `silence` duration, which is 500ms if it's zero.

| Mode | Silence is where | Classification |
| --- | --- | --- |
| `ModeNone` (default) | Never | Silence or unknown |
| `ModeSilence` | The RMS energy is below `threshold` (0.01 if zero) | Silence or unknown |
| `ModeVADQuality`, `ModeVADLowBitrate`, `ModeVADAggressive`, `ModeVADVeryAggressive` | There's no speech, or the RMS energy is below `threshold` if it's set | Silence, speech, music or other |

The voice activity detection (VAD) modes use [libfvad](https://github.com/dpirch/libfvad),
the WebRTC voice activity detector, in increasing order of aggressiveness: a
more aggressive mode is more likely to be right when it reports speech, but
misses more of it. `ModeVADAggressive` is a good choice. They work at any
sample rate, as the segmenter resamples a copy of the audio to 8000 Hz for
the detector.

## How segments are cut

- **At a fixed duration**, with `WithSegments(d)`.
- **At silence**: when silence has lasted the silence duration, the segment
  before it is output straight away, without waiting for the silence to end.
  The silence is then output as a segment of its own, classified as silence,
  when sound starts again or it reaches the segment duration. A segment
  shorter than half the segment duration is output with the start of the
  silence after it, rather than on its own.
- **Near times you give**, with `WithSegments(t1, t2, ...)`. The audio is cut
  into pieces of at most a second, and at silences, and the pieces are joined
  together up to each time. Each cut is at the edge between two pieces which
  is nearest the time, preferring an edge next to silence within a second of
  it. So with silence detection, cuts are usually in a pause, and without it,
  they're within half a second of the time. Times after the end of the audio
  are ignored.

## Classification

Each segment has a `Class`, and `Stats` which it was classified from:

| Class | Meaning |
| --- | --- |
| `ClassSilence` | The segment was detected as silence, or its RMS energy is below the threshold |
| `ClassSpeech` | Mostly speech |
| `ClassMusic` | Music, or other sound with a slowly changing spectrum |
| `ClassOther` | Other sound, such as noise |
| `ClassUnknown` | Not silence, and not classified further, as the mode isn't a VAD mode |

The voice activity detector alone reports most sounds as speech, including
music, tones and noise, so the classification also uses features measured over
20ms frames. Speech has pauses between syllables and words, so many frames
with low energy (`LowEnergy`), a zero-crossing rate which varies as voiced and
unvoiced sounds alternate (`ZCRVariation`), and a spectrum which changes
quickly (`Flux`). When background sound, such as in a news report from the
field, fills the pauses, fewer frames have low energy, so a zero-crossing rate
which varies a lot also counts as speech. Music has a spectrum which changes
more slowly, and the detector reports it as speech (`Speech`). Everything else
is other sound.

A segment is classified from at least the last six seconds of audio before
its end (`ClassifyWindow`), including audio before the segment if it's
shorter, so short segments are classified as reliably as long ones, without
waiting for any audio after them. `RMS` and `Peak` are measured over the
segment only, and the other stats over that window. `Stats` also includes
features which aren't used for classification yet, such as `Harmonicity`,
`Rhythm` and `Centroid`, to help tune it.

### Accuracy

The thresholds were chosen from the [MUSAN](https://www.openslr.org/17/)
corpus of speech, music and noise, a music library and a news podcast. With
`ModeVADAggressive` and two second segments, across all 109 hours of MUSAN:

| Recording | Speech | Music | Other | Silence |
| --- | --- | --- | --- | --- |
| Speech | **86.2%** | 7.2% | 2.2% | 4.4% |
| Music | 5.0% | **83.3%** | 8.8% | 2.9% |
| Noise | 11.8% | 23.3% | **48.3%** | 16.6% |

On a sample of 300 tracks from a personal music library, about 80% of the
audio was classified as music, and on a 25 minute news podcast, about 98% was
classified as speech.

### Limitations

- Loud, dense music, such as distorted rock, and sparse ambient music, are
  often classified as other sound.
- A quarter of noise is classified as music, as much of it is harmonic or
  rhythmic, such as engines, sirens and machinery.
- Speech over music is usually classified as music.
- Music with sound effects, such as a cartoon soundtrack, can be classified
  as speech.
- Just after a change, such as from music to speech, a short segment can be
  classified partly from the audio before it.

## Real-time

The segmenter works on live sources, such as a stream or a device, as well
as files, and adds little latency:

- It processes audio about 450 times faster than real time on one core of
  a Mac Studio (Apple M4 Max), with a VAD mode.
- A segment is output as soon as the silence after it has lasted the silence
  duration, or it reaches the segment duration. The voice activity detector
  takes about 120ms to notice that speech has ended.
- Classification only uses audio before the end of a segment.
- With `WithSegments(t1, t2, ...)`, a segment is output when audio has arrived
  up to a second after its time, to look for a silence near it.
- When the context passed to `Decode` is cancelled, such as to stop
  recording, the audio so far is output as the last segment.

To record from a microphone, open the device with its input format. On macOS,
with AVFoundation, `:0` is the first audio device with no video, and
`ffmpeg -f avfoundation -list_devices true -i ""` lists them:

```go
seg, err := segmenter.Open(":0", 16000,
	segmenter.WithReaderOpt(reader.WithInput("avfoundation")),
	segmenter.WithMode(segmenter.ModeVADAggressive, 0, 0.01),
)
```

With a VAD mode, a threshold such as 0.01 stops the background noise of a
quiet room, which the detector sometimes reports as speech, from cutting the
silence between speech into many short segments.

## Testing

```bash
make test-libfvad   # Tests for sys/libfvad and the segmenter
make musan          # Download the MUSAN corpus (11GB) to build/musan
make test-musan     # Classify 50 files of each kind from MUSAN, and summarise
make test-musan MUSAN_LIMIT=0   # Classify all of MUSAN
```

## The segmenter command

[cmd/segmenter](../cmd/segmenter) runs the segmenter from the command line,
locally. `segmenter segment` prints the segments of a file or URL, and given a
folder after it, writes each one to a file in the folder, encoded with
`--format` (`wav`, `mp3` or `aac`), so you can listen to them:

```bash
segmenter segment --segments 3m10s,7m45s --format mp3 podcast.mp3 parts
segmenter segment --mode silence --csv stats.csv podcast.mp3
```

`segmenter stream` shows the segments of a stream or a device, such as a
microphone, as it plays, without writing anything. A device is
`device://<format>/<address>`, and `segmenter stream` with no URL lists the
audio input devices. On a terminal, it shows a meter of the level, and the
class and stats of the last segment, below the segments as they're output.
It stops at the end of the stream, after `--duration`, or when you press
Ctrl-C. It only takes `--mode`, `--silence`, `--threshold` (0.01 by default,
so background noise isn't taken for sound) and `--duration`:

```bash
segmenter stream device://avfoundation/:0
```

`segmenter classify` classifies folders of labelled audio, and summarises how
each label was classified and how each feature is distributed. Each folder is
named for its label, or given as `label=folder`. `make test-musan` runs it on
MUSAN, writing the class and stats of each segment to `build/musan.csv`:

```bash
segmenter classify --limit 300 --csv music.csv music=/path/to/music
```

`segment` and `classify` take the same options as the package: `--mode`, `--silence`,
`--threshold`, `--segments`, `--rate` and `--stream`, and `--input-format` and
`--input-opts` for raw audio.
