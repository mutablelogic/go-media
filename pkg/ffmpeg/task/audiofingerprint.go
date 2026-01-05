package task

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	// Packages
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	segmenter "github.com/mutablelogic/go-media/pkg/segmenter"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AudioFingerprint generates an audio fingerprint and optionally performs AcoustID lookup
func (m *Manager) AudioFingerprint(ctx context.Context, req *schema.AudioFingerprintRequest) (*schema.AudioFingerprintResponse, error) {
	// Build segmenter options from Request if needed
	var opts []segmenter.Opt

	// Determine input source
	var inputPath string
	var reader io.Reader

	if req.Path != "" {
		inputPath = req.Path
	} else if req.Reader != nil {
		reader = req.Reader
	} else {
		return nil, fmt.Errorf("either Reader or Path must be set")
	}

	// Convert duration
	var dur time.Duration
	if req.Duration > 0 {
		dur = time.Duration(req.Duration * float64(time.Second))
	}

	// If lookup is requested, we need a client
	if req.Lookup {
		// Build metadata flags
		flags := metadataFlags(req.Metadata)

		// Perform match with lookup (using path or reader)
		var matches []*chromaprint.ResponseMatch

		if inputPath != "" {
			// Open file for matching
			f, err := os.Open(inputPath)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			matches, err = m.chromaprint.Match(ctx, f, dur, flags, opts...)
			if err != nil {
				return nil, err
			}

			// Re-open for fingerprint
			f2, err := os.Open(inputPath)
			if err != nil {
				return nil, err
			}
			defer f2.Close()

			fpResult, err := chromaprint.Fingerprint(ctx, f2, dur, opts...)
			if err != nil {
				return nil, err
			}

			// Build response
			resp := &schema.AudioFingerprintResponse{
				Fingerprint: fpResult.Fingerprint,
				Duration:    fpResult.Duration.Seconds(),
			}

			// Convert matches
			if len(matches) > 0 {
				resp.Matches = make([]schema.AudioFingerprintMatch, len(matches))
				for i, m := range matches {
					resp.Matches[i] = convertMatch(m)
				}
			}

			return resp, nil
		} else {
			// Using reader - can't re-read for lookup
			return nil, fmt.Errorf("lookup requires re-reading the file; use Path instead of Reader")
		}
	}

	// Just fingerprint, no lookup
	if inputPath != "" {
		f, err := os.Open(inputPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		reader = f
	}

	result, err := chromaprint.Fingerprint(ctx, reader, dur, opts...)
	if err != nil {
		return nil, err
	}

	return &schema.AudioFingerprintResponse{
		Fingerprint: result.Fingerprint,
		Duration:    result.Duration.Seconds(),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// metadataFlags converts string metadata flags to chromaprint.Meta flags
func metadataFlags(metadata []string) chromaprint.Meta {
	if len(metadata) == 0 {
		return chromaprint.META_ALL
	}

	var flags chromaprint.Meta
	for _, m := range metadata {
		switch m {
		case "recordings":
			flags |= chromaprint.META_RECORDING
		case "recordingids":
			flags |= chromaprint.META_RECORDINGID
		case "releases":
			flags |= chromaprint.META_RELEASE
		case "releaseids":
			flags |= chromaprint.META_RELEASEID
		case "releasegroups":
			flags |= chromaprint.META_RELEASEGROUP
		case "releasegroupids":
			flags |= chromaprint.META_RELEASEGROUPID
		case "tracks":
			flags |= chromaprint.META_TRACK
		case "compress":
			flags |= chromaprint.META_COMPRESS
		case "usermeta":
			flags |= chromaprint.META_USERMETA
		case "sources":
			flags |= chromaprint.META_SOURCE
		}
	}

	return flags
}

// convertMatch converts chromaprint.ResponseMatch to schema.AudioFingerprintMatch
func convertMatch(m *chromaprint.ResponseMatch) schema.AudioFingerprintMatch {
	match := schema.AudioFingerprintMatch{
		Id:    m.Id,
		Score: m.Score,
	}

	if len(m.Recordings) > 0 {
		match.Recordings = make([]schema.AudioFingerprintRecording, len(m.Recordings))
		for i, r := range m.Recordings {
			rec := schema.AudioFingerprintRecording{
				Id:       r.Id,
				Title:    r.Title,
				Duration: r.Duration,
			}

			if len(r.Artists) > 0 {
				rec.Artists = make([]schema.AudioFingerprintArtist, len(r.Artists))
				for j, a := range r.Artists {
					rec.Artists[j] = schema.AudioFingerprintArtist{
						Id:   a.Id,
						Name: a.Name,
					}
				}
			}

			if len(r.ReleaseGroups) > 0 {
				rec.ReleaseGroups = make([]schema.AudioFingerprintGroup, len(r.ReleaseGroups))
				for j, g := range r.ReleaseGroups {
					group := schema.AudioFingerprintGroup{
						Id:    g.Id,
						Type:  g.Type,
						Title: g.Title,
					}

					if len(g.Releases) > 0 {
						group.Releases = make([]schema.AudioFingerprintRelease, len(g.Releases))
						for k, rel := range g.Releases {
							release := schema.AudioFingerprintRelease{
								Id: rel.Id,
							}

							if len(rel.Mediums) > 0 {
								release.Mediums = make([]schema.AudioFingerprintMedium, len(rel.Mediums))
								for l, med := range rel.Mediums {
									medium := schema.AudioFingerprintMedium{
										Position: med.Position,
									}

									if len(med.Tracks) > 0 {
										medium.Tracks = make([]schema.AudioFingerprintTrack, len(med.Tracks))
										for n, t := range med.Tracks {
											medium.Tracks[n] = schema.AudioFingerprintTrack{
												Position: t.Position,
											}
										}
									}

									release.Mediums[l] = medium
								}
							}

							group.Releases[k] = release
						}
					}

					rec.ReleaseGroups[j] = group
				}
			}

			match.Recordings[i] = rec
		}
	}

	return match
}
