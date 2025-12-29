package ffmpeg

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TEST STRING OUTPUT

func Test_avutil_log_string(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		level    AVLog
		expected string
	}{
		{AV_LOG_QUIET, "QUIET"},
		{AV_LOG_PANIC, "PANIC"},
		{AV_LOG_FATAL, "FATAL"},
		{AV_LOG_ERROR, "ERROR"},
		{AV_LOG_WARNING, "WARN"},
		{AV_LOG_INFO, "INFO"},
		{AV_LOG_VERBOSE, "VERBOSE"},
		{AV_LOG_DEBUG, "DEBUG"},
		{AV_LOG_TRACE, "TRACE"},
	}

	for _, tc := range tests {
		str := tc.level.String()
		assert.Equal(tc.expected, str)
		t.Logf("Log level %v: %q", tc.level, str)
	}
}

func Test_avutil_log_string_invalid(t *testing.T) {
	assert := assert.New(t)

	// Test with invalid log level value
	invalidLevel := AVLog(999)
	str := invalidLevel.String()
	assert.NotEmpty(str)
	assert.Contains(str, "AVLog")
	assert.Contains(str, "999")
	t.Logf("Invalid log level string: %q", str)
}

////////////////////////////////////////////////////////////////////////////////
// TEST JSON MARSHALING

func Test_avutil_log_json_marshaling(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		level    AVLog
		expected string
	}{
		{AV_LOG_QUIET, `"QUIET"`},
		{AV_LOG_PANIC, `"PANIC"`},
		{AV_LOG_FATAL, `"FATAL"`},
		{AV_LOG_ERROR, `"ERROR"`},
		{AV_LOG_WARNING, `"WARN"`},
		{AV_LOG_INFO, `"INFO"`},
		{AV_LOG_VERBOSE, `"VERBOSE"`},
		{AV_LOG_DEBUG, `"DEBUG"`},
		{AV_LOG_TRACE, `"TRACE"`},
	}

	for _, tc := range tests {
		jsonBytes, err := json.Marshal(tc.level)
		assert.NoError(err)
		assert.Equal(tc.expected, string(jsonBytes))
		t.Logf("Log level %v marshals to: %s", tc.level, string(jsonBytes))
	}
}

func Test_avutil_log_json_in_struct(t *testing.T) {
	assert := assert.New(t)

	type LogEntry struct {
		Level   AVLog  `json:"level"`
		Message string `json:"message"`
	}

	entry := LogEntry{
		Level:   AV_LOG_ERROR,
		Message: "Test error message",
	}

	jsonBytes, err := json.Marshal(entry)
	assert.NoError(err)
	expected := `{"level":"ERROR","message":"Test error message"}`
	assert.Equal(expected, string(jsonBytes))
	t.Logf("Log entry JSON: %s", string(jsonBytes))
}

////////////////////////////////////////////////////////////////////////////////
// TEST LOG LEVEL GET/SET

func Test_avutil_log_level_operations(t *testing.T) {
	assert := assert.New(t)

	// Save original level to restore later
	originalLevel := AVUtil_log_get_level()
	defer AVUtil_log_set_level(originalLevel)

	levels := []AVLog{
		AV_LOG_QUIET,
		AV_LOG_PANIC,
		AV_LOG_FATAL,
		AV_LOG_ERROR,
		AV_LOG_WARNING,
		AV_LOG_INFO,
		AV_LOG_VERBOSE,
		AV_LOG_DEBUG,
		AV_LOG_TRACE,
	}

	for _, level := range levels {
		AVUtil_log_set_level(level)
		retrievedLevel := AVUtil_log_get_level()
		assert.Equal(level, retrievedLevel)
		t.Logf("Set log level to %s (%d), retrieved: %s (%d)", level, level, retrievedLevel, retrievedLevel)
	}
}

////////////////////////////////////////////////////////////////////////////////
// TEST LOGGING BASIC

func Test_avutil_log_basic(t *testing.T) {
	assert := assert.New(t)

	// Save and restore original level
	originalLevel := AVUtil_log_get_level()
	defer AVUtil_log_set_level(originalLevel)

	// Set log level to TRACE to see all messages
	AVUtil_log_set_level(AV_LOG_TRACE)
	assert.Equal(AV_LOG_TRACE, AVUtil_log_get_level())

	// Log messages at different levels
	AVUtil_log(nil, AV_LOG_TRACE, "This is a trace message\n")
	AVUtil_log(nil, AV_LOG_DEBUG, "This is a debug message\n")
	AVUtil_log(nil, AV_LOG_VERBOSE, "This is a verbose message\n")
	AVUtil_log(nil, AV_LOG_INFO, "This is an info message\n")
	AVUtil_log(nil, AV_LOG_WARNING, "This is a warning message\n")
	AVUtil_log(nil, AV_LOG_ERROR, "This is an error message\n")
	AVUtil_log(nil, AV_LOG_FATAL, "This is a fatal message\n")
	AVUtil_log(nil, AV_LOG_PANIC, "This is a panic message\n")

	t.Log("All log messages sent successfully")
}

func Test_avutil_log_with_formatting(t *testing.T) {
	// Save and restore original level
	originalLevel := AVUtil_log_get_level()
	defer AVUtil_log_set_level(originalLevel)

	AVUtil_log_set_level(AV_LOG_INFO)

	// Test formatted messages
	AVUtil_log(nil, AV_LOG_INFO, "Integer: %d\n", 42)
	AVUtil_log(nil, AV_LOG_INFO, "String: %s\n", "hello")
	AVUtil_log(nil, AV_LOG_INFO, "Float: %.2f\n", 3.14159)
	AVUtil_log(nil, AV_LOG_INFO, "Multiple: %s %d %.2f\n", "test", 123, 45.67)

	t.Log("Formatted log messages sent successfully")
}

////////////////////////////////////////////////////////////////////////////////
// TEST LOG CALLBACK

func Test_avutil_log_callback(t *testing.T) {
	// Save and restore original level and callback
	originalLevel := AVUtil_log_get_level()
	defer AVUtil_log_set_level(originalLevel)
	defer AVUtil_log_set_callback(nil) // Restore default callback

	// Set log level to ERROR
	AVUtil_log_set_level(AV_LOG_ERROR)

	// Track received messages
	var receivedMessages []string
	var mu sync.Mutex

	// Set custom log callback
	AVUtil_log_set_callback(func(level AVLog, message string, userInfo any) {
		mu.Lock()
		defer mu.Unlock()
		receivedMessages = append(receivedMessages, message)
		t.Logf("Callback received - Level=%v, Message=%q, userInfo=%v", level, strings.TrimSpace(message), userInfo)
	})

	// These should NOT trigger callback (below ERROR level)
	AVUtil_log(nil, AV_LOG_TRACE, "Trace message\n")
	AVUtil_log(nil, AV_LOG_DEBUG, "Debug message\n")
	AVUtil_log(nil, AV_LOG_VERBOSE, "Verbose message\n")
	AVUtil_log(nil, AV_LOG_INFO, "Info message\n")
	AVUtil_log(nil, AV_LOG_WARNING, "Warning message\n")

	// These SHOULD trigger callback (ERROR level and above)
	AVUtil_log(nil, AV_LOG_ERROR, "Error message\n")
	AVUtil_log(nil, AV_LOG_FATAL, "Fatal message\n")
	AVUtil_log(nil, AV_LOG_PANIC, "Panic message\n")

	mu.Lock()
	messageCount := len(receivedMessages)
	mu.Unlock()

	// We should have received 3 messages (ERROR, FATAL, PANIC)
	if messageCount != 3 {
		t.Errorf("Expected 3 messages at ERROR level and above, got %d", messageCount)
	}

	mu.Lock()
	for i, msg := range receivedMessages {
		t.Logf("Message %d: %q", i+1, strings.TrimSpace(msg))
	}
	mu.Unlock()
}

func Test_avutil_log_callback_reset(t *testing.T) {
	assert := assert.New(t)

	// Save original level
	originalLevel := AVUtil_log_get_level()
	defer AVUtil_log_set_level(originalLevel)

	AVUtil_log_set_level(AV_LOG_INFO)

	callbackCalled := false

	// Set callback
	AVUtil_log_set_callback(func(level AVLog, message string, userInfo any) {
		callbackCalled = true
	})

	// Log a message - callback should be called
	AVUtil_log(nil, AV_LOG_INFO, "Test message with callback\n")
	assert.True(callbackCalled, "Callback should have been called")

	// Reset to default callback
	callbackCalled = false
	AVUtil_log_set_callback(nil)

	// Log another message - our callback should NOT be called
	AVUtil_log(nil, AV_LOG_INFO, "Test message without callback\n")
	assert.False(callbackCalled, "Callback should not be called after reset")
}

////////////////////////////////////////////////////////////////////////////////
// TEST LOG CONSTANTS ORDERING

func Test_avutil_log_constants_ordering(t *testing.T) {
	assert := assert.New(t)

	// Log levels should be ordered from least to most verbose
	// QUIET < PANIC < FATAL < ERROR < WARNING < INFO < VERBOSE < DEBUG < TRACE

	t.Logf("AV_LOG_QUIET = %d", int(AV_LOG_QUIET))
	t.Logf("AV_LOG_PANIC = %d", int(AV_LOG_PANIC))
	t.Logf("AV_LOG_FATAL = %d", int(AV_LOG_FATAL))
	t.Logf("AV_LOG_ERROR = %d", int(AV_LOG_ERROR))
	t.Logf("AV_LOG_WARNING = %d", int(AV_LOG_WARNING))
	t.Logf("AV_LOG_INFO = %d", int(AV_LOG_INFO))
	t.Logf("AV_LOG_VERBOSE = %d", int(AV_LOG_VERBOSE))
	t.Logf("AV_LOG_DEBUG = %d", int(AV_LOG_DEBUG))
	t.Logf("AV_LOG_TRACE = %d", int(AV_LOG_TRACE))

	// Verify ordering (more negative/lower values = less verbose)
	assert.True(int(AV_LOG_QUIET) < int(AV_LOG_PANIC))
	assert.True(int(AV_LOG_PANIC) < int(AV_LOG_FATAL))
	assert.True(int(AV_LOG_FATAL) < int(AV_LOG_ERROR))
	assert.True(int(AV_LOG_ERROR) < int(AV_LOG_WARNING))
	assert.True(int(AV_LOG_WARNING) < int(AV_LOG_INFO))
	assert.True(int(AV_LOG_INFO) < int(AV_LOG_VERBOSE))
	assert.True(int(AV_LOG_VERBOSE) < int(AV_LOG_DEBUG))
	assert.True(int(AV_LOG_DEBUG) < int(AV_LOG_TRACE))
}

////////////////////////////////////////////////////////////////////////////////
// TEST LOG LEVEL FILTERING

func Test_avutil_log_level_filtering(t *testing.T) {
	assert := assert.New(t)

	// Save and restore
	originalLevel := AVUtil_log_get_level()
	defer AVUtil_log_set_level(originalLevel)
	defer AVUtil_log_set_callback(nil)

	testCases := []struct {
		setLevel      AVLog
		logLevel      AVLog
		shouldReceive bool
		description   string
	}{
		{AV_LOG_ERROR, AV_LOG_TRACE, false, "TRACE when level is ERROR"},
		{AV_LOG_ERROR, AV_LOG_DEBUG, false, "DEBUG when level is ERROR"},
		{AV_LOG_ERROR, AV_LOG_INFO, false, "INFO when level is ERROR"},
		{AV_LOG_ERROR, AV_LOG_WARNING, false, "WARNING when level is ERROR"},
		{AV_LOG_ERROR, AV_LOG_ERROR, true, "ERROR when level is ERROR"},
		{AV_LOG_ERROR, AV_LOG_FATAL, true, "FATAL when level is ERROR"},
		{AV_LOG_ERROR, AV_LOG_PANIC, true, "PANIC when level is ERROR"},
		{AV_LOG_INFO, AV_LOG_DEBUG, false, "DEBUG when level is INFO"},
		{AV_LOG_INFO, AV_LOG_INFO, true, "INFO when level is INFO"},
		{AV_LOG_INFO, AV_LOG_WARNING, true, "WARNING when level is INFO"},
		{AV_LOG_INFO, AV_LOG_ERROR, true, "ERROR when level is INFO"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			received := false

			AVUtil_log_set_level(tc.setLevel)
			AVUtil_log_set_callback(func(level AVLog, message string, userInfo any) {
				received = true
			})

			AVUtil_log(nil, tc.logLevel, "Test message\n")

			assert.Equal(tc.shouldReceive, received, "Message reception mismatch for %s", tc.description)
			t.Logf("Set level: %s, Log level: %s, Received: %v (expected: %v)",
				tc.setLevel, tc.logLevel, received, tc.shouldReceive)

			// Reset callback for next test
			AVUtil_log_set_callback(nil)
		})
	}
}
