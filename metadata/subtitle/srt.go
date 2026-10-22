package subtitle

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	charmap "golang.org/x/text/encoding/charmap"
	unicode "golang.org/x/text/encoding/unicode"
	transform "golang.org/x/text/transform"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// cue is a subtitle shown from its start, with its lines of text
type cue struct {
	number     uint64
	start, end time.Duration
	lines      []string
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

// srtTiming matches the timing of a cue, such as 00:01:02,500 --> 00:01:04,000,
// whose start and end are their hours, minutes, seconds and milliseconds. A
// period is accepted for the comma, as some files have one, and a cue whose
// end can't be read has its start.
var srtTiming = regexp.MustCompile(`^\s*(\d+):(\d{1,2}):(\d{1,2})[,.](\d{1,3})\s*-->\s*(?:(\d+):(\d{1,2}):(\d{1,2})[,.](\d{1,3}))?`)

// srtIndex matches the number of a cue, on the line before its timing
var srtIndex = regexp.MustCompile(`^\s*\d+\s*$`)

// srtTags matches the styling of a cue's text: HTML tags, such as <i> and
// <font color="...">, and ASS override tags, such as {\an8}
var srtTags = regexp.MustCompile(`<[^>]*>|\{\\[^}]*\}`)

// markdownEscapes escapes the characters of a cue's text which are
// Markdown's
var markdownEscapes = strings.NewReplacer(
	`\`, `\\`, "*", `\*`, "_", `\_`, "`", "\\`", "[", `\[`, "]", `\]`, "<", `\<`, ">", `\>`, "#", `\#`,
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	// SubRip subtitles are text, so .srt refines content sniffed as text
	if err := metadata.AddTextExtensionType(".srt", "application/x-subrip"); err != nil {
		panic(err)
	}

	// Add metadata handler for SubRip subtitles, which returns their cues as
	// Markdown, when the markdown namespace is requested
	metadata.AddHandler(regexp.MustCompile(`^application/x-subrip$`), "srt", func(ctx context.Context, r io.Reader, o *metadata.Opts) ([]gomedia.Metadata, error) {
		if !o.HasNamespace(gomedia.MetaMarkdown) {
			return nil, nil
		}
		w := metadata.NewMarkdownWriter(o)
		return w.Metadata(writeSRT(ctx, r, w))
	}, gomedia.MetaMarkdown)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// writeSRT writes the cues of the SubRip subtitles as Markdown, a paragraph
// for each, which starts with its number, and its start and end times, such
// as **1 · 00:01:41 → 00:01:44**, without their styling. A cue's number is the file's, or
// the previous cue's and one, when the file doesn't number it. A cue whose
// text is the previous cue's isn't written again. Stops with the writer's
// error, once it has its part.
func writeSRT(ctx context.Context, r io.Reader, w io.StringWriter) error {
	scanner := bufio.NewScanner(utf16Reader(r))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var current *cue
	var previous, index string
	var number uint64
	first := true
	flush := func() error {
		if current == nil {
			return nil
		}
		c := current
		current = nil
		text := c.text()
		if text == "" || text == previous {
			return nil
		}
		previous = text
		times := timestamp(c.start)
		if c.end > 0 {
			times += " → " + timestamp(c.end)
		}
		paragraph := "**" + strconv.FormatUint(c.number, 10) + " · " + times + "** " + text + "\n"
		if !first {
			paragraph = "\n" + paragraph
		}
		first = false
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := w.WriteString(paragraph)
		return err
	}

	for line := 0; scanner.Scan(); line++ {
		text := decode(scanner.Bytes())
		if line == 0 {
			text = strings.TrimPrefix(text, "\uFEFF")
		}
		text = strings.TrimRight(text, " \t\r")
		if start, end, ok := parseTiming(text); ok {
			// A cue's number is on the line before its timing, which is the
			// last line of the previous cue, when there's no blank line
			if current != nil && len(current.lines) > 0 && srtIndex.MatchString(current.lines[len(current.lines)-1]) {
				index = current.lines[len(current.lines)-1]
				current.lines = current.lines[:len(current.lines)-1]
			}
			if err := flush(); err != nil {
				return err
			}
			if n, err := strconv.ParseUint(strings.TrimSpace(index), 10, 64); err == nil && srtIndex.MatchString(index) {
				number = n
			} else {
				number++
			}
			current, index = &cue{number: number, start: start, end: end}, ""
		} else if strings.TrimSpace(text) == "" {
			if err := flush(); err != nil {
				return err
			}
			index = ""
		} else if current != nil {
			current.lines = append(current.lines, text)
		} else {
			index = text
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return flush()
}

// text returns the cue's lines without their styling, escaped as Markdown,
// a line each, without lines which are empty
func (c *cue) text() string {
	lines := make([]string, 0, len(c.lines))
	for _, line := range c.lines {
		line = strings.TrimSpace(srtTags.ReplaceAllString(line, ""))
		if line != "" {
			lines = append(lines, markdownEscapes.Replace(line))
		}
	}
	return strings.Join(lines, "\n")
}

// parseTiming returns the start of a cue, from its timing
func parseTiming(line string) (time.Duration, time.Duration, bool) {
	match := srtTiming.FindStringSubmatch(line)
	if match == nil {
		return 0, 0, false
	}
	start, ok := parseTime(match[1:5])
	if !ok {
		return 0, 0, false
	}
	end, ok := parseTime(match[5:9])
	if !ok || end < start {
		end = 0
	}
	return start, end, true
}

// parseTime returns the time of its hours, minutes, seconds and
// milliseconds, or false when there's none
func parseTime(parts []string) (time.Duration, bool) {
	var values [4]int
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil {
			return 0, false
		}
		values[i] = value
	}
	ms := values[3]
	for n := len(parts[3]); n < 3; n++ {
		ms *= 10
	}
	return time.Duration(values[0])*time.Hour + time.Duration(values[1])*time.Minute + time.Duration(values[2])*time.Second + time.Duration(ms)*time.Millisecond, true
}

// timestamp returns the time as hours, minutes and seconds, such as 01:02:03
func timestamp(d time.Duration) string {
	seconds := int64(d / time.Second)
	return fmt.Sprintf("%02d:%02d:%02d", seconds/3600, seconds/60%60, seconds%60)
}

// utf16Reader returns a reader of the text as UTF-8, when it's UTF-16, which
// it is when it starts with UTF-16's byte order mark, in the byte order the
// mark is in, or otherwise the text
func utf16Reader(r io.Reader) io.Reader {
	buffered := bufio.NewReader(r)
	bom, _ := buffered.Peek(2)
	switch {
	case len(bom) == 2 && bom[0] == 0xFF && bom[1] == 0xFE:
		return transform.NewReader(buffered, unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder())
	case len(bom) == 2 && bom[0] == 0xFE && bom[1] == 0xFF:
		return transform.NewReader(buffered, unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM).NewDecoder())
	default:
		return buffered
	}
}

// decode returns the line as UTF-8: as it is, when it's UTF-8, or otherwise
// from Windows-1252, which older subtitles are often in
func decode(line []byte) string {
	if utf8.Valid(line) {
		return string(line)
	}
	if decoded, err := charmap.Windows1252.NewDecoder().Bytes(line); err == nil {
		return string(decoded)
	}
	return strings.ToValidUTF8(string(line), "�")
}
