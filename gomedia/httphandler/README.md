# Tasks

## Metadata

### POST /probe/media

Returns the container format, streams, and metadata for a media file.

The request body can be either:

- `multipart/form-data` with a single file field named `file`; or
- any other content type, in which case the request body is read directly as the media stream.

`format` and `opts` query parameters can be used to force a specific input format (e.g. `mpegts`) and pass format-specific options, for input that FFmpeg can't detect automatically.

### POST /probe/source

Returns the container format, streams, and metadata for a URL, rather than an uploaded file - a network source FFmpeg can read directly (`http`, `https`, `rtmp`, ...), or a local capture device addressed as `device://<format>/<address>` (e.g. `device://avfoundation/0:0`, `device://v4l2//dev/video0`).

The `url` query parameter is required. `format` and `opts` query parameters can be used to force a specific input format and pass format-specific options, the same as `/probe/media`.

### POST /metadata

Returns the detected content type and container-level metadata (e.g. `title`, `artist`) for a media file - excludes artwork and chapters, and doesn't probe streams, unlike `/probe/media`.

The request body can be either:

- `multipart/form-data` with a single file field named `file`; or
- any other content type, in which case the request body is read directly as the media stream.

`filter` narrows the result to a single metadata namespace or key - `"namespace:"` (all keys in that namespace), `"name"` (this name in any namespace), or `"namespace:name"` (one specific key); omitted means include all keys. Unlike `/probe/media`, there's no `format`/`opts` override - metadata extraction detects content type from the bytes themselves rather than using FFmpeg's demuxer.
