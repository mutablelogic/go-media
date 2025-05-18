package ffmpeg_test

import (
	"fmt"
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avutil_dict_001(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	if !assert.NotNil(dict) {
		t.SkipNow()
	}
	assert.NoError(AVUtil_dict_set(dict, "a", "b", 0))
	assert.NoError(AVUtil_dict_set(dict, "b", "b", 0))

	t.Log(dict)

	keys := AVUtil_dict_keys(dict)
	assert.Equal(2, len(keys))

	entries := AVUtil_dict_entries(dict)
	assert.Equal(2, len(entries))

	AVUtil_dict_free(dict)
}

func Test_avutil_dict_002(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	if !assert.NotNil(dict) {
		t.SkipNow()
	}
	defer AVUtil_dict_free(dict)

	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		assert.NoError(AVUtil_dict_set(dict, key, value, 0))
	}

	t.Log(dict)
}
