package thumbnailer

import (
	"testing"
	"time"

	// Packages
	assert "github.com/stretchr/testify/assert"
)

// samples returns a sample a second for each histogram bin, as a frame which
// is all that luma
func samples(bins ...int) []Sample {
	result := make([]Sample, len(bins))
	for i, bin := range bins {
		result[i] = Sample{Time: time.Duration(i) * time.Second, Stats: Stats{Score: 0.5}}
		result[i].hist[bin] = 1
	}
	return result
}

func TestDetect_Flash(t *testing.T) {
	// A flash, and the change back from it, isn't a scene change
	scenes := detect(samples(0, 0, 0, 1, 0, 0, 0), 0)
	assert.Equal(t, 1, len(scenes))

	// A cut is, after a flash or not
	scenes = detect(samples(0, 0, 1, 0, 0, 2, 2, 2), 0)
	if assert.Equal(t, 2, len(scenes)) {
		assert.Equal(t, TransitionCut, scenes[1].Transition)
		assert.Equal(t, 5, scenes[1].First)
	}
}
