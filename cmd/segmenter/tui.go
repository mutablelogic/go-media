package main

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	// Packages
	segmenter "github.com/mutablelogic/go-media/segmenter"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// display shows a meter of the audio level, and the class and stats of the
// last segment, at the bottom of a terminal, with the segments printed above
// it as they're output
type display struct {
	sync.Mutex
	w       io.Writer
	level   level              // the last level
	peak    float64            // the peak level, which is held for a while
	peakAt  time.Time          // when the peak was set
	last    *segmenter.Segment // the last segment, without its samples
	segment int                // number of the last segment
	lines   int                // number of lines drawn
	drawn   time.Time          // when the lines were last drawn
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	refresh    = 100 * time.Millisecond // How often the display is drawn
	meterWidth = 40                     // Width of the meter, in characters
	minDB      = -60                    // Level of the left of the meter, in dBFS
	peakHold   = time.Second            // How long the peak is shown for
)

// ANSI escape sequences
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
)

var classColors = map[segmenter.Class]string{
	segmenter.ClassSpeech:  green,
	segmenter.ClassMusic:   purple,
	segmenter.ClassOther:   yellow,
	segmenter.ClassSilence: dim,
	segmenter.ClassUnknown: blue,
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newDisplay(w io.Writer) *display {
	fmt.Fprint(w, "\033[?25l") // Hide the cursor
	return &display{w: w}
}

// Close removes the display, and shows the cursor again
func (d *display) Close() {
	d.Lock()
	defer d.Unlock()
	d.clear()
	fmt.Fprint(d.w, "\033[?25h")
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// update sets the level of the last frame, and draws the display if it's
// been long enough since it was last drawn
func (d *display) update(l level) {
	d.Lock()
	defer d.Unlock()
	d.level = l
	if now := time.Now(); l.peak >= d.peak || now.Sub(d.peakAt) > peakHold {
		d.peak, d.peakAt = l.peak, now
	}
	if time.Since(d.drawn) >= refresh {
		d.draw()
	}
}

// print prints a line above the display, for a segment if it's not nil
func (d *display) print(line string, s *segmenter.Segment) {
	d.Lock()
	defer d.Unlock()
	d.clear()
	fmt.Fprint(d.w, line)
	if s != nil {
		d.segment++
		d.last = &segmenter.Segment{Start: s.Start, Duration: s.Duration, Class: s.Class, Stats: s.Stats}
	}
	d.draw()
}

// clear removes the lines drawn
func (d *display) clear() {
	if d.lines > 0 {
		fmt.Fprintf(d.w, "\033[%dA\r\033[J", d.lines)
		d.lines = 0
	}
}

// draw draws the display, replacing the lines drawn before
func (d *display) draw() {
	l := d.level
	lines := []string{
		fmt.Sprintf("%s%10v%s  %s  %s RMS  %s peak", bold, l.time.Round(10*time.Millisecond), reset, meter(l.rms, d.peak), db(l.rms), db(d.peak)),
	}
	if s := d.last; s == nil {
		lines = append(lines, dim+"No segments yet"+reset)
	} else {
		color := classColors[s.Class]
		lines = append(lines, fmt.Sprintf("%s●%s %s%-8s%s  segment %d, %v to %v  %sspeech%s %3.0f%%  %slow energy%s %.2f  %szcr variation%s %.2f  %sflux%s %.2f",
			color, reset, color+bold, strings.ToUpper(className(s.Class)), reset, d.segment, round(s.Start), round(s.Start+s.Duration),
			dim, reset, 100*s.Stats.Speech, dim, reset, s.Stats.LowEnergy, dim, reset, s.Stats.ZCRVariation, dim, reset, s.Stats.Flux))
	}

	d.clear()
	for _, line := range lines {
		fmt.Fprintf(d.w, "\r\033[2K%s\n", line)
	}
	d.lines, d.drawn = len(lines), time.Now()
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// meter returns a meter of an RMS level, with a peak marker, coloured green,
// yellow and red towards the top
func meter(rms, peak float64) string {
	position := func(v float64) int {
		return max(0, min(meterWidth, int(math.Round((dbfs(v)-minDB)/-minDB*meterWidth))))
	}
	level, mark := position(rms), position(peak)

	var b strings.Builder
	b.WriteString("▕")
	for i := range meterWidth {
		var color string
		switch db := minDB + float64(i)*-minDB/meterWidth; {
		case db >= -6:
			color = red
		case db >= -18:
			color = yellow
		default:
			color = green
		}
		switch {
		case i < level:
			b.WriteString(color + "█" + reset)
		case i == mark-1:
			b.WriteString(color + "│" + reset)
		default:
			b.WriteString(dim + "·" + reset)
		}
	}
	b.WriteString("▏")
	return b.String()
}

// dbfs returns a level in dBFS
func dbfs(v float64) float64 {
	if v <= 0 {
		return math.Inf(-1)
	}
	return 20 * math.Log10(v)
}

// db returns a level in dBFS as text
func db(v float64) string {
	if v := dbfs(v); v > minDB {
		return fmt.Sprintf("%5.1f dBFS", v)
	}
	return fmt.Sprintf("<%d dBFS", minDB)
}
