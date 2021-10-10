
# go-media ffmpeg bindings

This package provides bindings for some parts of the [ffmpeg API](https://ffmpeg.org/doxygen/trunk/)
including `libavformat`, `libavcodec` and `libavutil`.

This package is part of a wider project, `github.com/mutablogic/go-media`.
Please see the [module documentation](https://github.com/mutablogic/go-media/blob/master/README.md)
for more information.

## Introduction

You can use this package to access the ffmpeg API from Go. For example, in order to read the metadata and streams in a media file, use:

```
func main() {
    ctx := ffmpeg.NewAVFormatContext()
    err := ctx.OpenInput(os.Args[1], nil)
    if err != nil {
        panic(err)
    }
    defer ctx.CloseInput()
    fmt.Println("metadata=",ctx.GetMetadata())
}
```

You could also use this package to write a media file. For example, to create a video file from a video stream:

TODO

## Media Input

### Reading from a Path

### Reading from a URL

### Reading from an io.Reader

## Media Output

### Writing to Path

### Writing to an io.Writer

## Formats

TODO

## Codecs

TODO

## Reading Packets and Frames

TODO

### Metadata and Artwork

### Logging

