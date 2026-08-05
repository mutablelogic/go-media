# metadata

Format-agnostic metadata extraction registry. Handlers register themselves against a
content-type regular expression and one or more namespaces (see `AddHandler`), and
`GetMetadata` runs the handlers that match a given content type, optionally narrowed to
specific namespaces via `WithNamespace`.

This document covers six namespaces specifically: `dc` (Dublin Core), `exif` (and its `tiff`
sibling), `image`, `video`, `audio`, and `tmdb` - metadata looked up from
[The Movie Database](https://www.themoviedb.org/) for video files.

## `dc` (Dublin Core) namespace

`dc:` entries come from five different kinds of source, so the set of keys isn't fixed:

1. **Container tags**, mapped onto a small, fixed set of `dc:` keys by the video/audio/RAW
   handlers.
2. **A TMDB match** (video files only), mirroring the title/overview/release-date of the
   movie found by the `tmdb` namespace handler - see [`tmdb` namespace](#tmdb-namespace)
   below.
3. **The most specific available EXIF capture date/time** (JPEG, RAW, and HEIF/AVIF - see
   [`exif` namespace](#exif-namespace-and-its-tiff-sibling) below), mirrored as `dc:date`.
4. **A video file's `creation-time` container tag** (see
   [`video`/`audio` namespaces](#video-and-audio-namespaces) below), reformatted as RFC 3339
   and mirrored as `dc:date`.
5. **Embedded XMP**, where any [Dublin Core](https://www.dublincore.org/specifications/dublin-core/dces/)
   element present in the file's XMP document is surfaced verbatim as `dc:<element>` - not
   limited to the fixed set below.

| Key | Value | Source |
| --- | --- | --- |
| `dc:title` | Title | Video `title` tag, audio `title`/`tracktitle` tag, or a TMDB match's title |
| `dc:creator` | Artist/director | Video `director` tag, audio `artist`/`artists`/`album-artist` tag, or a RAW file's own artist field |
| `dc:description` | Description | Video `description` tag (or `synopsis` if there's no `description` tag), a RAW file's own description field, or a TMDB match's overview |
| `dc:date` | Release date, `YYYY-MM-DD`, or full RFC 3339 timestamp | A TMDB match's release date, a JPEG/RAW/HEIF file's `exif:DateTimeOriginal` (preferred), `exif:DateTimeDigitized`, or `tiff:DateTime`, or a video file's `creation-time` tag |

## `exif` namespace (and its `tiff` sibling)

Unlike `dc`, `image`, and `tmdb`, `exif:`/`tiff:` keys aren't a fixed, enumerable set - they
mirror whatever standard EXIF tag names a file's own EXIF data actually contains. The prefix
follows the tag's IFD (Image File Directory), per `(*exif.Tag).Key()` in `pkg/exif/tag.go`:
`tiff:<Name>` for IFD0/IFD1 tags, `exif:<Name>` for Exif/GPS/Interop IFD tags.

JPEG (`metadata/image/exif.go`), a RAW file's embedded thumbnail (`metadata/image/raw.go`),
and HEIF/AVIF (`metadata/image/heif.go`) all feed their EXIF tags through the same shared
`exifTagsToMetadata` helper, so a few tags are enriched identically across all three: date/time
tags are merged with their UTC-offset tag and reformatted, and single-rational tags (GPS
coordinates, `FNumber`, `ExposureTime`, `FocalLength`, ...) get a parsed numeric value via
`Any()` alongside the unchanged human-readable string via `Value()`.

All three also mirror the most specific available capture date/time as `dc:date`
(`mirrorDCDate` in `metadata/image/exif.go`) - see [`dc` namespace](#dc-dublin-core-namespace)
above.

### Common keys

Real values from `etc/test/PXL_20250102_172041079.MP.jpg` (a JPEG, enriched):

| Key | `Value()` |
| --- | --- |
| `tiff:Make` | `Google` |
| `tiff:Model` | `Pixel 8` |
| `tiff:Software` | `HDR+ 1.0.702121622zd` |
| `tiff:DateTime` | `2025-01-02T18:20:41+01:00` |
| `exif:DateTimeOriginal` | `2025-01-02T18:20:41+01:00` |
| `exif:ISOSpeedRatings` | `410` |
| `exif:FNumber` | `f/1.7` |
| `exif:ExposureTime` | `1/80 sec.` |
| `exif:FocalLength` | `6.9 mm` |

### GPS keys

`Value()` for a GPS coordinate is **always** the raw degrees/minutes/seconds string - only
`Any()` is parsed, uniformly as a signed `float64` regardless of source (JPEG, RAW, or
HEIF/AVIF). Real values from `etc/test/photo.HEIC`:

| Key | `Value()` | `Any()` |
| --- | --- | --- |
| `exif:GPSLatitude` | `51, 32, 9.71` | `51.536030555555556` (`float64`, signed - negative south) |
| `exif:GPSLongitude` | `0, 8, 39.53` | `-0.14431388888888888` (`float64`, signed - negative west) |
| `exif:GPSAltitude` | `35.6820` | `35.68202589489718` (`float64`, negative if `GPSAltitudeRef` says below sea level) |

So: to compute with a coordinate, use `Any().(float64)` - safe across all three producers.
`exif:GPSLatitudeRef`/`GPSLongitudeRef`/`GPSAltitudeRef` (`N`/`S`, `E`/`W`, sea level or
below) are consumed and dropped from the output once merged into the corresponding signed
value.

RAW files also start from libraw's own curated `tiff:`/`exif:` fields (Make, Model,
Software, ISOSpeedRatings, ExposureTime, FNumber, FocalLength, DateTimeOriginal - see
`pkg/raw/meta.go`), overridden by the embedded thumbnail's enriched EXIF wherever the two
overlap.

## `image` namespace

`image:` entries describe the decoded image itself (dimensions and format). Exactly one of
two producers runs per file, since they're matched by disjoint content-type patterns -
the generic image handler's `^image/.*$` vs. the RAW handler's own `raw.ContentTypes` - so
there's no ambiguity about which meaning of `image:format` applies to a given file.

### Generic image handler (`metadata/image/image.go`)

Any `image/*` content type Go's stdlib `image` package can decode (plus the registered
HEIF/AVIF, BMP, TIFF, and WebP decoders).

| Key | Value |
| --- | --- |
| `image:format` | Decoder-reported format name: `jpeg`, `png`, `gif`, `bmp`, `tiff`, `webp`, `heif`, or `avif` |
| `image:width` | Pixel width |
| `image:height` | Pixel height |

### RAW camera handler (`metadata/image/raw.go`)

| Key | Value |
| --- | --- |
| `image:format` | The camera manufacturer name, lowercased (e.g. `olympus`, `canon`) - *not* a file-format identifier like `cr2`/`orf`, since libraw doesn't expose one |
| `image:width` | Processed pixel width, after crop/rotation |
| `image:height` | Processed pixel height, after crop/rotation |

## `video` and `audio` namespaces

Like `exif`/`tiff`, `video:`/`audio:` keys aren't a fully fixed set: `metadata/video/handler.go`
and `metadata/audio/handler.go` map a handful of container tag names onto canonical
`dc:`/`video:` (or `dc:`/`audio:`) keys, drop a list of noisy/uninteresting tags, and pass
everything else through as `video:<tag>`/`audio:<tag>` - so an unrecognized container tag
still shows up, just without a canonical mapping. Sanitizing a tag name lowercases it and
collapses any run of non-alphanumeric characters (including `_`) into a single `-`.

Both handlers are registered under `"dc"` as well as their own namespace, since `dc:title`/
`dc:creator` are two of the canonical mappings.

### `video:` (`metadata/video/handler.go`)

Real values from `etc/test/sample.mp4`:

| Key | Value | Source |
| --- | --- | --- |
| `video:duration` | `5.312` (seconds) | Always present |
| `dc:title` | `Sample From Big Buck Bunny` | `title` tag |
| `dc:creator` | - | `director` tag |
| `dc:description` | - | `description` tag, or `synopsis` if there's no `description` tag (see below) |
| `video:year` | - | `date`/`year` tag |
| `video:encoder` | `Lavf53.24.2` | Any other tag, sanitized and passed through |
| `video:creation-time` | `1970-01-01T00:00:00.000000Z` | Any other tag, sanitized and passed through |
| `dc:date` | `1970-01-01T00:00:00Z` | Mirrors `video:creation-time`, reformatted as RFC 3339, when it parses as a timestamp |

`synopsis` is never surfaced as its own `video:synopsis` key: it only becomes `dc:description`,
and only when there's no dedicated `description` tag to prefer instead (`buildVideoEntries` in
`metadata/video/handler.go`).

Dropped entirely (noisy/uninteresting): `compatible-brands`, `major-brand`, `minor-version`,
`comment`, `itunes-cddb-1`, `itunmovi`, `gapless-playback`, `itunextc`.

### `audio:` (`metadata/audio/handler.go`)

Real values from `etc/test/sample.mp3` (an untagged file, hence just `Duration`) and
`sample_with_artwork.mp3`:

| Key | Value | Source |
| --- | --- | --- |
| `audio:duration` | `20.734717` (seconds) | Always present |
| `dc:title` | - | `title`/`tracktitle` tag |
| `dc:creator` | - | `artist`/`artists`/`album-artist` tag |
| `audio:Album` | - | `album`/`albumtitle` tag |
| `audio:Genre` | - | `genre`/`music-genre` tag |
| `audio:year` | - | `originalyear`/`year`/`date`/`originaldate`/`tdor` tag |
| `audio:Track` | - | `track`/`tracknumber`/`itunes-cddb-tracknumber` tag |
| `audio:encoder` | `Lavf62.12.102` | Any other tag, sanitized and passed through |

Dropped entirely (mostly ID3/iTunes-specific noise): `comment`, `itunnorm`, `itunsmpb`,
`itunextc`, `gapless-playback`, `tagging-time`, `accurateripdiscid`, `accurateripresult`,
`id3v1-comment`, `id3v2-priv-averagelevel`, `id3v2-priv-google-originalclientid`,
`id3v2-priv-www-amazon-com`, `itunes-cddb-1`, `itunes-cddb-ids`, `account-id`,
`compatible-brands`, `major-brand`, `minor-version`, `eitunnorm`.

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
