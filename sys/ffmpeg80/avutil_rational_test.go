package ffmpeg

import (
	"encoding/json"
	"math"
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST RATIONAL CREATION

func Test_avutil_rational_create(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(3, 4)
	assert.Equal(3, r.Num())
	assert.Equal(4, r.Den())
	assert.False(r.IsZero())
}

func Test_avutil_rational_create_zero(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(0, 1)
	assert.Equal(0, r.Num())
	assert.Equal(1, r.Den())
	assert.True(r.IsZero())
}

func Test_avutil_rational_create_negative(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(-3, 4)
	assert.Equal(-3, r.Num())
	assert.Equal(4, r.Den())
	assert.False(r.IsZero())
}

////////////////////////////////////////////////////////////////////////////////
// TEST PROPERTIES

func Test_avutil_rational_num_den(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		num int
		den int
	}{
		{1, 2},
		{3, 4},
		{16, 9},
		{1920, 1080},
		{-5, 3},
	}

	for _, tc := range tests {
		r := AVUtil_rational(tc.num, tc.den)
		assert.Equal(tc.num, r.Num())
		assert.Equal(tc.den, r.Den())
		t.Logf("Rational: %d/%d", r.Num(), r.Den())
	}
}

func Test_avutil_rational_is_zero(t *testing.T) {
	assert := assert.New(t)

	zero := AVUtil_rational(0, 1)
	assert.True(zero.IsZero())

	nonZero := AVUtil_rational(1, 2)
	assert.False(nonZero.IsZero())

	negativeZero := AVUtil_rational(0, 100)
	assert.True(negativeZero.IsZero())
}

func Test_avutil_rational_float(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(1, 2) // 0.5
	result := r.Float(100)
	assert.Equal(50.0, result)

	r = AVUtil_rational(3, 4) // 0.75
	result = r.Float(100)
	assert.Equal(75.0, result)

	r = AVUtil_rational(1, 1000)
	result = r.Float(1000000)
	assert.Equal(1000.0, result)
}

func Test_avutil_rational_float_divide_by_zero(t *testing.T) {
	assert := assert.New(t)

	// Test division by zero handling
	r := AVUtil_rational(5, 0)
	result := r.Float(100)
	assert.Equal(0.0, result, "Float with zero denominator should return 0")
}

////////////////////////////////////////////////////////////////////////////////
// TEST CONVERSION

func Test_avutil_rational_d2q(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		value    float64
		max      int
		expected string
	}{
		{0.5, 0, "1/2"},
		{0.75, 0, "3/4"},
		{0.25, 0, "1/4"},
		{1.5, 0, "3/2"},
		{2.5, 0, "5/2"},
	}

	for _, tc := range tests {
		r := AVUtil_rational_d2q(tc.value, tc.max)
		assert.NotNil(r)
		assert.False(r.IsZero())
		t.Logf("d2q(%f, %d) = %d/%d", tc.value, tc.max, r.Num(), r.Den())
	}
}

func Test_avutil_rational_d2q_zero(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational_d2q(0.0, 0)
	assert.True(r.IsZero())
}

func Test_avutil_rational_q2d(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		num      int
		den      int
		expected float64
	}{
		{1, 2, 0.5},
		{3, 4, 0.75},
		{1, 4, 0.25},
		{3, 2, 1.5},
		{5, 2, 2.5},
		{0, 1, 0.0},
	}

	for _, tc := range tests {
		r := AVUtil_rational(tc.num, tc.den)
		result := AVUtil_rational_q2d(r)
		assert.InDelta(tc.expected, result, 0.0001)
		t.Logf("q2d(%d/%d) = %f", tc.num, tc.den, result)
	}
}

func Test_avutil_rational_d2q_and_q2d_roundtrip(t *testing.T) {
	assert := assert.New(t)

	values := []float64{0.5, 0.25, 0.75, 1.5, 2.5, 0.333333}

	for _, val := range values {
		r := AVUtil_rational_d2q(val, 0)
		result := AVUtil_rational_q2d(r)
		assert.InDelta(val, result, 0.01, "Roundtrip failed for %f", val)
		t.Logf("Roundtrip: %f -> %d/%d -> %f", val, r.Num(), r.Den(), result)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST COMPARISON

func Test_avutil_rational_equal(t *testing.T) {
	assert := assert.New(t)

	r1 := AVUtil_rational(1, 2)
	r2 := AVUtil_rational(1, 2)
	assert.True(AVUtil_rational_equal(r1, r2))

	r3 := AVUtil_rational(2, 4) // Equivalent to 1/2
	assert.True(AVUtil_rational_equal(r1, r3))
}

func Test_avutil_rational_not_equal(t *testing.T) {
	assert := assert.New(t)

	r1 := AVUtil_rational(1, 2)
	r2 := AVUtil_rational(1, 3)
	assert.False(AVUtil_rational_equal(r1, r2))

	r3 := AVUtil_rational(3, 4)
	assert.False(AVUtil_rational_equal(r1, r3))
}

func Test_avutil_rational_equal_zero(t *testing.T) {
	assert := assert.New(t)

	r1 := AVUtil_rational(0, 1)
	r2 := AVUtil_rational(0, 100)
	assert.True(AVUtil_rational_equal(r1, r2))
}

////////////////////////////////////////////////////////////////////////////////
// TEST INVERT

func Test_avutil_rational_invert(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(3, 4)
	inv := AVUtil_rational_invert(r)
	assert.Equal(4, inv.Num())
	assert.Equal(3, inv.Den())
	t.Logf("Invert %d/%d = %d/%d", r.Num(), r.Den(), inv.Num(), inv.Den())
}

func Test_avutil_rational_invert_twice(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(5, 7)
	inv := AVUtil_rational_invert(r)
	invInv := AVUtil_rational_invert(inv)

	// Should get back to original (might be normalized)
	assert.True(AVUtil_rational_equal(r, invInv))
}

func Test_avutil_rational_invert_one(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(1, 1)
	inv := AVUtil_rational_invert(r)
	assert.True(AVUtil_rational_equal(r, inv))
}

////////////////////////////////////////////////////////////////////////////////
// TEST RESCALE

func Test_avutil_rational_rescale_q(t *testing.T) {
	assert := assert.New(t)

	// Example: Convert 1 second in milliseconds to microseconds
	// 1000 ms * (1000000 µs / 1 s) / (1000 ms / 1 s) = 1000000 µs
	a := int64(1000)                  // 1000 milliseconds
	bq := AVUtil_rational(1000000, 1) // microseconds per second
	cq := AVUtil_rational(1000, 1)    // milliseconds per second

	result := AVUtil_rational_rescale_q(a, bq, cq)
	assert.Equal(int64(1000000), result)
	t.Logf("Rescale: %d with %d/%d to %d/%d = %d", a, bq.Num(), bq.Den(), cq.Num(), cq.Den(), result)
}

func Test_avutil_rational_rescale_q_timebase(t *testing.T) {
	assert := assert.New(t)

	// Convert from one timebase to another
	pts := int64(90000)                 // PTS in timebase 1/90000
	src_tb := AVUtil_rational(1, 90000) // Source timebase
	dst_tb := AVUtil_rational(1, 1000)  // Destination timebase

	result := AVUtil_rational_rescale_q(pts, src_tb, dst_tb)
	assert.Equal(int64(1000), result) // Should be 1000 in 1/1000 timebase
	t.Logf("Timebase conversion: %d * (%d/%d) / (%d/%d) = %d",
		pts, src_tb.Num(), src_tb.Den(), dst_tb.Num(), dst_tb.Den(), result)
}

func Test_avutil_rational_rescale_q_zero(t *testing.T) {
	assert := assert.New(t)

	result := AVUtil_rational_rescale_q(0, AVUtil_rational(1, 1000), AVUtil_rational(1, 1))
	assert.Equal(int64(0), result)
}

////////////////////////////////////////////////////////////////////////////////
// TEST RESCALE WITH ROUNDING

func Test_avutil_rescale_rnd_zero(t *testing.T) {
	assert := assert.New(t)

	result := AVUtil_rescale_rnd(150, 1, 100, AV_ROUND_ZERO)
	assert.Equal(int64(1), result)
	t.Logf("ROUND_ZERO: 150 * 1 / 100 = %d", result)
}

func Test_avutil_rescale_rnd_inf(t *testing.T) {
	assert := assert.New(t)

	result := AVUtil_rescale_rnd(150, 1, 100, AV_ROUND_INF)
	assert.Equal(int64(2), result)
	t.Logf("ROUND_INF: 150 * 1 / 100 = %d", result)
}

func Test_avutil_rescale_rnd_down(t *testing.T) {
	assert := assert.New(t)

	result := AVUtil_rescale_rnd(150, 1, 100, AV_ROUND_DOWN)
	assert.Equal(int64(1), result)
	t.Logf("ROUND_DOWN: 150 * 1 / 100 = %d", result)
}

func Test_avutil_rescale_rnd_up(t *testing.T) {
	assert := assert.New(t)

	result := AVUtil_rescale_rnd(150, 1, 100, AV_ROUND_UP)
	assert.Equal(int64(2), result)
	t.Logf("ROUND_UP: 150 * 1 / 100 = %d", result)
}

func Test_avutil_rescale_rnd_near_inf(t *testing.T) {
	assert := assert.New(t)

	// Test rounding to nearest
	result1 := AVUtil_rescale_rnd(140, 1, 100, AV_ROUND_NEAR_INF)
	assert.Equal(int64(1), result1)

	result2 := AVUtil_rescale_rnd(160, 1, 100, AV_ROUND_NEAR_INF)
	assert.Equal(int64(2), result2)

	t.Logf("ROUND_NEAR_INF: 140/100=%d, 160/100=%d", result1, result2)
}

func Test_avutil_rescale_rnd_various_modes(t *testing.T) {
	assert := assert.New(t)

	modes := []AVRounding{
		AV_ROUND_ZERO,
		AV_ROUND_INF,
		AV_ROUND_DOWN,
		AV_ROUND_UP,
		AV_ROUND_NEAR_INF,
	}

	a := int64(355) // 3.55 when divided by 100
	b := int64(1)
	c := int64(100)

	for _, mode := range modes {
		result := AVUtil_rescale_rnd(a, b, c, mode)
		t.Logf("Rescale 355*1/100 with mode %d = %d", mode, result)
		assert.Greater(result, int64(0))
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST COMPARE TIMESTAMPS

func Test_avutil_compare_ts_equal(t *testing.T) {
	assert := assert.New(t)

	// Same timestamp in same timebase
	a := int64(1000)
	a_tb := AVUtil_rational(1, 1000)
	b := int64(1000)
	b_tb := AVUtil_rational(1, 1000)

	result := AVUtil_compare_ts(a, a_tb, b, b_tb)
	assert.Equal(0, result)
}

func Test_avutil_compare_ts_before(t *testing.T) {
	assert := assert.New(t)

	// a is before b
	a := int64(500)
	a_tb := AVUtil_rational(1, 1000)
	b := int64(1000)
	b_tb := AVUtil_rational(1, 1000)

	result := AVUtil_compare_ts(a, a_tb, b, b_tb)
	assert.Equal(-1, result)
}

func Test_avutil_compare_ts_after(t *testing.T) {
	assert := assert.New(t)

	// a is after b
	a := int64(1500)
	a_tb := AVUtil_rational(1, 1000)
	b := int64(1000)
	b_tb := AVUtil_rational(1, 1000)

	result := AVUtil_compare_ts(a, a_tb, b, b_tb)
	assert.Equal(1, result)
}

func Test_avutil_compare_ts_different_timebases(t *testing.T) {
	assert := assert.New(t)

	// Same actual time in different timebases
	// 1000 in 1/1000 timebase = 1 second
	// 90000 in 1/90000 timebase = 1 second
	a := int64(1000)
	a_tb := AVUtil_rational(1, 1000)
	b := int64(90000)
	b_tb := AVUtil_rational(1, 90000)

	result := AVUtil_compare_ts(a, a_tb, b, b_tb)
	assert.Equal(0, result, "Same time in different timebases should be equal")
}

func Test_avutil_compare_ts_different_timebases_before(t *testing.T) {
	assert := assert.New(t)

	// 500ms vs 1000ms
	a := int64(500)
	a_tb := AVUtil_rational(1, 1000)
	b := int64(90000)
	b_tb := AVUtil_rational(1, 90000)

	result := AVUtil_compare_ts(a, a_tb, b, b_tb)
	assert.Equal(-1, result)
}

func Test_avutil_compare_ts_zero(t *testing.T) {
	assert := assert.New(t)

	a := int64(0)
	a_tb := AVUtil_rational(1, 1000)
	b := int64(0)
	b_tb := AVUtil_rational(1, 90000)

	result := AVUtil_compare_ts(a, a_tb, b, b_tb)
	assert.Equal(0, result)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avutil_rational_marshal_json(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(3, 4)
	data, err := json.Marshal(r)
	assert.NoError(err)
	assert.Equal(`"3/4"`, string(data))
	t.Logf("JSON: %s", string(data))
}

func Test_avutil_rational_marshal_json_zero(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(0, 1)
	data, err := json.Marshal(r)
	assert.NoError(err)
	assert.Equal("0", string(data))
	t.Logf("Zero JSON: %s", string(data))
}

func Test_avutil_rational_marshal_json_negative(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(-5, 3)
	data, err := json.Marshal(r)
	assert.NoError(err)
	assert.Equal(`"-5/3"`, string(data))
	t.Logf("Negative JSON: %s", string(data))
}

func Test_avutil_rational_marshal_json_various(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		num      int
		den      int
		expected string
	}{
		{1, 2, `"1/2"`},
		{16, 9, `"16/9"`},
		{1920, 1080, `"1920/1080"`},
		{0, 100, "0"},
		{-1, 4, `"-1/4"`},
	}

	for _, tc := range tests {
		r := AVUtil_rational(tc.num, tc.den)
		data, err := json.Marshal(r)
		assert.NoError(err)
		assert.Equal(tc.expected, string(data))
		t.Logf("%d/%d -> %s", tc.num, tc.den, string(data))
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST STRING

func Test_avutil_rational_string(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(16, 9)
	str := r.String()
	assert.NotEmpty(str)
	assert.Contains(str, "16/9")
	t.Logf("String: %s", str)
}

func Test_avutil_rational_string_zero(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(0, 1)
	str := r.String()
	assert.NotEmpty(str)
	assert.Contains(str, "0")
	t.Logf("Zero string: %s", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST EDGE CASES

func Test_avutil_rational_large_values(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational(1920*1080, 60)
	assert.Equal(1920*1080, r.Num())
	assert.Equal(60, r.Den())
	assert.False(r.IsZero())
}

func Test_avutil_rational_very_small_float(t *testing.T) {
	assert := assert.New(t)

	r := AVUtil_rational_d2q(0.00001, 1000000)
	result := AVUtil_rational_q2d(r)
	assert.InDelta(0.00001, result, 0.000001)
}

func Test_avutil_rational_very_large_rescale(t *testing.T) {
	assert := assert.New(t)

	// Large timestamp value
	pts := int64(1000000000)
	src_tb := AVUtil_rational(1, 90000)
	dst_tb := AVUtil_rational(1, 1000)

	result := AVUtil_rational_rescale_q(pts, src_tb, dst_tb)
	assert.Greater(result, int64(0))
	t.Logf("Large rescale: %d -> %d", pts, result)
}

func Test_avutil_rational_negative_rescale(t *testing.T) {
	assert := assert.New(t)

	// Negative values
	pts := int64(-1000)
	src_tb := AVUtil_rational(1, 1000)
	dst_tb := AVUtil_rational(1, 1)

	result := AVUtil_rational_rescale_q(pts, src_tb, dst_tb)
	assert.Less(result, int64(0))
	t.Logf("Negative rescale: %d -> %d", pts, result)
}

func Test_avutil_rational_precision(t *testing.T) {
	assert := assert.New(t)

	// Test precision with different max values
	value := math.Pi

	r1 := AVUtil_rational_d2q(value, 10)
	r2 := AVUtil_rational_d2q(value, 100)
	r3 := AVUtil_rational_d2q(value, 1000)

	result1 := AVUtil_rational_q2d(r1)
	result2 := AVUtil_rational_q2d(r2)
	result3 := AVUtil_rational_q2d(r3)

	t.Logf("Pi with max=10:   %d/%d = %f (diff: %f)", r1.Num(), r1.Den(), result1, math.Abs(value-result1))
	t.Logf("Pi with max=100:  %d/%d = %f (diff: %f)", r2.Num(), r2.Den(), result2, math.Abs(value-result2))
	t.Logf("Pi with max=1000: %d/%d = %f (diff: %f)", r3.Num(), r3.Den(), result3, math.Abs(value-result3))

	// Higher precision should be closer to actual value
	assert.True(math.Abs(value-result3) <= math.Abs(value-result1))
}
