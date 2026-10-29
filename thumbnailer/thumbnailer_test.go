package thumbnailer_test

import (
	"context"
	"image"
	"image/color"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
	thumbnailer "github.com/mutablelogic/go-media/thumbnailer"
	writer "github.com/mutablelogic/go-media/writer"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

///////////////////////////////////////////////////////////////////////////////
// HELPERS

const (
	testWidth  = 320
	testHeight = 240
	testRate   = 25
)

// scene is a part of a test video, drawn by draw for each frame
type scene struct {
	duration time.Duration
	draw     func(img *image.NRGBA, n int)
}

// random returns a pseudo-random number in [0, 256) for i
func random(i int) uint8 {
	x := uint32(i)*1103515245 + 12345
	x ^= x >> 16
	x *= 0x45d9f3b
	x ^= x >> 16
	return uint8(x)
}

// textured draws a moving gradient of a colour, with noise, so it's sharp
func textured(c color.NRGBA) func(*image.NRGBA, int) {
	return func(img *image.NRGBA, n int) {
		b := img.Bounds()
		for y := range b.Dy() {
			for x := range b.Dx() {
				// The gradient and its noise move together, as texture does
				sx := (x + 2*n) % b.Dx()
				v := float64(sx) / float64(b.Dx())
				noise := float64(random(sx+y*b.Dx())) / 255
				f := 0.5 + 0.3*v + 0.2*noise
				img.SetNRGBA(x, y, color.NRGBA{uint8(float64(c.R) * f), uint8(float64(c.G) * f), uint8(float64(c.B) * f), 0xFF})
			}
		}
	}
}

// checks draws a moving checkerboard of two colours
func checks(a, b color.NRGBA) func(*image.NRGBA, int) {
	return func(img *image.NRGBA, n int) {
		for y := range img.Bounds().Dy() {
			for x := range img.Bounds().Dx() {
				if ((x+n)/20+y/20)%2 == 0 {
					img.SetNRGBA(x, y, a)
				} else {
					img.SetNRGBA(x, y, b)
				}
			}
		}
	}
}

// black draws a black frame
func black(img *image.NRGBA, _ int) {
	for i := range img.Pix {
		img.Pix[i] = 0
	}
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 0xFF
	}
}

// writeVideo writes a video of scenes, with mpeg4's default of a keyframe
// every 12 frames, in a format such as mp4 or mpegts
func writeVideo(t *testing.T, format string, scenes ...scene) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test."+format)
	p, err := profile.NewVideoProfile("mpeg4")
	require.NoError(t, err)
	require.NoError(t, p.Set(profile.OptionWidth, uint64(testWidth)))
	require.NoError(t, p.Set(profile.OptionHeight, uint64(testHeight)))
	require.NoError(t, p.Set(profile.OptionPixelFormat, "yuv420p"))
	require.NoError(t, p.Set(profile.OptionFrameRate, float64(testRate)))
	require.NoError(t, p.Set(profile.OptionVideoBitrate, uint64(2000000)))

	w, err := writer.Create(&url.URL{Path: path}, profile.OutputWithName(format), writer.WithProfile(0, p))
	require.NoError(t, err)
	img := image.NewNRGBA(image.Rect(0, 0, testWidth, testHeight))
	f, err := frame.NewVideoFrame(0)
	require.NoError(t, err)
	defer f.Close()
	var n int
	for _, s := range scenes {
		for range int(s.duration.Seconds() * testRate) {
			s.draw(img, n)
			require.NoError(t, f.FromImage(img))
			require.NoError(t, w.Encode(f))
			n++
		}
	}
	require.NoError(t, w.Flush(0))
	require.NoError(t, w.Close())
	return path
}

// testVideo returns an mp4 video of three scenes, a cut between the first
// two, and a fade through black before the third
func testVideo(t *testing.T) string {
	return testVideoFormat(t, "mp4")
}

func testVideoFormat(t *testing.T, format string) string {
	return writeVideo(t, format,
		scene{3 * time.Second, textured(color.NRGBA{255, 80, 40, 255})},
		scene{3 * time.Second, checks(color.NRGBA{30, 60, 220, 255}, color.NRGBA{250, 230, 60, 255})},
		scene{time.Second, black},
		scene{3 * time.Second, textured(color.NRGBA{40, 220, 90, 255})},
	)
}

func testDataPath(t *testing.T, name string) string {
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "..", "etc", "test", name)
}

///////////////////////////////////////////////////////////////////////////////
// TESTS

func TestThumbnailer_Scenes(t *testing.T) {
	path := testVideo(t)
	for _, mode := range []thumbnailer.Mode{thumbnailer.ModeAuto, thumbnailer.ModeKeyframes, thumbnailer.ModeSeek} {
		t.Run(mode.String(), func(t *testing.T) {
			th, err := thumbnailer.Open(path, thumbnailer.WithMode(mode), thumbnailer.WithMaxSamples(40))
			require.NoError(t, err)
			defer th.Close()

			a, err := th.Analyse(context.Background())
			require.NoError(t, err)
			t.Logf("%v: %d samples, %d scenes", a.Mode, len(a.Samples), len(a.Scenes))
			for _, s := range a.Scenes {
				t.Logf("  %v %v+%v samples=%d best=%d", s.Transition, s.Start, s.Duration, s.Count, s.Best)
			}
			if mode == thumbnailer.ModeAuto {
				// A keyframe every 12 frames is about two each second
				assert.Equal(t, thumbnailer.ModeKeyframes, a.Mode)
			} else {
				assert.Equal(t, mode, a.Mode)
			}

			// The scenes start at about 0s, 3s (a cut) and 7s (a fade)
			require.Equal(t, 3, len(a.Scenes))
			assert.Equal(t, thumbnailer.TransitionStart, a.Scenes[0].Transition)
			assert.Equal(t, thumbnailer.TransitionCut, a.Scenes[1].Transition)
			assert.Equal(t, thumbnailer.TransitionFade, a.Scenes[2].Transition)
			assert.InDelta(t, 3, a.Scenes[1].Start.Seconds(), 0.6)
			assert.InDelta(t, 7, a.Scenes[2].Start.Seconds(), 0.6)
			for i, s := range a.Scenes {
				assert.GreaterOrEqualf(t, s.Best, 0, "scene %d", i)
			}

			// The black frames are blank
			var blank int
			for _, s := range a.Samples {
				if s.Time >= 6200*time.Millisecond && s.Time < 6800*time.Millisecond {
					assert.True(t, s.Stats.Blank, s.Time)
				}
				if s.Stats.Blank {
					blank++
				}
			}
			assert.Greater(t, blank, 0)
		})
	}
}

func TestThumbnailer_StartTime(t *testing.T) {
	// MPEG-TS starts its timestamps after a delay, but times are from the
	// first frame, so the scenes are at the same times as in mp4
	th, err := thumbnailer.Open(testVideoFormat(t, "mpegts"), thumbnailer.WithMode(thumbnailer.ModeSeek), thumbnailer.WithMaxSamples(40))
	require.NoError(t, err)
	defer th.Close()

	a, err := th.Analyse(context.Background())
	require.NoError(t, err)
	for _, s := range a.Scenes {
		t.Logf("  %v %v+%v", s.Transition, s.Start, s.Duration)
	}
	require.Equal(t, 3, len(a.Scenes))
	assert.InDelta(t, 3, a.Scenes[1].Start.Seconds(), 0.6)
	assert.InDelta(t, 7, a.Scenes[2].Start.Seconds(), 0.6)
	assert.Less(t, a.Samples[0].Time, time.Second)

	// A thumbnail is the frame at its time
	thumbs, err := th.Thumbnails(context.Background(), 3)
	require.NoError(t, err)
	require.Equal(t, 3, len(thumbs))
	r, g, _, _ := thumbs[2].Image.At(80, 60).RGBA()
	assert.Greater(t, g, r)
}

func TestThumbnailer_Thumbnails(t *testing.T) {
	th, err := thumbnailer.Open(testVideo(t), thumbnailer.WithSize(160, 0))
	require.NoError(t, err)
	defer th.Close()

	// One from each scene, in time order, at the size
	thumbs, err := th.Thumbnails(context.Background(), 3)
	require.NoError(t, err)
	require.Equal(t, 3, len(thumbs))
	for i, thumb := range thumbs {
		t.Logf("%v scene=%d score=%.2f", thumb.Time, thumb.Scene, thumb.Stats.Score)
		assert.Equal(t, i, thumb.Scene)
		assert.Equal(t, image.Rect(0, 0, 160, 120), thumb.Image.Bounds())
		assert.False(t, thumb.Stats.Blank)
		if i > 0 {
			assert.Greater(t, thumb.Time, thumbs[i-1].Time)
		}
	}

	// The thumbnail is the frame at its time: the first scene is red, the
	// last is green
	r, g, _, _ := thumbs[0].Image.At(80, 60).RGBA()
	assert.Greater(t, r, g)
	r, g, _, _ = thumbs[2].Image.At(80, 60).RGBA()
	assert.Greater(t, g, r)

	// More than there are scenes come from the longest scenes, far apart
	thumbs, err = th.Thumbnails(context.Background(), 6)
	require.NoError(t, err)
	assert.Equal(t, 6, len(thumbs))

	poster, err := th.Poster(context.Background())
	require.NoError(t, err)
	assert.False(t, poster.Stats.Blank)
}

func TestThumbnailer_Sample(t *testing.T) {
	// A real video, of 5.3s with few keyframes, so it's seeked
	th, err := thumbnailer.Open(testDataPath(t, "sample.mp4"))
	require.NoError(t, err)
	defer th.Close()

	a, err := th.Analyse(context.Background())
	require.NoError(t, err)
	t.Logf("%v: %d samples, %d scenes", a.Mode, len(a.Samples), len(a.Scenes))
	assert.Greater(t, len(a.Samples), 1)
	poster, err := th.Poster(context.Background())
	require.NoError(t, err)
	t.Logf("poster at %v, score %.2f, %v", poster.Time, poster.Stats.Score, poster.Image.Bounds())
	assert.Equal(t, image.Rect(0, 0, 640, 360), poster.Image.Bounds())
}

func TestThumbnailer_FrameBeforeAnalyse(t *testing.T) {
	// Analysis is of the whole video, even after Frame has moved the reader
	th, err := thumbnailer.Open(testVideo(t), thumbnailer.WithMode(thumbnailer.ModeKeyframes))
	require.NoError(t, err)
	defer th.Close()

	_, err = th.Frame(context.Background(), 8*time.Second)
	require.NoError(t, err)
	a, err := th.Analyse(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 3, len(a.Scenes))
	assert.Less(t, a.Samples[0].Time, time.Second)
}

func TestThumbnailer_Short(t *testing.T) {
	// A video shorter than the time between seeks has a sample
	th, err := thumbnailer.Open(writeVideo(t, "mp4", scene{200 * time.Millisecond, textured(color.NRGBA{255, 80, 40, 255})}), thumbnailer.WithMode(thumbnailer.ModeSeek))
	require.NoError(t, err)
	defer th.Close()

	a, err := th.Analyse(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, len(a.Samples))
	_, err = th.Poster(context.Background())
	assert.NoError(t, err)
}

func TestThumbnailer_Closed(t *testing.T) {
	// After closing, thumbnails are an error, even once analysed
	th, err := thumbnailer.Open(testVideo(t))
	require.NoError(t, err)
	_, err = th.Analyse(context.Background())
	require.NoError(t, err)
	require.NoError(t, th.Close())

	_, err = th.Thumbnails(context.Background(), 3)
	assert.Error(t, err)
	_, err = th.Poster(context.Background())
	assert.Error(t, err)
	_, err = th.Frame(context.Background(), time.Second)
	assert.Error(t, err)
}

func TestThumbnailer_Invalid(t *testing.T) {
	_, err := thumbnailer.Open(testDataPath(t, "sample.mp3"))
	assert.Error(t, err)
	_, err = thumbnailer.Open(testDataPath(t, "sample.mp4"), thumbnailer.WithSize(-1, 0))
	assert.Error(t, err)
	_, err = thumbnailer.Open(testDataPath(t, "sample.mp4"), thumbnailer.WithMode(thumbnailer.ModeSeek+1))
	assert.Error(t, err)
	_, err = thumbnailer.Open(testDataPath(t, "sample.mp4"), thumbnailer.WithStream(1))
	assert.Error(t, err)
}

func TestThumbnailer_NewReader(t *testing.T) {
	// A video read from a file, which is seeked, has the same scenes and
	// thumbnails as when it's opened
	path := testVideo(t)
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	th, err := thumbnailer.NewReader(f, thumbnailer.WithSize(160, 0))
	require.NoError(t, err)
	defer th.Close()
	a, err := th.Analyse(context.Background())
	require.NoError(t, err)
	require.Equal(t, 3, len(a.Scenes))

	opened, err := thumbnailer.Open(path, thumbnailer.WithSize(160, 0))
	require.NoError(t, err)
	defer opened.Close()
	b, err := opened.Analyse(context.Background())
	require.NoError(t, err)
	for i := range a.Scenes {
		assert.Equal(t, b.Scenes[i].Start, a.Scenes[i].Start)
	}

	thumbs, err := th.Thumbnails(context.Background(), 3)
	require.NoError(t, err)
	require.Equal(t, 3, len(thumbs))
	r, g, _, _ := thumbs[2].Image.At(80, 60).RGBA()
	assert.Greater(t, g, r)
}

func TestThumbnailer_NewReaderNoSeek(t *testing.T) {
	// An MP4 video which can't be seeked can't be read, as its index is at
	// its end
	f, err := os.Open(testVideo(t))
	require.NoError(t, err)
	defer f.Close()
	_, err = thumbnailer.NewReader(struct{ io.Reader }{f})
	assert.Error(t, err)
}

func TestThumbnailer_SeekKeyframes(t *testing.T) {
	// A video with more keyframes than the maximum number of samples, such
	// as a film, has its keyframes sampled at evenly spaced times
	th, err := thumbnailer.Open(testVideo(t), thumbnailer.WithMode(thumbnailer.ModeAuto), thumbnailer.WithMaxSamples(8))
	require.NoError(t, err)
	defer th.Close()

	a, err := th.Analyse(context.Background())
	require.NoError(t, err)
	assert.Equal(t, thumbnailer.ModeKeyframes, a.Mode)
	assert.LessOrEqual(t, len(a.Samples), 8)
	assert.Greater(t, len(a.Samples), 4)
	for i, s := range a.Samples {
		assert.True(t, s.Keyframe)
		if i > 0 {
			assert.Greater(t, s.Time, a.Samples[i-1].Time)
		}
	}

	// The thumbnails are still from different scenes
	thumbs, err := th.Thumbnails(context.Background(), 2)
	require.NoError(t, err)
	require.Equal(t, 2, len(thumbs))
	assert.NotEqual(t, thumbs[0].Scene, thumbs[1].Scene)
}
