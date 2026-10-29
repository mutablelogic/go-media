package thumbnailer

import (
	"image"
	"math"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Stats are measurements of a frame, of how good a thumbnail it makes. They
// are measured on a small copy of the frame, so they don't depend much on
// the video's size
type Stats struct {
	Brightness    float64 // Mean luma (0.0-1.0)
	Contrast      float64 // Standard deviation of luma (0.0-0.5)
	Sharpness     float64 // Variance of the Laplacian of luma, in 8-bit units, which is low for a blurred frame
	Colourfulness float64 // How colourful the frame is, from grey (0.0) to vivid (1.0)
	Entropy       float64 // Entropy of the luma histogram, from uniform (0.0) to detailed (1.0)
	EdgeDensity   float64 // Fraction of pixels on a strong edge (0.0-1.0), which is high for text and graphics
	Blank         bool    // Whether the frame is black, white or uniform, such as during a fade
	Score         float64 // How good a thumbnail the frame makes (0.0-1.0), which is zero when it's blank
}

// histogram is the normalised histogram of a frame's luma and chroma, to
// compare frames
type histogram [3 * histogramBins]float64

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// Number of bins for each of Y, Cb and Cr in a frame's histogram
const histogramBins = 16

// Thresholds for a blank frame
const (
	minBrightness = 0.05 // Below is black
	maxBrightness = 0.95 // Above is white
	minContrast   = 0.03 // Below is uniform
	minEntropy    = 0.05 // Below is nearly one value, as graphics of a few colours are above it
)

// Thresholds and weights for a frame's score
const (
	edgeThreshold      = 64   // Gradient magnitude, in 8-bit units, of a strong edge
	sharpnessReference = 2000 // Laplacian variance of a sharp frame
	contrastReference  = 0.25 // Standard deviation of luma of a contrasty frame
	graphicsEdges      = 0.15 // Edge density above which a grey frame is likely text or graphics
	graphicsColour     = 0.1  // Colourfulness below which an edgy frame is likely text or graphics
	graphicsPenalty    = 0.7  // Score multiplier for text and graphics
	minExposure        = 0.02 // Video-range luma level (0.0-1.0) below which a frame scores nothing
	darkExposure       = 0.2  // Video-range luma level below which a frame's score is reduced
	brightExposure     = 0.8  // Video-range luma level above which a frame's score is reduced
)

// Letterbox and pillarbox bars, which are left out of a frame's stats
const (
	barLevel = 24   // Most mean luma, in 8-bit units, of a row or column of a bar, as video black is 16
	barRange = 12   // Most difference between its lightest and darkest pixels, as a bar is flat
	maxBar   = 0.25 // Most of a frame's height or width which each bar can be
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// measure returns the stats and histogram of a 4:2:0 frame
func measure(img *image.YCbCr) (Stats, histogram) {
	hist := histogramOf(img)
	if picture, ok := img.SubImage(pictureOf(img)).(*image.YCbCr); ok {
		img = picture
	}
	var stats Stats
	w, h := img.Rect.Dx(), img.Rect.Dy()
	if w < 3 || h < 3 {
		stats.Blank = true
		return stats, hist
	}
	origin := img.Rect.Min
	luma := func(x, y int) float64 { return float64(img.Y[img.YOffset(origin.X+x, origin.Y+y)]) }

	// Luma: mean, standard deviation and entropy
	var sum, sumSq float64
	var counts [256]int
	for y := range h {
		for x := range w {
			v := img.Y[img.YOffset(origin.X+x, origin.Y+y)]
			sum += float64(v)
			sumSq += float64(v) * float64(v)
			counts[v]++
		}
	}
	n := float64(w * h)
	mean := sum / n
	stats.Brightness = mean / 255
	stats.Contrast = math.Sqrt(max(0, sumSq/n-mean*mean)) / 255
	for _, c := range counts {
		if c > 0 {
			p := float64(c) / n
			stats.Entropy -= p * math.Log2(p) / 8
		}
	}

	// Sharpness, as the variance of the Laplacian, and edges, inside the
	// border
	var lapSum, lapSumSq float64
	var edges int
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			c := luma(x, y)
			l, r, u, d := luma(x-1, y), luma(x+1, y), luma(x, y-1), luma(x, y+1)
			lap := 4*c - l - r - u - d
			lapSum += lap
			lapSumSq += lap * lap
			if math.Abs(r-l)+math.Abs(d-u) > edgeThreshold {
				edges++
			}
		}
	}
	inner := float64((w - 2) * (h - 2))
	lapMean := lapSum / inner
	stats.Sharpness = max(0, lapSumSq/inner-lapMean*lapMean)
	stats.EdgeDensity = float64(edges) / inner

	// Chroma: colourfulness, from the spread and distance from grey of Cb and
	// Cr (after Hasler and Süsstrunk)
	var cbSum, crSum, cbSumSq, crSumSq float64
	cw, ch := (w+1)/2, (h+1)/2
	for y := range ch {
		for x := range cw {
			i := img.COffset(origin.X+2*x, origin.Y+2*y)
			cb, cr := float64(img.Cb[i])-128, float64(img.Cr[i])-128
			cbSum += cb
			crSum += cr
			cbSumSq += cb * cb
			crSumSq += cr * cr
		}
	}
	cn := float64(cw * ch)
	cbMean, crMean := cbSum/cn, crSum/cn
	spread := math.Sqrt(max(0, cbSumSq/cn-cbMean*cbMean) + max(0, crSumSq/cn-crMean*crMean))
	stats.Colourfulness = min(1, (spread+0.3*math.Hypot(cbMean, crMean))/64)

	stats.Blank = stats.Brightness < minBrightness || stats.Brightness > maxBrightness ||
		stats.Contrast < minContrast || stats.Entropy < minEntropy
	stats.Score = score(stats)
	return stats, hist
}

// score returns how good a thumbnail a frame makes, which is zero when it's
// blank, and lower for text and graphics
func score(s Stats) float64 {
	if s.Blank {
		return 0
	}
	sharpness := min(1, math.Log1p(s.Sharpness)/math.Log1p(sharpnessReference))
	contrast := min(1, s.Contrast/contrastReference)
	result := 0.4*sharpness + 0.25*contrast + 0.2*s.Colourfulness + 0.15*s.Entropy
	if s.EdgeDensity > graphicsEdges && s.Colourfulness < graphicsColour {
		result *= graphicsPenalty
	}
	return result * exposure(s.Brightness)
}

// exposure returns how well a frame of a brightness is exposed, from 0, too
// dark or bright, to 1, from its luma level in video range, where black is
// 16 and white is 235, so a dark frame, such as a silhouette, scores less,
// however sharp and contrasty its few bright parts make it
func exposure(brightness float64) float64 {
	level := (brightness*255 - 16) / 219
	switch {
	case level < darkExposure:
		return max(0, (level-minExposure)/(darkExposure-minExposure))
	case level > brightExposure:
		return max(0, (1-level)/(1-brightExposure))
	}
	return 1
}

// histogramOf returns the normalised histogram of the luma and chroma of a
// frame, as a whole, so frames are compared alike
func histogramOf(img *image.YCbCr) histogram {
	var hist histogram
	w, h := img.Rect.Dx(), img.Rect.Dy()
	cw, ch := (w+1)/2, (h+1)/2
	if w == 0 || h == 0 {
		return hist
	}
	origin := img.Rect.Min
	for y := range h {
		for x := range w {
			hist[int(img.Y[img.YOffset(origin.X+x, origin.Y+y)])*histogramBins/256]++
		}
	}
	for y := range ch {
		for x := range cw {
			i := img.COffset(origin.X+2*x, origin.Y+2*y)
			hist[histogramBins+int(img.Cb[i])*histogramBins/256]++
			hist[2*histogramBins+int(img.Cr[i])*histogramBins/256]++
		}
	}
	n, cn := float64(w*h), float64(cw*ch)
	for i := range histogramBins {
		hist[i] /= n
		hist[histogramBins+i] /= cn
		hist[2*histogramBins+i] /= cn
	}
	return hist
}

// pictureOf returns the part of a frame which is its picture, without
// letterbox bars at its top and bottom, or pillarbox bars at its sides,
// which are rows or columns which are flat and nearly black, at most maxBar
// of it each, and at an even offset, for its chroma
func pictureOf(img *image.YCbCr) image.Rectangle {
	r := img.Rect
	w, h := r.Dx(), r.Dy()
	bar := func(x0, y0, dx, dy, n int) bool {
		var sum int
		lo, hi := 255, 0
		for i := range n {
			v := int(img.Y[img.YOffset(x0+i*dx, y0+i*dy)])
			sum += v
			lo, hi = min(lo, v), max(hi, v)
		}
		return sum <= barLevel*n && hi-lo <= barRange
	}
	maxRows, maxCols := int(float64(h)*maxBar), int(float64(w)*maxBar)
	top, bottom, left, right := 0, 0, 0, 0
	for top < maxRows && bar(r.Min.X, r.Min.Y+top, 1, 0, w) {
		top++
	}
	for bottom < maxRows && bar(r.Min.X, r.Max.Y-1-bottom, 1, 0, w) {
		bottom++
	}
	for left < maxCols && bar(r.Min.X+left, r.Min.Y, 0, 1, h) {
		left++
	}
	for right < maxCols && bar(r.Max.X-1-right, r.Min.Y, 0, 1, h) {
		right++
	}
	return image.Rect(r.Min.X+(left+1)&^1, r.Min.Y+(top+1)&^1, r.Max.X-right, r.Max.Y-bottom)
}

// distance returns how different two frames' histograms are, from the same
// (0.0) to nothing in common (1.0)
func distance(a, b histogram) float64 {
	var sum float64
	for i := range a {
		sum += math.Abs(a[i] - b[i])
	}
	// Each channel's difference is at most two, and there are three
	return sum / 6
}
