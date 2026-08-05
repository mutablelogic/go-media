package tmdb_test

import (
	"testing"

	tmdb "github.com/mutablelogic/go-media/tmdb/httpclient"
)

func Test_query_000(t *testing.T) {
	tests := []struct {
		path string
		want tmdb.Query
	}{
		// Plain movie title, no year or release tags.
		{"Hunt for Red October.mp4", tmdb.Query{Name: "Hunt for Red October"}},

		// Year in parentheses.
		{"The Matrix (1999).mp4", tmdb.Query{Name: "The Matrix", Year: 1999}},

		// Dot-separated with year and release tags.
		{"The.Matrix.1999.1080p.BluRay.x264-GROUP.mkv", tmdb.Query{Name: "The Matrix", Year: 1999}},

		// Space-separated with year and release tags.
		{"Inception (2010) 1080p BluRay x264-GROUP.mp4", tmdb.Query{Name: "Inception", Year: 2010}},

		// Release tags but no year at all.
		{"Show Name 1080p WEBRip x264-GROUP.mkv", tmdb.Query{Name: "Show Name"}},

		// TV episode marker, dot-separated with release tags.
		{"Breaking.Bad.S01E01.720p.HDTV.x264-GROUP.mkv", tmdb.Query{Name: "Breaking Bad", Season: 1, Episode: 1}},

		// TV episode marker, lower case.
		{"the.wire.s04e05.mkv", tmdb.Query{Name: "the wire", Season: 4, Episode: 5}},

		// A leading year-shaped number that's actually part of the title
		// isn't mistaken for a release year.
		{"2001 A Space Odyssey (1968).mp4", tmdb.Query{Name: "2001 A Space Odyssey", Year: 1968}},

		// A hyphenated title isn't mistaken for a trailing release group.
		{"Spider-Man.mkv", tmdb.Query{Name: "Spider-Man"}},

		// A generic disc-rip filename with no usable title falls back to
		// the parent directory, which carries the title/year.
		{"/movies/The Matrix (1999)/VTS_01_1.VOB", tmdb.Query{Name: "The Matrix", Year: 1999}},

		// A directory that itself has no usable title (and the filename
		// didn't either) falls back to the filename-derived query as a
		// last resort, rather than a wrongly "confident" empty result.
		{"/movies/1/VTS_01_1.VOB", tmdb.Query{Name: "VTS 01 1"}},

		// A bare filename with no directory component: the fallback is
		// skipped rather than misreading "." as a parent directory.
		{"VTS_01_1.VOB", tmdb.Query{Name: "VTS 01 1"}},

		// The following cases are drawn from guessit's real-world test
		// fixture (https://github.com/guessit-io/guessit, MIT licensed),
		// trimmed to the fields we extract (name/year/season/episode).
		// guessit itself does much more (episode titles, codecs, release
		// groups, dates, ...); these are only the cases where its expected
		// title/year/season/episode line up with what our simpler,
		// regex-based parser produces.

		// A year embedded in an episode filename is stripped independently
		// of the episode marker.
		{"The.Flash.2014.S01E01.PREAIR.WEBRip.XviD-EVO.avi", tmdb.Query{Name: "The Flash", Year: 2014, Season: 1, Episode: 1}},

		// A two-digit season number.
		{"House.Hunters.International.S56E06.720p.hdtv.x264.mp4", tmdb.Query{Name: "House Hunters International", Season: 56, Episode: 6}},

		// An episode title following the episode marker is dropped along
		// with the rest of the release tags, not mistaken for the title.
		{"Hostages.S01E01.Pilot.for.Air.720p.WEB-DL.DD5.1.H.264-NTb.nfo", tmdb.Query{Name: "Hostages", Season: 1, Episode: 1}},

		// No real file extension (a bare release-tag suffix instead)
		// doesn't throw off the parse.
		{"some.series.S03E14.Title.Here.720p", tmdb.Query{Name: "some series", Season: 3, Episode: 14}},

		// An apostrophe in the title survives cleanup.
		{"Jack's.Show.S03E01.blah.1080p", tmdb.Query{Name: "Jack's Show", Season: 3, Episode: 1}},

		// A double-digit episode number, with a full path prefix.
		{"/media/live/A/Anger.Management.S02E82.720p.HDTV.X264-DIMENSION.mkv", tmdb.Query{Name: "Anger Management", Season: 2, Episode: 82}},

		// The filename has a good title but no year; the parent directory
		// (a "Title (Year)" convention) supplies the year alone, without
		// its title overriding the filename's.
		{"Movies/Fear and Loathing in Las Vegas (1998)/Fear.and.Loathing.in.Las.Vegas.720p.HDDVD.DTS.x264-ESiR.mkv", tmdb.Query{Name: "Fear and Loathing in Las Vegas", Year: 1998}},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			if got := tmdb.ExtractQuery(test.path); got != test.want {
				t.Errorf("ExtractQuery(%q) = %+v, want %+v", test.path, got, test.want)
			}
		})
	}
}
