package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parameters_001(t *testing.T) {
	assert := assert.New(t)

	for ch := 1; ch < 8; ch++ {
		params, err := newAudioParameters(ch, "flt", 44100)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(params)
		t.Log(params)
	}

}

func Test_parameters_002(t *testing.T) {
	assert := assert.New(t)

	for ch := 1; ch < 8; ch++ {
		params, err := newVideoParameters("100x100", "rgba", 25)
		if !assert.NoError(err) {
			t.FailNow()
		}
		assert.NotNil(params)
		t.Log(params)
	}

}
