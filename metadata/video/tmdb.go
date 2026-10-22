package video

import (
	"context"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	imageutil "github.com/mutablelogic/go-media/metadata/image"
	tmdbclient "github.com/mutablelogic/go-media/tmdb/httpclient"
)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

// tmdbCandidates is how many of the search's results are checked for a
// runtime which matches the stream's duration, when their titles aren't
// the stream's. Every result on the search's first page whose title is the
// stream's is checked, as a common title, such as Ambition, can be shared by
// many films, and the film can be far down the results.
const tmdbCandidates = 5

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// Add metadata handler for video files: looks up the episode of a TV
	// show, or the movie, named by the input's filename (see
	// gomedia.NamedReader) on TMDB. Requires a TMDB client (see WithTMDB);
	// does nothing without one. When the "artwork" namespace is requested,
	// the movie's or show's poster and backdrop, and an episode's still, are
	// downloaded too.
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
		if q.Name == "" {
			return nil, nil
		}

		// A file with a season and an episode, such as S01E02, is an episode
		// of a TV show, and otherwise it's a movie
		var entries map[string]gomedia.Metadata
		var images []tmdbImage
		var err error
		switch {
		case q.Episode != 0:
			entries, images, err = tvEntries(ctx, client, q, o.Duration())
		default:
			entries, images, err = movieEntries(ctx, client, q, o.Duration(), o.Year())
		}
		if err != nil || entries == nil {
			return nil, err
		}

		// Download its images, when artwork is requested
		if o.HasNamespace("artwork") {
			if err := addArtwork(ctx, client, entries, images...); err != nil {
				return nil, err
			}
		}
		return metadata.FilterMetadata(entries, o), nil
	}, "tmdb", "dc", "artwork")
}

////////////////////////////////////////////////////////////////////////////////
// TYPES

// tmdbImage is an image on TMDB, with the key of its artwork, its size, and
// its path, which is empty when there's none
type tmdbImage struct {
	key, size, path string
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// apostrophes are removed from a title, rather than separating words, so
// Schindler's List is Schindlers List, as a file name often has it
var apostrophes = strings.NewReplacer("'", "", "\u2019", "", "\u2018", "", "\u02bc", "")

// normalizeTitle returns the words of a title, in lowercase, separated by a
// space, without apostrophes or other punctuation, or a leading "the", which
// a file's name often leaves out, so The Hunt for Red October is Hunt for Red
// October. Other articles aren't removed, as they're part of other titles,
// such as A Beautiful Mind.
func normalizeTitle(title string) string {
	words := strings.FieldsFunc(apostrophes.Replace(strings.ToLower(title)), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	if len(words) > 1 && words[0] == "the" {
		words = words[1:]
	}
	for i, word := range words {
		if american, exists := americanSpellings[word]; exists {
			words[i] = american
		}
	}
	return strings.Join(words, " ")
}

// americanSpellings are the American spellings of words, by their British
// spellings, as TMDB has titles in American spelling, and its search doesn't
// find other spellings. They're whole words, as a rule, such as -our to -or,
// would change others, such as four, and words with another meaning in
// American English, such as cheque, aren't included, nor -ise, which British
// English spells -ize too.
var americanSpellings = map[string]string{
	"aeroplane": "airplane", "aeroplanes": "airplanes", "ageing": "aging",
	"aluminium": "aluminum", "analogue": "analog", "ardour": "ardor",
	"armour": "armor", "behaviour": "behavior", "behaviours": "behaviors",
	"calibre": "caliber", "cancelled": "canceled", "candour": "candor",
	"catalogue": "catalog", "centre": "center", "centres": "centers",
	"clamour": "clamor", "colour": "color", "coloured": "colored",
	"colourful": "colorful", "colours": "colors", "defence": "defense",
	"endeavour": "endeavor", "favour": "favor", "favourite": "favorite",
	"favourites": "favorites", "fervour": "fervor", "fibre": "fiber",
	"flavour": "flavor", "flavours": "flavors", "grey": "gray",
	"greys": "grays", "harbour": "harbor", "harbours": "harbors",
	"honour": "honor", "honourable": "honorable", "honours": "honors",
	"humour": "humor", "jewellery": "jewelry", "judgement": "judgment",
	"labour": "labor", "licence": "license", "litre": "liter",
	"lustre": "luster", "manoeuvre": "maneuver", "manoeuvres": "maneuvers",
	"marvellous": "marvelous", "meagre": "meager", "metre": "meter",
	"metres": "meters", "modelling": "modeling", "mould": "mold",
	"moustache": "mustache", "neighbour": "neighbor",
	"neighbourhood": "neighborhood", "neighbours": "neighbors", "odour": "odor",
	"offence": "offense", "parlour": "parlor", "plough": "plow",
	"pretence": "pretense", "programme": "program", "pyjamas": "pajamas",
	"rancour": "rancor", "rigour": "rigor", "rumour": "rumor",
	"rumours": "rumors", "sabre": "saber", "saviour": "savior",
	"savour": "savor", "sceptic": "skeptic", "sceptical": "skeptical",
	"sombre": "somber", "spectre": "specter", "splendour": "splendor",
	"theatre": "theater", "theatres": "theaters", "travelled": "traveled",
	"traveller": "traveler", "travellers": "travelers",
	"travelling": "traveling", "valour": "valor", "vapour": "vapor",
	"vigour": "vigor",
}

// wordRe matches the words of a name
var wordRe = regexp.MustCompile(`\p{L}+`)

// spellings returns the name, and then its American spelling, when it has a
// British spelling, keeping the case of the first letter of each word
func spellings(name string) []string {
	american := wordRe.ReplaceAllStringFunc(name, func(word string) string {
		replacement, exists := americanSpellings[strings.ToLower(word)]
		if !exists {
			return word
		} else if r := []rune(word); unicode.IsUpper(r[0]) {
			return strings.ToUpper(replacement[:1]) + replacement[1:]
		}
		return replacement
	})
	if american == name {
		return []string{name}
	}
	return []string{name, american}
}

// addArtwork adds the images, which are downloaded from TMDB, to the entries
// as artwork, resized as embedded artwork is. An image without a path isn't
// added, nor is one which can't be downloaded or decoded, unless the context
// is done.
func addArtwork(ctx context.Context, client *tmdbclient.Client, entries map[string]gomedia.Metadata, images ...tmdbImage) error {
	for _, image := range images {
		if image.path == "" {
			continue
		}
		data, err := client.Image(ctx, image.size, image.path)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
		if m, err := imageutil.ExtractArtwork(data, image.key); err == nil {
			entries[image.key] = m
		}
	}
	return nil
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
