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

## Encoder

### POST /encode

Starts an asynchronous encode of an uploaded media file and returns immediately - encoding is a long-running task, so this doesn't wait for it to finish. The task stays tracked by the task manager afterwards, so it can also be polled, cancelled, or watched via `GET /task/{uuid}`, `DELETE /task/{uuid}`, and `GET /task/event`.

The request body is always `multipart/form-data`, with two fields:

- `file` - the media file to encode; and
- `request` - a JSON-encoded object with `output` (required) and optional `audio`, `video`, and `subtitle` encoding profiles - at least one of which must be set. Each profile targets the first stream of its type found in the input.

The response depends on the request's `Accept` header:

- `Accept: text/event-stream` streams the task's own events - the same shape as `GET /task/event` filtered to this task's uuid - until it finishes or the client disconnects.
- Anything else returns the task's current status as `application/json` (the same shape as `GET /task/{uuid}`), reflecting whatever state the task is in at the moment it was started - typically `not_started` or `running`, not yet the final result.
