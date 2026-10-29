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
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// measure returns the stats and histogram of a 4:2:0 frame
func measure(img *image.YCbCr) (Stats, histogram) {
	var stats Stats
	var hist histogram
	w, h := img.Rect.Dx(), img.Rect.Dy()
	if w < 3 || h < 3 {
		stats.Blank = true
		return stats, hist
	}
	origin := img.Rect.Min
	luma := func(x, y int) float64 { return float64(img.Y[img.YOffset(origin.X+x, origin.Y+y)]) }

	// Luma: mean, standard deviation, histogram and entropy
	var sum, sumSq float64
	var counts [256]int
	for y := range h {
		for x := range w {
			v := img.Y[img.YOffset(origin.X+x, origin.Y+y)]
			sum += float64(v)
			sumSq += float64(v) * float64(v)
			counts[v]++
			hist[int(v)*histogramBins/256]++
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
	// Cr (after Hasler and Süsstrunk), and histogram
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
			hist[histogramBins+int(img.Cb[i])*histogramBins/256]++
			hist[2*histogramBins+int(img.Cr[i])*histogramBins/256]++
		}
	}
	cn := float64(cw * ch)
	cbMean, crMean := cbSum/cn, crSum/cn
	spread := math.Sqrt(max(0, cbSumSq/cn-cbMean*cbMean) + max(0, crSumSq/cn-crMean*crMean))
	stats.Colourfulness = min(1, (spread+0.3*math.Hypot(cbMean, crMean))/64)

	// Normalise the histogram, so each channel sums to one
	for i := range histogramBins {
		hist[i] /= n
		hist[histogramBins+i] /= cn
		hist[2*histogramBins+i] /= cn
	}

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
	return result
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
