package ffmpeg

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST BASIC OPERATIONS

func Test_avutil_dict_alloc_free(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	assert.NotNil(dict)
	assert.Equal(0, AVUtil_dict_count(dict))

	AVUtil_dict_free(dict)
}

func Test_avutil_dict_free_nil(t *testing.T) {
	// Should not panic
	AVUtil_dict_free(nil)
}

func Test_avutil_dict_set_and_get(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Set a key-value pair
	assert.NoError(AVUtil_dict_set(dict, "key1", "value1", 0))
	assert.Equal(1, AVUtil_dict_count(dict))

	// Get the entry
	entry := AVUtil_dict_get(dict, "key1", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)
	assert.Equal("key1", entry.Key())
	assert.Equal("value1", entry.Value())
}

func Test_avutil_dict_set_multiple(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "a", "value_a", 0))
	assert.NoError(AVUtil_dict_set(dict, "b", "value_b", 0))
	assert.NoError(AVUtil_dict_set(dict, "c", "value_c", 0))

	assert.Equal(3, AVUtil_dict_count(dict))
}

func Test_avutil_dict_set_overwrite(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Set initial value
	assert.NoError(AVUtil_dict_set(dict, "key", "value1", 0))
	entry := AVUtil_dict_get(dict, "key", nil, AV_DICT_MATCH_CASE)
	assert.Equal("value1", entry.Value())

	// Overwrite
	assert.NoError(AVUtil_dict_set(dict, "key", "value2", 0))
	assert.Equal(1, AVUtil_dict_count(dict))
	entry = AVUtil_dict_get(dict, "key", nil, AV_DICT_MATCH_CASE)
	assert.Equal("value2", entry.Value())
}

func Test_avutil_dict_dont_overwrite(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Set initial value
	assert.NoError(AVUtil_dict_set(dict, "key", "value1", 0))

	// Try to overwrite with DONT_OVERWRITE flag - FFmpeg doesn't return error but preserves original
	err := AVUtil_dict_set(dict, "key", "value2", AV_DICT_DONT_OVERWRITE)
	assert.NoError(err)

	// Original value should remain unchanged
	entry := AVUtil_dict_get(dict, "key", nil, AV_DICT_MATCH_CASE)
	assert.Equal("value1", entry.Value())
}

func Test_avutil_dict_append(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Set initial value
	assert.NoError(AVUtil_dict_set(dict, "key", "value1", 0))

	// Append to existing key
	assert.NoError(AVUtil_dict_set(dict, "key", " value2", AV_DICT_APPEND))

	entry := AVUtil_dict_get(dict, "key", nil, AV_DICT_MATCH_CASE)
	assert.Equal("value1 value2", entry.Value())
}

////////////////////////////////////////////////////////////////////////////////
// TEST DELETE

func Test_avutil_dict_delete(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Add multiple entries
	assert.NoError(AVUtil_dict_set(dict, "key1", "value1", 0))
	assert.NoError(AVUtil_dict_set(dict, "key2", "value2", 0))
	assert.Equal(2, AVUtil_dict_count(dict))

	// Delete one entry
	_, err := AVUtil_dict_delete(dict, "key1")
	assert.NoError(err)
	assert.Equal(1, AVUtil_dict_count(dict))

	// Verify key1 is gone
	entry := AVUtil_dict_get(dict, "key1", nil, AV_DICT_MATCH_CASE)
	assert.Nil(entry)

	// Verify key2 still exists
	entry = AVUtil_dict_get(dict, "key2", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)
}

func Test_avutil_dict_delete_nonexistent(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Delete from empty dictionary
	_, err := AVUtil_dict_delete(dict, "nonexistent")
	assert.NoError(err)
}

func Test_avutil_dict_delete_nil(t *testing.T) {
	result, err := AVUtil_dict_delete(nil, "key")
	assert.Nil(t, result)
	assert.NoError(t, err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST COPY

func Test_avutil_dict_copy(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "key1", "value1", 0))
	assert.NoError(AVUtil_dict_set(dict, "key2", "value2", 0))

	// Copy the dictionary
	dict2, err := AVUtil_dict_copy(dict, 0)
	assert.NoError(err)
	assert.NotNil(dict2)
	defer AVUtil_dict_free(dict2)

	// Verify copy has same entries
	assert.Equal(AVUtil_dict_count(dict), AVUtil_dict_count(dict2))

	entry := AVUtil_dict_get(dict2, "key1", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)
	assert.Equal("value1", entry.Value())
}

func Test_avutil_dict_copy_nil(t *testing.T) {
	dict2, err := AVUtil_dict_copy(nil, 0)
	assert.Nil(t, dict2)
	assert.NoError(t, err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST KEYS AND ENTRIES

func Test_avutil_dict_keys(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "apple", "red", 0))
	assert.NoError(AVUtil_dict_set(dict, "banana", "yellow", 0))
	assert.NoError(AVUtil_dict_set(dict, "cherry", "red", 0))

	keys := AVUtil_dict_keys(dict)
	assert.Equal(3, len(keys))
	assert.Contains(keys, "apple")
	assert.Contains(keys, "banana")
	assert.Contains(keys, "cherry")
}

func Test_avutil_dict_keys_empty(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	keys := AVUtil_dict_keys(dict)
	assert.Equal(0, len(keys))
}

func Test_avutil_dict_entries(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "a", "value_a", 0))
	assert.NoError(AVUtil_dict_set(dict, "b", "value_b", 0))

	entries := AVUtil_dict_entries(dict)
	assert.Equal(2, len(entries))

	for _, entry := range entries {
		assert.NotNil(entry)
		if entry.Key() == "a" {
			assert.Equal("value_a", entry.Value())
		} else if entry.Key() == "b" {
			assert.Equal("value_b", entry.Value())
		}
	}
}

func Test_avutil_dict_entries_nil(t *testing.T) {
	entries := AVUtil_dict_entries(nil)
	assert.Nil(t, entries)
}

////////////////////////////////////////////////////////////////////////////////
// TEST PARSE STRING

func Test_avutil_dict_parse_string(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Parse a string like "key1=value1:key2=value2"
	err := AVUtil_dict_parse_string(dict, "key1=value1:key2=value2", "=", ":", 0)
	assert.NoError(err)
	assert.Equal(2, AVUtil_dict_count(dict))

	entry := AVUtil_dict_get(dict, "key1", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)
	assert.Equal("value1", entry.Value())
}

func Test_avutil_dict_parse_string_custom_separators(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Parse with custom separators: "key1:value1,key2:value2"
	err := AVUtil_dict_parse_string(dict, "key1:value1,key2:value2", ":", ",", 0)
	assert.NoError(err)
	assert.Equal(2, AVUtil_dict_count(dict))
}

func Test_avutil_dict_parse_string_nil(t *testing.T) {
	err := AVUtil_dict_parse_string(nil, "key=value", "=", ":", 0)
	assert.NoError(t, err)
}

////////////////////////////////////////////////////////////////////////////////
// TEST COUNT

func Test_avutil_dict_count(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.Equal(0, AVUtil_dict_count(dict))

	assert.NoError(AVUtil_dict_set(dict, "key1", "value1", 0))
	assert.Equal(1, AVUtil_dict_count(dict))

	assert.NoError(AVUtil_dict_set(dict, "key2", "value2", 0))
	assert.Equal(2, AVUtil_dict_count(dict))
}

func Test_avutil_dict_count_nil(t *testing.T) {
	count := AVUtil_dict_count(nil)
	assert.Equal(t, 0, count)
}

////////////////////////////////////////////////////////////////////////////////
// TEST GET WITH FLAGS

func Test_avutil_dict_get_case_sensitive(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "Key", "value", 0))

	// Case-sensitive search
	entry := AVUtil_dict_get(dict, "Key", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)

	entry = AVUtil_dict_get(dict, "key", nil, AV_DICT_MATCH_CASE)
	assert.Nil(entry)
}

func Test_avutil_dict_get_case_insensitive(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "Key", "value", 0))

	// Case-insensitive search (default)
	entry := AVUtil_dict_get(dict, "key", nil, 0)
	assert.NotNil(entry)
	assert.Equal("value", entry.Value())
}

func Test_avutil_dict_get_ignore_suffix(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "video:codec", "h264", 0))

	// Search with IGNORE_SUFFIX flag
	entry := AVUtil_dict_get(dict, "video", nil, AV_DICT_IGNORE_SUFFIX)
	assert.NotNil(entry)
	assert.Equal("video:codec", entry.Key())
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avutil_dict_marshal_json(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "key1", "value1", 0))
	assert.NoError(AVUtil_dict_set(dict, "key2", "value2", 0))

	// Marshal to JSON
	data, err := json.Marshal(dict)
	assert.NoError(err)
	assert.NotEmpty(data)

	// Unmarshal and verify
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	assert.NoError(err)
	assert.Equal(float64(2), result["count"])
	assert.NotNil(result["elems"])

	t.Logf("JSON: %s", string(data))
}

func Test_avutil_dict_marshal_json_empty(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	data, err := json.Marshal(dict)
	assert.NoError(err)
	assert.Contains(string(data), `"count":0`)
}

func Test_avutil_dict_entry_marshal_json(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "testkey", "testvalue", 0))

	entry := AVUtil_dict_get(dict, "testkey", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)

	// Marshal entry to JSON
	data, err := json.Marshal(entry)
	assert.NoError(err)
	assert.Contains(string(data), `"testkey"`)
	assert.Contains(string(data), `"testvalue"`)

	t.Logf("Entry JSON: %s", string(data))
}

////////////////////////////////////////////////////////////////////////////////
// TEST STRING OUTPUT

func Test_avutil_dict_string(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "name", "test", 0))
	assert.NoError(AVUtil_dict_set(dict, "version", "1.0", 0))

	str := dict.String()
	assert.NotEmpty(str)
	assert.Contains(str, "count")
	assert.Contains(str, "name")
	assert.Contains(str, "version")

	t.Logf("String output:\n%s", str)
}

func Test_avutil_dict_string_empty(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	str := dict.String()
	assert.NotEmpty(str)
	assert.Contains(str, `"count": 0`)
}

////////////////////////////////////////////////////////////////////////////////
// TEST SPECIAL CHARACTERS

func Test_avutil_dict_special_characters(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Test with special characters
	assert.NoError(AVUtil_dict_set(dict, "key with spaces", "value with spaces", 0))
	assert.NoError(AVUtil_dict_set(dict, "key=with=equals", "value=with=equals", 0))
	assert.NoError(AVUtil_dict_set(dict, "unicode_🔥", "emoji_value_✨", 0))

	assert.Equal(3, AVUtil_dict_count(dict))

	// Verify retrieval
	entry := AVUtil_dict_get(dict, "key with spaces", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)
	assert.Equal("value with spaces", entry.Value())

	entry = AVUtil_dict_get(dict, "unicode_🔥", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)
	assert.Equal("emoji_value_✨", entry.Value())
}

func Test_avutil_dict_empty_values(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Set key with empty value
	assert.NoError(AVUtil_dict_set(dict, "empty_key", "", 0))
	assert.Equal(1, AVUtil_dict_count(dict))

	entry := AVUtil_dict_get(dict, "empty_key", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)
	assert.Equal("", entry.Value())
}

////////////////////////////////////////////////////////////////////////////////
// TEST LARGE DICTIONARIES

func Test_avutil_dict_large(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Add many entries
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		assert.NoError(AVUtil_dict_set(dict, key, value, 0))
	}

	assert.Equal(10000, AVUtil_dict_count(dict))

	// Verify some entries
	entry := AVUtil_dict_get(dict, "key_5000", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)
	assert.Equal("value_5000", entry.Value())
}

func Test_avutil_dict_iteration(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	assert.NoError(AVUtil_dict_set(dict, "first", "1", 0))
	assert.NoError(AVUtil_dict_set(dict, "second", "2", 0))
	assert.NoError(AVUtil_dict_set(dict, "third", "3", 0))

	// Iterate through all entries
	var keys []string
	entry := AVUtil_dict_get(dict, "", nil, AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		keys = append(keys, entry.Key())
		entry = AVUtil_dict_get(dict, "", entry, AV_DICT_IGNORE_SUFFIX)
	}

	assert.Equal(3, len(keys))
	assert.Contains(keys, "first")
	assert.Contains(keys, "second")
	assert.Contains(keys, "third")
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avutil_dict_long_values(t *testing.T) {
	assert := assert.New(t)

	dict := AVUtil_dict_alloc()
	defer AVUtil_dict_free(dict)

	// Test with very long value
	longValue := strings.Repeat("x", 10000)
	assert.NoError(AVUtil_dict_set(dict, "long_key", longValue, 0))

	entry := AVUtil_dict_get(dict, "long_key", nil, AV_DICT_MATCH_CASE)
	assert.NotNil(entry)
	assert.Equal(longValue, entry.Value())
	assert.Equal(10000, len(entry.Value()))
}
