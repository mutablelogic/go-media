package schema

import (
	"net/url"
	"strconv"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// TVSearchRequest searches for TV shows by their name
type TVSearchRequest struct {
	Term             string  `json:"query" arg:"" name:"query" help:"Search term."`
	FirstAirDateYear *string `json:"first_air_date_year,omitempty" help:"Filter results by the year the show was first aired."`
	IncludeAdult     *bool   `json:"include_adult,omitempty" negatable:"" help:"Include adult content in search results."`
	Language         *string `json:"language,omitempty" help:"Language for the search results."`
	Page             *uint   `json:"page,omitempty" help:"Page number of the results."`
	Year             *string `json:"year,omitempty" help:"Filter results by the year of any of the show's episodes."`
}

// TVSearchResponse is a page of the TV shows which match a search, with the
// most relevant first
type TVSearchResponse struct {
	Page         uint64   `json:"page"`
	Results      []TVShow `json:"results"`
	TotalPages   uint64   `json:"total_pages"`
	TotalResults uint64   `json:"total_results"`
}

// TVShow is a TV show, as a search returns it
type TVShow struct {
	ID               uint64   `json:"id"`
	Name             string   `json:"name"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	FirstAirDate     string   `json:"first_air_date"`
	Overview         string   `json:"overview"`
	Genres           []uint64 `json:"genre_ids"`
	OriginCountry    []string `json:"origin_country,omitempty"`

	BackdropPath string `json:"backdrop_path"`
	PosterPath   string `json:"poster_path"`

	Popularity  float32 `json:"popularity,omitempty"`
	VoteAverage float32 `json:"vote_average,omitempty"`
	VoteCount   uint64  `json:"vote_count,omitempty"`

	Adult bool `json:"adult,omitempty"`
}

// TVShowDetails is a TV show, with the details which a search doesn't
// return, such as the runtimes of its episodes
type TVShowDetails struct {
	TVShow

	// The typical runtimes of the show's episodes in minutes, which is empty
	// when they aren't known
	EpisodeRunTime []uint64 `json:"episode_run_time,omitempty"`

	NumberOfSeasons  uint64 `json:"number_of_seasons"`
	NumberOfEpisodes uint64 `json:"number_of_episodes"`
	LastAirDate      string `json:"last_air_date,omitempty"`
	Status           string `json:"status,omitempty"`

	// The show's identifiers in other databases, such as IMDb, when they're
	// requested with its details
	ExternalIDs *ExternalIDs `json:"external_ids,omitempty"`
}

// TVEpisode is an episode of a TV show
type TVEpisode struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	Overview       string `json:"overview"`
	AirDate        string `json:"air_date"`
	SeasonNumber   uint64 `json:"season_number"`
	EpisodeNumber  uint64 `json:"episode_number"`
	ProductionCode string `json:"production_code,omitempty"`

	// The runtime in minutes, which is zero when it isn't known
	Runtime uint64 `json:"runtime"`

	StillPath   string  `json:"still_path"`
	VoteAverage float32 `json:"vote_average,omitempty"`
	VoteCount   uint64  `json:"vote_count,omitempty"`

	// The episode's identifiers in other databases, such as IMDb, when
	// they're requested with its details
	ExternalIDs *ExternalIDs `json:"external_ids,omitempty"`
}

// ExternalIDs are the identifiers of a TV show or episode in other
// databases, such as IMDb, each of which is empty when it isn't known
type ExternalIDs struct {
	IMDbID   string `json:"imdb_id,omitempty"`
	TVDBID   uint64 `json:"tvdb_id,omitempty"`
	WikiData string `json:"wikidata_id,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r *TVSearchRequest) Query() url.Values {
	response := url.Values{}
	if r.Term != "" {
		response.Set("query", r.Term)
	}
	if year := types.Value(r.FirstAirDateYear); year != "" {
		response.Set("first_air_date_year", year)
	}
	if r.IncludeAdult != nil {
		response.Set("include_adult", strconv.FormatBool(types.Value(r.IncludeAdult)))
	}
	if language := types.Value(r.Language); language != "" {
		response.Set("language", language)
	}
	if r.Page != nil {
		response.Set("page", strconv.FormatUint(uint64(types.Value(r.Page)), 10))
	}
	if year := types.Value(r.Year); year != "" {
		response.Set("year", year)
	}
	return response
}
