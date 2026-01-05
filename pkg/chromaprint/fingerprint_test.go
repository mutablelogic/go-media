package chromaprint_test

import (
	"encoding/binary"
	"io"
	"os"
	"testing"
	"time"

	// Packages
	"github.com/mutablelogic/go-media/pkg/chromaprint"
	"github.com/stretchr/testify/assert"
)

const (
	// Test data
	testData1 = "../../etc/test/audio_22050_1ch_5m35.s16le.sw"
)

func Test_fingerprint_000(t *testing.T) {
	chromaprint.PrintVersion(os.Stderr)
}

func Test_fingerprint_001(t *testing.T) {
	assert := assert.New(t)
	fingerprint := chromaprint.New(22050, 1, 5*time.Minute+35*time.Second)
	assert.NotNil(fingerprint)
	assert.NoError(fingerprint.Close())
}

func Test_fingerprint_002(t *testing.T) {
	assert := assert.New(t)
	fingerprint := chromaprint.New(22050, 1, 5*time.Minute+35*time.Second)
	assert.NotNil(fingerprint)

	r, err := os.Open(testData1)
	if !assert.NoError(err) {
		t.SkipNow()
	}
	defer r.Close()

	buf := make([]int16, 1024)
	for {
		if err := binary.Read(r, binary.LittleEndian, buf); err == io.EOF {
			break
		}
		assert.NoError(err)
		n, err := fingerprint.Write(buf)
		assert.NoError(err)
		assert.NotZero(n)
	}
	str, err := fingerprint.Finish()
	assert.NoError(err)
	assert.NotEmpty(str)
	assert.NoError(fingerprint.Close())
}
