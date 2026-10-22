package video

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	tmdbclient "github.com/mutablelogic/go-media/tmdb/httpclient"
	tmdbschema "github.com/mutablelogic/go-media/tmdb/schema"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// movieEntries returns the metadata of the movie whose name is the query's,
// as TMDB has it, and its poster and backdrop, or nil when none matches. The
// year is the stream's, from its own metadata, or zero when it isn't known.
func movieEntries(ctx context.Context, client *tmdbclient.Client, q tmdbclient.Query, duration time.Duration, year int) (map[string]gomedia.Metadata, []tmdbImage, error) {
	// Search for the movie by its name, and then by its American spelling,
	// as TMDB's search doesn't find other spellings, such as Judgment at
	// Nuremberg for Judgement at Nuremberg. The details of the movies are
	// got once, for every search.
	cache := make(map[uint64]*tmdbschema.MovieDetails)
	var movie *tmdbschema.Movie
	var details *tmdbschema.MovieDetails
	var err error
	for _, name := range spellings(q.Name) {
		if movie, details, err = findMovie(ctx, client, name, q.Year, year, duration, cache); err != nil || movie != nil {
			break
		}
	}
	if err != nil || movie == nil {
		return nil, nil, err
	}
	entries, images := movieMetadata(movie, details)
	return entries, images, nil
}

// movieByID returns the metadata of the movie with the identifier, as TMDB
// has it, and its poster and backdrop, or nil when TMDB doesn't have it
func movieByID(ctx context.Context, client *tmdbclient.Client, id uint64) (map[string]gomedia.Metadata, []tmdbImage, error) {
	details, err := client.GetMovie(ctx, id)
	if errors.Is(err, httpresponse.ErrNotFound) {
		return nil, nil, nil
	} else if err != nil {
		return nil, nil, err
	}
	entries, images := movieMetadata(&details.Movie, details)
	return entries, images, nil
}

// movieMetadata returns the metadata of the movie, with its details, which
// are nil when they aren't known, and its poster and backdrop
func movieMetadata(movie *tmdbschema.Movie, details *tmdbschema.MovieDetails) (map[string]gomedia.Metadata, []tmdbImage) {
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
	return entries, []tmdbImage{
		{"artwork:poster", tmdbclient.SizePoster, movie.PosterPath},
		{"artwork:backdrop", tmdbclient.SizeBackdrop, movie.BackdropPath},
	}
}

// findMovie returns the movie with the name, and its details, or nil when
// none matches. The year is the year in the stream's name, and the other year
// is the year known otherwise, such as from the stream's own metadata, which
// are zero when they aren't known.
func findMovie(ctx context.Context, client *tmdbclient.Client, name string, year, other int, duration time.Duration, cache map[uint64]*tmdbschema.MovieDetails) (*tmdbschema.Movie, *tmdbschema.MovieDetails, error) {
	// Search with the year in the name. The year known otherwise isn't used
	// in the first search, as it can be the year a file was made, rather
	// than the film's.
	known := year
	if known == 0 {
		known = other
	}
	movie, details, err := searchMovie(ctx, client, name, year, known, duration, cache)
	if err != nil || movie != nil {
		return movie, details, err
	}

	// When nothing matches, search again: in any year, when the name has a
	// year, as it can be the film's premiere, a year before its release on
	// TMDB, such as A Room with a View (1985), released in 1986; or in the
	// year known otherwise, as a common title's film can be beyond the first
	// page of results without it, such as Boy (2010)
	switch {
	case year > 0:
		return searchMovie(ctx, client, name, 0, year, duration, cache)
	case other > 0:
		return searchMovie(ctx, client, name, other, other, duration, cache)
	default:
		return nil, nil, nil
	}
}

// searchMovie searches for the movie with the name, released in the year
// searched, which is zero for any year, and returns the movie of the results
// which the stream, whose year is the year, which is zero when it isn't
// known, is, as matchMovie does
func searchMovie(ctx context.Context, client *tmdbclient.Client, name string, searched, year int, duration time.Duration, details map[uint64]*tmdbschema.MovieDetails) (*tmdbschema.Movie, *tmdbschema.MovieDetails, error) {
	req := tmdbschema.MovieSearchRequest{Term: name}
	if searched > 0 {
		req.PrimaryReleaseYear = types.Ptr(strconv.Itoa(searched))
	}
	resp, err := client.SearchMovies(ctx, req)
	if err != nil || resp == nil {
		return nil, nil, err
	}
	return matchMovie(ctx, client, name, year, resp.Results, duration, details)
}

// matchMovie returns the movie of the results which the stream, whose title
// is the name, and whose year is the year, which is zero when it isn't
// known, is, with its details, such as its runtime and IMDb identifier.
//
// With the year, it's the first of the results whose title is the name, and
// which was released in the year, whatever its runtime, as a version of a
// film, such as a different cut, can have a different runtime. Otherwise,
// without the stream's duration, it's the first of the results whose title
// is the name, without its details, or nil when none is.
// With it, it's the first of the results whose title is the name, and whose
// runtime is within 5% of the duration, or else the first of the top results
// whose runtime is within 10% of it, or nil when none is, including when
// their runtimes aren't known. A search's top results can be films whose
// titles only contain the name, such as recent ones, before the film itself.
func matchMovie(ctx context.Context, client *tmdbclient.Client, name string, year int, results []tmdbschema.Movie, duration time.Duration, details map[uint64]*tmdbschema.MovieDetails) (*tmdbschema.Movie, *tmdbschema.MovieDetails, error) {
	if len(results) == 0 {
		return nil, nil, nil
	}

	// The details of each result, which are got once, and kept in the map
	if details == nil {
		details = make(map[uint64]*tmdbschema.MovieDetails, len(results))
	}
	detailsOf := func(movie tmdbschema.Movie) (*tmdbschema.MovieDetails, error) {
		if _, exists := details[movie.ID]; !exists {
			d, err := client.GetMovie(ctx, movie.ID)
			if err != nil {
				return nil, err
			}
			details[movie.ID] = d
		}
		return details[movie.ID], nil
	}

	// A film with the name, released in the year
	if year > 0 {
		for i, result := range results {
			if titleMatches(name, result) && strings.HasPrefix(result.ReleaseDate, strconv.Itoa(year)) {
				d, err := detailsOf(result)
				if err != nil {
					return nil, nil, err
				}
				return &results[i], d, nil
			}
		}
	}
	// Without the duration, the first film with the name, as any film would
	// otherwise match, such as another whose title contains the name
	if duration <= 0 {
		for i, result := range results {
			if titleMatches(name, result) {
				return &results[i], nil, nil
			}
		}
		return nil, nil, nil
	}

	// The runtime of each result, from its details
	runtime := func(movie tmdbschema.Movie) (time.Duration, error) {
		d, err := detailsOf(movie)
		if err != nil {
			return 0, err
		}
		return time.Duration(d.Runtime) * time.Minute, nil
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

	// Any of the top films, within 10%
	for i, result := range results[:min(len(results), tmdbCandidates)] {
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
