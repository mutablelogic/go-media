package video

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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

// tvEntries returns the metadata of the episode of a TV show, whose name,
// season and episode are the query's, as TMDB has it, or nil when no show
// has the episode, or none matches the stream's duration
func tvEntries(ctx context.Context, client *tmdbclient.Client, q tmdbclient.Query, duration time.Duration) (map[string]gomedia.Metadata, error) {
	req := tmdbschema.TVSearchRequest{Term: q.Name}
	if q.Year > 0 {
		req.FirstAirDateYear = types.Ptr(strconv.Itoa(q.Year))
	}
	resp, err := client.SearchTVShows(ctx, req)
	if err != nil {
		return nil, err
	} else if resp == nil || len(resp.Results) == 0 {
		return nil, nil
	}

	show, episode, runtime, err := matchEpisode(ctx, client, q.Name, uint64(q.Season), uint64(q.Episode), resp.Results, duration)
	if err != nil || show == nil {
		return nil, err
	}

	// The title is the show's, with the season and episode, and the
	// episode's name, such as Breaking Bad – S01E02 – Cat's in the Bag...
	title := fmt.Sprintf("%s – S%02dE%02d", show.Name, episode.SeasonNumber, episode.EpisodeNumber)
	if episode.Name != "" {
		title += " – " + episode.Name
	}
	entries := map[string]gomedia.Metadata{
		"tmdb:id":      meta{key: "tmdb:id", value: strconv.FormatUint(episode.ID, 10)},
		"tmdb:showid":  meta{key: "tmdb:showid", value: strconv.FormatUint(show.ID, 10)},
		"tmdb:show":    meta{key: "tmdb:show", value: show.Name},
		"tmdb:season":  meta{key: "tmdb:season", value: episode.SeasonNumber},
		"tmdb:episode": meta{key: "tmdb:episode", value: episode.EpisodeNumber},
		"dc:title":     meta{key: "dc:title", value: title},
	}
	if episode.Name != "" {
		entries["tmdb:title"] = meta{key: "tmdb:title", value: episode.Name}
	}
	if episode.AirDate != "" {
		entries["tmdb:releasedate"] = meta{key: "tmdb:releasedate", value: episode.AirDate}
		entries["dc:date"] = meta{key: "dc:date", value: episode.AirDate}
	}
	if episode.Overview != "" {
		entries["tmdb:overview"] = meta{key: "tmdb:overview", value: episode.Overview}
		entries["dc:description"] = meta{key: "dc:description", value: episode.Overview}
	}
	if runtime > 0 {
		entries["tmdb:runtime"] = meta{key: "tmdb:runtime", value: runtime}
	}
	if episode.ExternalIDs != nil && episode.ExternalIDs.IMDbID != "" {
		entries["tmdb:imdbid"] = meta{key: "tmdb:imdbid", value: episode.ExternalIDs.IMDbID}
	}
	return entries, nil
}

// matchEpisode returns the show of the results, whose name is the name, and
// its episode, and the episode's runtime, as movies are matched: without the
// stream's duration, it's the first show with the name and the episode, or
// nil when none has it. With it, it's the first of the results whose name is the
// name, and whose episode's runtime is within 5% of the duration, or else the
// first of the top results whose episode is within 10% of it, or nil when
// none is. A show which doesn't have the episode isn't a match. An
// episode's runtime is the show's typical one when it isn't known.
func matchEpisode(ctx context.Context, client *tmdbclient.Client, name string, season, number uint64, results []tmdbschema.TVShow, duration time.Duration) (*tmdbschema.TVShow, *tmdbschema.TVEpisode, time.Duration, error) {
	// The episode of each show, which is got once, and is nil when the show
	// doesn't have it
	episodes := make(map[uint64]*tmdbschema.TVEpisode, len(results))
	episodeOf := func(show tmdbschema.TVShow) (*tmdbschema.TVEpisode, error) {
		if episode, exists := episodes[show.ID]; exists {
			return episode, nil
		}
		episode, err := client.GetTVEpisode(ctx, show.ID, season, number)
		if errors.Is(err, httpresponse.ErrNotFound) {
			episode, err = nil, nil
		} else if err != nil {
			return nil, err
		}
		episodes[show.ID] = episode
		return episode, nil
	}

	// The runtime of the show's episode, or else the show's typical one,
	// which is got once
	runtimes := make(map[uint64]time.Duration, len(results))
	runtimeOf := func(show tmdbschema.TVShow, episode *tmdbschema.TVEpisode) (time.Duration, error) {
		if episode.Runtime > 0 {
			return time.Duration(episode.Runtime) * time.Minute, nil
		} else if runtime, exists := runtimes[show.ID]; exists {
			return runtime, nil
		}
		details, err := client.GetTVShow(ctx, show.ID)
		if err != nil {
			return 0, err
		}
		var runtime time.Duration
		if len(details.EpisodeRunTime) > 0 {
			runtime = time.Duration(details.EpisodeRunTime[0]) * time.Minute
		}
		runtimes[show.ID] = runtime
		return runtime, nil
	}

	// match returns the first show with the episode, whose name is the name
	// when exact is set, of all the results, or else of the top results,
	// and whose runtime is within the percent of the duration, unless the
	// duration isn't known
	match := func(exact bool, percent int64) (*tmdbschema.TVShow, *tmdbschema.TVEpisode, time.Duration, error) {
		candidates := results
		if !exact {
			candidates = results[:min(len(results), tmdbCandidates)]
		}
		for i, show := range candidates {
			if exact && !showMatches(name, show) {
				continue
			}
			episode, err := episodeOf(show)
			if err != nil {
				return nil, nil, 0, err
			} else if episode == nil {
				continue
			}
			runtime, err := runtimeOf(show, episode)
			if err != nil {
				return nil, nil, 0, err
			}
			if duration <= 0 || runtimeMatches(runtime, duration, percent) {
				return &results[i], episode, runtime, nil
			}
		}
		return nil, nil, 0, nil
	}

	// A show with the name, within 5%, and then any show, within 10%. Without
	// the duration, only a show with the name, as any show with the episode
	// would otherwise match, such as another show whose name contains it.
	if show, episode, runtime, err := match(true, 5); err != nil || show != nil || duration <= 0 {
		return show, episode, runtime, err
	}
	return match(false, 10)
}

// showMatches returns true when the show's name, or its original name, is
// the name, as a movie's title is matched
func showMatches(name string, show tmdbschema.TVShow) bool {
	name = normalizeTitle(name)
	return name != "" && (normalizeTitle(show.Name) == name || normalizeTitle(show.OriginalName) == name)
}
