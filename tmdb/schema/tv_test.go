package schema_test

import (
	"encoding/json"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/tmdb/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

func Test_TVSearchRequest_Query(t *testing.T) {
	// Only the fields which are set are parameters
	if got := (&schema.TVSearchRequest{Term: "Breaking Bad"}).Query().Encode(); got != "query=Breaking+Bad" {
		t.Errorf("query = %q", got)
	}
	req := schema.TVSearchRequest{
		Term:             "The Office",
		FirstAirDateYear: types.Ptr("2005"),
		IncludeAdult:     types.Ptr(false),
		Language:         types.Ptr("en-US"),
		Page:             types.Ptr(uint(2)),
		Year:             types.Ptr("2006"),
	}
	want := "first_air_date_year=2005&include_adult=false&language=en-US&page=2&query=The+Office&year=2006"
	if got := req.Query().Encode(); got != want {
		t.Errorf("query = %q, want %q", got, want)
	}
}

func Test_TV_Decode(t *testing.T) {
	// The responses of TMDB's TV endpoints, trimmed
	var search schema.TVSearchResponse
	if err := json.Unmarshal([]byte(`{"page":1,"results":[{"id":1396,"name":"Breaking Bad","original_name":"Breaking Bad","first_air_date":"2008-01-20","origin_country":["US"],"genre_ids":[18,80],"popularity":410.6}],"total_pages":1,"total_results":5}`), &search); err != nil {
		t.Fatal(err)
	}
	if len(search.Results) != 1 || search.Results[0].ID != 1396 || search.Results[0].Name != "Breaking Bad" || search.Results[0].FirstAirDate != "2008-01-20" || search.TotalResults != 5 {
		t.Errorf("search = %+v", search)
	}

	// A show whose episodes' runtimes aren't known, as TMDB often has
	var show schema.TVShowDetails
	if err := json.Unmarshal([]byte(`{"id":1396,"name":"Breaking Bad","episode_run_time":[],"number_of_seasons":5,"number_of_episodes":62,"last_air_date":"2013-09-29","status":"Ended"}`), &show); err != nil {
		t.Fatal(err)
	}
	if show.Name != "Breaking Bad" || len(show.EpisodeRunTime) != 0 || show.NumberOfSeasons != 5 || show.NumberOfEpisodes != 62 || show.Status != "Ended" {
		t.Errorf("show = %+v", show)
	}

	// An episode, and one whose runtime is null, which is zero
	var episode schema.TVEpisode
	if err := json.Unmarshal([]byte(`{"id":62086,"name":"Cat's in the Bag...","air_date":"2008-01-27","season_number":1,"episode_number":2,"runtime":49}`), &episode); err != nil {
		t.Fatal(err)
	}
	if episode.Name != "Cat's in the Bag..." || episode.SeasonNumber != 1 || episode.EpisodeNumber != 2 || episode.Runtime != 49 {
		t.Errorf("episode = %+v", episode)
	}
	var unknown schema.TVEpisode
	if err := json.Unmarshal([]byte(`{"id":1,"runtime":null}`), &unknown); err != nil {
		t.Fatal(err)
	}
	if unknown.Runtime != 0 {
		t.Errorf("runtime = %d, want 0", unknown.Runtime)
	}

	var ids schema.ExternalIDs
	if err := json.Unmarshal([]byte(`{"id":62086,"imdb_id":"tt1054724","tvdb_id":349233,"wikidata_id":null}`), &ids); err != nil {
		t.Fatal(err)
	}
	if ids.IMDbID != "tt1054724" || ids.TVDBID != 349233 || ids.WikiData != "" {
		t.Errorf("ids = %+v", ids)
	}
}
