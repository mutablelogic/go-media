package video

import (
	"context"
	"io"
	"regexp"
	"strconv"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	tmdbclient "github.com/mutablelogic/go-media/tmdb/httpclient"
	tmdbschema "github.com/mutablelogic/go-media/tmdb/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

// tmdbCandidates is how many of the search's results are checked for a
// runtime which matches the stream's duration
const tmdbCandidates = 5

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Add metadata handler for video files: looks up the movie named by the
	// input's filename (see gomedia.NamedReader) on TMDB. Requires a TMDB
	// client (see WithTMDB); does nothing without one.
	metadata.AddHandler(regexp.MustCompile(`^video/.*$`), "tmdb", func(ctx context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		client := o.TMDB()
		if client == nil {
			return nil, nil
		}

		named, ok := r.(gomedia.NamedReader)
		if !ok || named.Name() == "" {
			return nil, nil
		}

		q := tmdbclient.ExtractQuery(named.Name())
		if q.Name == "" || q.Episode != 0 {
			// No usable title, or this looks like a TV episode: TMDB
			// search only covers movies for now.
			return nil, nil
		}

		req := tmdbschema.MovieSearchRequest{Term: q.Name}
		if q.Year > 0 {
			req.PrimaryReleaseYear = types.Ptr(strconv.Itoa(q.Year))
		}

		resp, err := client.SearchMovies(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp == nil || len(resp.Results) == 0 {
			return nil, nil
		}

		// TMDB's search endpoint returns results ranked by relevance. When
		// the stream's duration is known, take the first whose runtime
		// matches it, or else nothing. Otherwise, take the top match.
		movie, details, err := matchMovie(ctx, client, resp.Results, o.Duration())
		if err != nil {
			return nil, err
		} else if movie == nil {
			return nil, nil
		}
		entries := map[string]gomedia.Metadata{
			"tmdb:id":    meta{key: "tmdb:id", value: strconv.FormatUint(movie.ID, 10)},
			"tmdb:title": meta{key: "tmdb:title", value: movie.Title},
			"dc:title":   meta{key: "dc:title", value: movie.Title},
		}
		if movie.ReleaseDate != "" {
			entries["tmdb:releasedate"] = meta{key: "tmdb:releasedate", value: movie.ReleaseDate}
			entries["dc:date"] = meta{key: "dc:date", value: movie.ReleaseDate}
		}
		if movie.Overview != "" {
			entries["tmdb:overview"] = meta{key: "tmdb:overview", value: movie.Overview}
			entries["dc:description"] = meta{key: "dc:description", value: movie.Overview}
		}
		if details != nil {
			entries["tmdb:runtime"] = meta{key: "tmdb:runtime", value: time.Duration(details.Runtime) * time.Minute}
			if details.IMDbID != "" {
				entries["tmdb:imdbid"] = meta{key: "tmdb:imdbid", value: details.IMDbID}
			}
		}

		return metadata.FilterMetadata(entries, o), nil
	}, "tmdb", "dc")
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// matchMovie returns the movie of the results which the stream is. Without
// the stream's duration, it's the top result, without its details. With it,
// it's the first of the top results whose runtime is within 10% of the
// duration, with its details, such as its runtime and IMDb identifier, or
// nil when none is, including when their runtimes aren't known.
func matchMovie(ctx context.Context, client *tmdbclient.Client, results []tmdbschema.Movie, duration time.Duration) (*tmdbschema.Movie, *tmdbschema.MovieDetails, error) {
	if len(results) == 0 {
		return nil, nil, nil
	} else if duration <= 0 {
		return &results[0], nil, nil
	}
	for _, result := range results[:min(len(results), tmdbCandidates)] {
		details, err := client.GetMovie(ctx, result.ID)
		if err != nil {
			return nil, nil, err
		}
		if runtimeMatches(time.Duration(details.Runtime)*time.Minute, duration) {
			return &result, details, nil
		}
	}
	return nil, nil, nil
}

// runtimeMatches returns true when the runtime is known, and within 10% of
// the stream's duration
func runtimeMatches(runtime, duration time.Duration) bool {
	if runtime <= 0 {
		return false
	}
	diff := duration - runtime
	if diff < 0 {
		diff = -diff
	}
	return diff*10 <= duration
}
