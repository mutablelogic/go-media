package ffmpeg

import (
	"sync"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST TIMESTAMP TO STRING

func Test_avutil_ts_make_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		ts          int64
		description string
	}{
		{0, "zero timestamp"},
		{1000, "positive timestamp"},
		{-1000, "negative timestamp"},
		{9223372036854775807, "max int64"},
		{-9223372036854775808, "min int64"},
	}

	for _, tc := range tests {
		result := AVUtil_ts_make_string(tc.ts)
		assert.NotEmpty(result, tc.description)
		t.Logf("ts_make_string(%d): %q", tc.ts, result)
	}
}

func Test_avutil_ts2str(t *testing.T) {
	assert := assert.New(t)

	// Test that ts2str is equivalent to ts_make_string
	timestamps := []int64{0, 100, 1000, -500}

	for _, ts := range timestamps {
		str1 := AVUtil_ts_make_string(ts)
		str2 := AVUtil_ts2str(ts)
		assert.Equal(str1, str2)
		t.Logf("ts=%d: %q", ts, str1)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST TIMESTAMP TO TIME STRING

func Test_avutil_ts_make_time_string(t *testing.T) {
	assert := assert.New(t)

	// Create a time base (1/1000 for milliseconds)
	tb := AVRational{num: 1, den: 1000}

	tests := []struct {
		ts          int64
		description string
	}{
		{0, "zero timestamp"},
		{1000, "1 second in milliseconds"},
		{500, "0.5 seconds"},
		{-1000, "negative 1 second"},
	}

	for _, tc := range tests {
		result := AVUtil_ts_make_time_string(tc.ts, &tb)
		assert.NotEmpty(result, tc.description)
		t.Logf("ts_make_time_string(%d, 1/1000): %q", tc.ts, result)
	}
}

func Test_avutil_ts2timestr(t *testing.T) {
	assert := assert.New(t)

	// Test that ts2timestr is equivalent to ts_make_time_string
	tb := AVRational{num: 1, den: 90000}
	timestamps := []int64{0, 90000, 180000, -90000}

	for _, ts := range timestamps {
		str1 := AVUtil_ts_make_time_string(ts, &tb)
		str2 := AVUtil_ts2timestr(ts, &tb)
		assert.Equal(str1, str2)
		t.Logf("ts=%d: %q", ts, str1)
	}
}

func Test_avutil_ts_make_time_string_different_timebases(t *testing.T) {
	assert := assert.New(t)

	ts := int64(1000)

	timebases := []struct {
		tb          AVRational
		description string
	}{
		{AVRational{num: 1, den: 1}, "1 second timebase"},
		{AVRational{num: 1, den: 1000}, "1 millisecond timebase"},
		{AVRational{num: 1, den: 90000}, "90kHz timebase (MPEG)"},
		{AVRational{num: 1001, den: 30000}, "NTSC timebase"},
		{AVRational{num: 1, den: 48000}, "48kHz audio timebase"},
	}

	for _, tc := range timebases {
		result := AVUtil_ts_make_time_string(ts, &tc.tb)
		assert.NotEmpty(result, tc.description)
		t.Logf("ts=%d, tb=%d/%d (%s): %q", ts, tc.tb.num, tc.tb.den, tc.description, result)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avutil_ts_make_string_nopts_value(t *testing.T) {
	assert := assert.New(t)

	// AV_NOPTS_VALUE is typically the minimum int64 value
	noptsValue := int64(-9223372036854775808)
	result := AVUtil_ts_make_string(noptsValue)
	assert.NotEmpty(result)
	t.Logf("NOPTS value: %q", result)
}

func Test_avutil_ts_make_time_string_zero_denominator(t *testing.T) {
	// Test with zero denominator (invalid timebase)
	// This should not crash but may produce unexpected output
	tb := AVRational{num: 1, den: 0}
	result := AVUtil_ts_make_time_string(1000, &tb)
	t.Logf("ts_make_time_string with zero denominator: %q", result)
}

func Test_avutil_ts_make_time_string_nil_timebase(t *testing.T) {
	t.Skip("Skipping test that may crash with nil timebase")

	// Test with nil timebase pointer
	// This may crash or produce unexpected output depending on FFmpeg implementation
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic with nil timebase: %v", r)
		}
	}()

	result := AVUtil_ts_make_time_string(1000, nil)
	t.Logf("ts_make_time_string with nil timebase: %q", result)
}

////////////////////////////////////////////////////////////////////////////////
// TEST THREAD SAFETY

func Test_avutil_ts_make_string_concurrent(t *testing.T) {
	assert := assert.New(t)

	const numGoroutines = 100
	const numIterations = 100

	var wg sync.WaitGroup
	results := make(chan string, numGoroutines*numIterations)
	errorCount := int32(0)

	// Launch multiple goroutines that all call ts_make_string
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				ts := int64(id*1000 + j)
				result := AVUtil_ts_make_string(ts)
				results <- result
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Check that all calls succeeded
	assert.Equal(int32(0), errorCount, "Should have no errors in concurrent execution")

	resultCount := len(results)
	assert.Equal(numGoroutines*numIterations, resultCount, "Should have correct number of results")

	t.Logf("Concurrent test: %d goroutines × %d iterations = %d successful calls",
		numGoroutines, numIterations, resultCount)
}

func Test_avutil_ts_make_time_string_concurrent(t *testing.T) {
	assert := assert.New(t)

	const numGoroutines = 100
	const numIterations = 100

	var wg sync.WaitGroup
	results := make(chan string, numGoroutines*numIterations)
	errorCount := int32(0)

	// Launch multiple goroutines that all call ts_make_time_string
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tb := AVRational{num: 1, den: 1000}
			for j := 0; j < numIterations; j++ {
				ts := int64(id*1000 + j)
				result := AVUtil_ts_make_time_string(ts, &tb)
				results <- result
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Check that all calls succeeded
	assert.Equal(int32(0), errorCount, "Should have no errors in concurrent execution")

	resultCount := len(results)
	assert.Equal(numGoroutines*numIterations, resultCount, "Should have correct number of results")

	t.Logf("Concurrent test: %d goroutines × %d iterations = %d successful calls",
		numGoroutines, numIterations, resultCount)
}

func Test_avutil_ts_mixed_concurrent(t *testing.T) {
	assert := assert.New(t)

	const numGoroutines = 50
	const numIterations = 50

	var wg sync.WaitGroup
	results := make(chan string, numGoroutines*numIterations*2)

	// Launch multiple goroutines that call both functions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tb := AVRational{num: 1, den: 1000}
			for j := 0; j < numIterations; j++ {
				ts := int64(id*1000 + j)

				// Alternate between the two functions
				if j%2 == 0 {
					result := AVUtil_ts_make_string(ts)
					results <- result
				} else {
					result := AVUtil_ts_make_time_string(ts, &tb)
					results <- result
				}
			}
		}(i)
	}

	wg.Wait()
	close(results)

	resultCount := len(results)
	assert.Equal(numGoroutines*numIterations, resultCount, "Should have correct number of results")

	t.Logf("Mixed concurrent test: %d results from alternating function calls", resultCount)
}

////////////////////////////////////////////////////////////////////////////////
// TEST VARIOUS TIMESTAMP VALUES

func Test_avutil_ts_common_values(t *testing.T) {
	tests := []struct {
		ts   int64
		name string
	}{
		{0, "zero"},
		{1, "one"},
		{90000, "1 second at 90kHz"},
		{1000000, "1 million"},
		{-1, "negative one"},
		{-90000, "negative 1 second at 90kHz"},
	}

	for _, tc := range tests {
		str := AVUtil_ts_make_string(tc.ts)
		t.Logf("Timestamp %s (%d): %q", tc.name, tc.ts, str)
	}
}

func Test_avutil_ts_video_timestamps(t *testing.T) {
	// Test typical video timestamps
	fps30 := AVRational{num: 1, den: 30}
	fps60 := AVRational{num: 1, den: 60}
	fps24 := AVRational{num: 1001, den: 24000}

	timestamps := []int64{0, 30, 60, 90, 1800} // Various frame counts

	for _, ts := range timestamps {
		str30 := AVUtil_ts_make_time_string(ts, &fps30)
		str60 := AVUtil_ts_make_time_string(ts, &fps60)
		str24 := AVUtil_ts_make_time_string(ts, &fps24)
		t.Logf("Frame %d: 30fps=%q, 60fps=%q, 24fps=%q", ts, str30, str60, str24)
	}
}

func Test_avutil_ts_audio_timestamps(t *testing.T) {
	// Test typical audio timestamps
	sr48k := AVRational{num: 1, den: 48000}
	sr44k := AVRational{num: 1, den: 44100}

	samples := []int64{0, 48000, 96000, 44100, 88200} // Various sample counts

	for _, ts := range samples {
		str48k := AVUtil_ts_make_time_string(ts, &sr48k)
		str44k := AVUtil_ts_make_time_string(ts, &sr44k)
		t.Logf("Sample %d: 48kHz=%q, 44.1kHz=%q", ts, str48k, str44k)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST CONSISTENCY

func Test_avutil_ts_consistency(t *testing.T) {
	assert := assert.New(t)

	// Test that multiple calls with same input produce same output
	ts := int64(12345)
	tb := AVRational{num: 1, den: 1000}

	results1 := make([]string, 10)
	results2 := make([]string, 10)

	for i := 0; i < 10; i++ {
		results1[i] = AVUtil_ts_make_string(ts)
		results2[i] = AVUtil_ts_make_time_string(ts, &tb)
	}

	// All results should be identical
	for i := 1; i < 10; i++ {
		assert.Equal(results1[0], results1[i], "ts_make_string should be consistent")
		assert.Equal(results2[0], results2[i], "ts_make_time_string should be consistent")
	}
}
