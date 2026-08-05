# metadata

Format-agnostic metadata extraction registry. Handlers register themselves against a
content-type regular expression and one or more namespaces (see `AddHandler`), and
`GetMetadata` runs the handlers that match a given content type, optionally narrowed to
specific namespaces via `WithNamespace`.

This document covers two namespaces specifically: `dc` (Dublin Core) and `tmdb` - metadata
looked up from [The Movie Database](https://www.themoviedb.org/) for video files.

## `dc` (Dublin Core) namespace

`dc:` entries come from three different kinds of source, so the set of keys isn't fixed:

1. **Container tags**, mapped onto a small, fixed set of `dc:` keys by the video/audio/RAW
   handlers.
2. **A TMDB match** (video files only), mirroring the title/overview/release-date of the
   movie found by the `tmdb` namespace handler - see [`tmdb` namespace](#tmdb-namespace)
   below.
3. **Embedded XMP**, where any [Dublin Core](https://www.dublincore.org/specifications/dublin-core/dces/)
   element present in the file's XMP document is surfaced verbatim as `dc:<element>` - not
   limited to the fixed set below.

### Container-tag / TMDB keys (`metadata/video/handler.go`, `metadata/audio/handler.go`, `pkg/raw/meta.go`, `metadata/video/tmdb.go`)

| Key | Value | Source |
| --- | --- | --- |
| `dc:title` | Title | Video `title` tag, audio `title`/`tracktitle` tag, or a TMDB match's title |
| `dc:creator` | Artist/director | Video `director` tag, audio `artist`/`artists`/`album-artist` tag, or a RAW file's own artist field |
| `dc:description` | Description | Video `description` tag, a RAW file's own description field, or a TMDB match's overview |
| `dc:date` | Release date, `YYYY-MM-DD` | A TMDB match's release date |

### XMP-derived keys (HEIF/AVIF via `metadata/image/heif.go`; Photoshop via `metadata/application/photoshop.go`)

Any Dublin Core element found in the file's embedded XMP - `dc:title`, `dc:creator`,
`dc:description`, `dc:subject`, `dc:date`, `dc:rights`, `dc:publisher`, and so on - is
surfaced with whatever value the XMP document holds, not just the fixed set above. Which
elements appear, if any, depends entirely on what the file's own XMP contains.

## `tmdb` namespace

Unlike `dc` and the other content-derived namespaces (`exif`, `tiff`, `image`, `artwork`,
...), `tmdb` entries come from a network lookup: the input's filename is parsed into a
search query (see `tmdb.ExtractQuery` in
`tmdb/httpclient/query.go`) and used to search TMDB for a matching movie. The handler is a
no-op unless both of the following are true:

- A TMDB client was configured via `metadata.WithTMDB(token)` - without one, no lookup is
  attempted and no `tmdb:` entries are returned.
- The input reader implements `gomedia.NamedReader` (i.e. it has a filename) and that
  filename parses to a non-empty title.

Only movies are searched. If the filename looks like a TV episode (an `S01E01`-style marker
was found), the handler returns no metadata, since TMDB's TV search isn't wired up yet.

| Key | Value | Always present? |
| --- | --- | --- |
| `tmdb:id` | The matched movie's TMDB ID (numeric string) | Yes, when a match is found |
| `tmdb:title` | The matched movie's official TMDB title | Yes, when a match is found |
| `tmdb:releasedate` | Release date, `YYYY-MM-DD` | Only if TMDB returned one |
| `tmdb:overview` | Synopsis/overview text | Only if TMDB returned one |

The same handler also mirrors `tmdb:title`, `tmdb:overview`, and `tmdb:releasedate` as
`dc:title`, `dc:description`, and `dc:date` respectively (see [`dc` namespace](#dc-dublin-core-namespace)
above), so it's registered under both the `tmdb` and `dc` namespaces.

If the search returns no results (or none of the above conditions are met), the handler
returns no metadata and no error - a missing TMDB match isn't treated as a failure.

TMDB's search endpoint returns results ranked by relevance; the handler takes the top match
without further disambiguation (e.g. by release year proximity).
