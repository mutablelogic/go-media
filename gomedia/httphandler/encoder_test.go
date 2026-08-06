package httphandler_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	// Packages
	test "github.com/mutablelogic/go-media/gomedia/test"
	profile "github.com/mutablelogic/go-media/profile/schema"
	taskencoder "github.com/mutablelogic/go-media/task/encoder"
	taskschema "github.com/mutablelogic/go-media/task/schema"
	require "github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func sampleFilePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "etc", "test", name)
}

// aacRequestJSON is the JSON encoding of an EncodeRequest's non-Reader
// fields (Output, Audio, ...) - what a client would send as the "request"
// form field.
func aacRequestJSON(t *testing.T) string {
	t.Helper()

	output := profile.OutputWithName("mp4")
	require.NotNil(t, output)

	audio, err := profile.NewAudioProfile("aac")
	require.NoError(t, err)
	require.NoError(t, audio.Set(profile.OptionSampleRate, uint64(44100)))
	require.NoError(t, audio.Set(profile.OptionSampleFormat, "fltp"))
	require.NoError(t, audio.Set(profile.OptionChannelLayout, "stereo"))

	req := taskencoder.EncodeRequest{Output: output, Audio: audio}
	data, err := json.Marshal(req)
	require.NoError(t, err)
	return string(data)
}

// postEncode uploads sample.mp3 with the given request JSON and Accept
// header, returning the raw HTTP response.
func postEncode(t *testing.T, requestJSON, accept string) *http.Response {
	t.Helper()

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	require.NoError(t, err)
	defer f.Close()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	fw, err := w.CreateFormFile("file", "sample.mp3")
	require.NoError(t, err)
	_, err = io.Copy(fw, f)
	require.NoError(t, err)

	require.NoError(t, w.WriteField("request", requestJSON))
	require.NoError(t, w.Close())

	req, err := http.NewRequest(http.MethodPost, test.ServerURL(t)+"/encode", &body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

// TestEncodeHandler_JSON checks the default (non text/event-stream) response:
// the task's status, returned promptly, not the finished result.
func TestEncodeHandler_JSON(t *testing.T) {
	require := require.New(t)
	test.Begin(t)
	defer test.End(t)

	resp := postEncode(t, aacRequestJSON(t), "")
	defer resp.Body.Close()

	require.Equal(http.StatusAccepted, resp.StatusCode)
	require.Equal("application/json", strings.Split(resp.Header.Get("Content-Type"), ";")[0])

	var status taskschema.Status
	require.NoError(json.NewDecoder(resp.Body).Decode(&status))
	require.NotZero(status.UUID)
	require.Equal("encoder", status.Task)
}

// TestEncodeHandler_JSON_FileSurvivesAsyncRead is a regression test: Encode
// doesn't wait for the task to finish, so for the default (JSON) response
// this handler returns almost immediately - well before the task's own
// goroutine has necessarily even started reading the uploaded file. A
// defer tied to the handler's own return would close (and for a large
// upload spooled to disk, delete) that file out from under the still-running
// task. This drives the request through the real HTTP handler and then
// waits for the task itself to finish, to prove the file lived long enough.
func TestEncodeHandler_JSON_FileSurvivesAsyncRead(t *testing.T) {
	require := require.New(t)
	test.Begin(t)
	defer test.End(t)

	resp := postEncode(t, aacRequestJSON(t), "")
	defer resp.Body.Close()
	require.Equal(http.StatusAccepted, resp.StatusCode)

	var status taskschema.Status
	require.NoError(json.NewDecoder(resp.Body).Decode(&status))

	final, err := test.TaskManager(t).Wait(context.Background(), status.UUID)
	require.NoError(err)
	require.Equal(taskschema.StateDone, final.State())

	result, ok := final.Result.(*taskencoder.EncodeResponse)
	require.True(ok, "expected *taskencoder.EncodeResponse, got %T", final.Result)
	require.NotEmpty(result.Path)
	defer os.Remove(result.Path)
}

// TestEncodeHandler_EventStream checks the text/event-stream response:
// the task's own events, streamed until it finishes.
func TestEncodeHandler_EventStream(t *testing.T) {
	require := require.New(t)
	test.Begin(t)
	defer test.End(t)

	resp := postEncode(t, aacRequestJSON(t), "text/event-stream")
	defer resp.Body.Close()

	require.Equal(http.StatusOK, resp.StatusCode)
	require.Equal("text/event-stream", strings.Split(resp.Header.Get("Content-Type"), ";")[0])

	// Read "event: <name>\ndata: <json>\n\n" blocks until "finished" arrives,
	// or the test's own deadline.
	scanner := bufio.NewScanner(resp.Body)
	var eventName string
	var finished *taskschema.Status
	deadline := time.Now().Add(10 * time.Second)
	for finished == nil && time.Now().Before(deadline) && scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			eventName = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			if eventName != "finished" {
				continue
			}
			var status taskschema.Status
			require.NoError(json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &status))
			finished = &status
		}
	}

	require.NotNil(finished, "did not observe a finished event")
	require.Empty(finished.Err)
	require.Equal(taskschema.StateDone, finished.State())

	result, ok := finished.Result.(map[string]any)
	require.True(ok, "expected a result object, got %T", finished.Result)
	path, _ := result["path"].(string)
	require.NotEmpty(path)
	defer os.Remove(path)
}

func TestEncodeHandler_NoProfile(t *testing.T) {
	require := require.New(t)
	test.Begin(t)
	defer test.End(t)

	output := profile.OutputWithName("mp4")
	require.NotNil(output)
	data, err := json.Marshal(taskencoder.EncodeRequest{Output: output})
	require.NoError(err)

	resp := postEncode(t, string(data), "")
	defer resp.Body.Close()

	// Media.Encode validates before creating the task, so a malformed
	// request is rejected with a 400 rather than accepted as a task that's
	// doomed to fail asynchronously.
	require.Equal(http.StatusBadRequest, resp.StatusCode)
}
