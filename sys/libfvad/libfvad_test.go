package libfvad_test

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/libfvad"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_Libfvad_000(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("Unexpected nil return from NewContext")
	}
	ctx.Free()
}

func Test_Libfvad_001(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	for mode := MODE_QUALITY; mode <= MODE_VERY_AGGRESSIVE; mode++ {
		if err := ctx.SetMode(mode); err != nil {
			t.Error(mode, err)
		}
	}
	for _, mode := range []Mode{-1, 4} {
		if err := ctx.SetMode(mode); err == nil {
			t.Error("Expected error for", mode)
		}
	}
}

func Test_Libfvad_002(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	for _, rate := range []int{8000, 16000, 32000, 48000} {
		if err := ctx.SetSampleRate(rate); err != nil {
			t.Error(rate, err)
		}
	}
	for _, rate := range []int{0, 11025, 22050, 44100} {
		if err := ctx.SetSampleRate(rate); err == nil {
			t.Error("Expected error for", rate)
		}
	}
}

func Test_Libfvad_003(t *testing.T) {
	ctx := NewContext()
	defer ctx.Free()
	if err := ctx.SetSampleRate(16000); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{160, 320, 480} {
		if voice, err := ctx.Process(make([]int16, n)); err != nil {
			t.Error(n, err)
		} else if voice {
			t.Error("Expected no voice in silence for", n)
		}
	}
	for _, n := range []int{0, 80, 240, 640} {
		if _, err := ctx.Process(make([]int16, n)); err == nil {
			t.Error("Expected error for", n)
		}
	}
}

// Test_Libfvad_004 reproduces libfvad's own test, tests/wavtest.sh, which
// compares the output of examples/fvadwav with testdata/wavtest.expect
func Test_Libfvad_004(t *testing.T) {
	expect, err := readExpect(filepath.Join("testdata", "wavtest.expect"))
	if err != nil {
		t.Fatal(err)
	}
	for _, khz := range []int{8, 16, 32, 48} {
		name := fmt.Sprintf("audio_tiny%d.wav", khz)
		rate, samples, err := readWav(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		for mode := MODE_QUALITY; mode <= MODE_VERY_AGGRESSIVE; mode++ {
			for _, ms := range []int{10, 20, 30} {
				title := fmt.Sprintf("%s; mode %d; %d ms", name, mode, ms)
				t.Run(title, func(t *testing.T) {
					got, err := fvadwav(rate, samples, mode, rate/1000*ms)
					if err != nil {
						t.Fatal(err)
					}
					if want, exists := expect[title]; !exists {
						t.Fatal("No expected result")
					} else if got != want {
						t.Errorf("got:\n%s\nwant:\n%s", got, want)
					}
				})
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// fvadwav returns the summary which libfvad's examples/fvadwav prints
func fvadwav(rate int, samples []int16, mode Mode, framelen int) (string, error) {
	ctx := NewContext()
	defer ctx.Free()
	if err := ctx.SetMode(mode); err != nil {
		return "", err
	}
	if err := ctx.SetSampleRate(rate); err != nil {
		return "", err
	}

	// fvadwav reads samples as doubles, which libsndfile normalizes by 0x8000,
	// and converts them back to int16 by multiplying by INT16_MAX
	frame := make([]int16, framelen)
	var frames, segments [2]int
	prev := -1
	for i := 0; i+framelen <= len(samples); i += framelen {
		for j := range frame {
			frame[j] = int16(float64(samples[i+j]) / 0x8000 * 0x7FFF)
		}
		voice, err := ctx.Process(frame)
		if err != nil {
			return "", err
		}
		res := 0
		if voice {
			res = 1
		}
		frames[res]++
		if prev != res {
			segments[res]++
		}
		prev = res
	}

	ratio := func(a, b int) float64 {
		if b == 0 {
			return 0
		}
		return float64(a) / float64(b)
	}
	total := frames[0] + frames[1]
	return fmt.Sprintf(
		"voice detected in %d of %d frames (%.2f%%)\n"+
			"%d voice segments, average length %.2f frames\n"+
			"%d non-voice segments, average length %.2f frames",
		frames[1], total, 100*ratio(frames[1], total),
		segments[1], ratio(frames[1], segments[1]),
		segments[0], ratio(frames[0], segments[0]),
	), nil
}

// readExpect returns the results in wavtest.expect, keyed by title
func readExpect(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]string)
	var title string
	var lines []string
	flush := func() {
		if title != "" {
			result[title] = strings.Join(lines, "\n")
		}
		title, lines = "", nil
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "=== ") && strings.HasSuffix(line, " ==="):
			flush()
			title = strings.TrimSuffix(strings.TrimPrefix(line, "=== "), " ===")
		case line != "":
			lines = append(lines, line)
		}
	}
	flush()
	return result, scanner.Err()
}

// readWav returns the sample rate and samples of a mono 16-bit PCM wav file
func readWav(path string) (int, []int16, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return 0, nil, fmt.Errorf("%s: not a wav file", path)
	}
	var rate int
	for chunk := data[12:]; len(chunk) >= 8; {
		id, size := string(chunk[0:4]), int(binary.LittleEndian.Uint32(chunk[4:8]))
		if 8+size > len(chunk) {
			return 0, nil, fmt.Errorf("%s: truncated %q chunk", path, id)
		}
		body := chunk[8 : 8+size]
		switch id {
		case "fmt ":
			if size < 16 {
				return 0, nil, fmt.Errorf("%s: short fmt chunk", path)
			}
			format, channels, bits := binary.LittleEndian.Uint16(body[0:2]), binary.LittleEndian.Uint16(body[2:4]), binary.LittleEndian.Uint16(body[14:16])
			if format != 1 || channels != 1 || bits != 16 {
				return 0, nil, fmt.Errorf("%s: not mono 16-bit PCM", path)
			}
			rate = int(binary.LittleEndian.Uint32(body[4:8]))
		case "data":
			if rate == 0 {
				return 0, nil, fmt.Errorf("%s: data chunk before fmt chunk", path)
			}
			samples := make([]int16, size/2)
			for i := range samples {
				samples[i] = int16(binary.LittleEndian.Uint16(body[i*2:]))
			}
			return rate, samples, nil
		}
		// Chunks are padded to an even size
		chunk = chunk[8+size+size%2:]
	}
	return 0, nil, fmt.Errorf("%s: no data chunk", path)
}
