# Tasks

## Metadata

### POST /probe

Returns the container format, streams, and metadata for a media file.

The request body can be either:

- `multipart/form-data` with a single file field named `file`; or
- any other content type, in which case the request body is read directly as the media stream.

`format` and `opts` query parameters can be used to force a specific input format (e.g. `mpegts`) and pass format-specific options, for input that FFmpeg can't detect automatically.
