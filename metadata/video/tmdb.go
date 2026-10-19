package video

import (
	"context"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

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
		movie, details, err := matchMovie(ctx, client, q.Name, resp.Results, o.Duration())
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

// matchMovie returns the movie of the results which the stream, whose title
// is the name, is. Without the stream's duration, it's the top result,
// without its details. With it, it's the first of the top results whose
// title is the name, and whose runtime is within 5% of the duration, or
// else the first whose runtime is within 10% of it, with its details, such
// as its runtime and IMDb identifier, or nil when none is, including when
// their runtimes aren't known. A search's top results can be films whose
// titles only contain the name, such as recent ones, before the film itself.
func matchMovie(ctx context.Context, client *tmdbclient.Client, name string, results []tmdbschema.Movie, duration time.Duration) (*tmdbschema.Movie, *tmdbschema.MovieDetails, error) {
	if len(results) == 0 {
		return nil, nil, nil
	} else if duration <= 0 {
		return &results[0], nil, nil
	}
	results = results[:min(len(results), tmdbCandidates)]

	// The details of each result, which are got once
	details := make(map[uint64]*tmdbschema.MovieDetails, len(results))
	runtime := func(movie tmdbschema.Movie) (time.Duration, error) {
		if _, exists := details[movie.ID]; !exists {
			d, err := client.GetMovie(ctx, movie.ID)
			if err != nil {
				return 0, err
			}
			details[movie.ID] = d
		}
		return time.Duration(details[movie.ID].Runtime) * time.Minute, nil
	}

	// A film with the name, within 5%
	for i, result := range results {
		if !titleMatches(name, result) {
			continue
		} else if r, err := runtime(result); err != nil {
			return nil, nil, err
		} else if runtimeMatches(r, duration, 5) {
			return &results[i], details[result.ID], nil
		}
	}

	// Any film, within 10%
	for i, result := range results {
		if r, err := runtime(result); err != nil {
			return nil, nil, err
		} else if runtimeMatches(r, duration, 10) {
			return &results[i], details[result.ID], nil
		}
	}
	return nil, nil, nil
}

// titleMatches returns true when the movie's title, or its original title,
// is the name, in any case, and ignoring punctuation and spaces between
// words, so Mission Impossible is Mission: Impossible
func titleMatches(name string, movie tmdbschema.Movie) bool {
	name = normalizeTitle(name)
	return name != "" && (normalizeTitle(movie.Title) == name || normalizeTitle(movie.OriginalTitle) == name)
}

// normalizeTitle returns the words of a title, in lowercase, separated by a
// space, without punctuation
func normalizeTitle(title string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(title), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}), " ")
}

// runtimeMatches returns true when the runtime is known, and within the
// percent of the stream's duration
func runtimeMatches(runtime, duration time.Duration, percent int64) bool {
	if runtime <= 0 {
		return false
	}
	diff := duration - runtime
	if diff < 0 {
		diff = -diff
	}
	return int64(diff)*100 <= int64(duration)*percent
}
