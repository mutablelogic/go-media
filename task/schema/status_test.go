package schema_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	// Packages
	uuid "github.com/google/uuid"
	schema "github.com/mutablelogic/go-media/task/schema"
)

func TestStatus_MarshalJSON_OmitsNilErr(t *testing.T) {
	status := schema.Status{UUID: uuid.New(), Name: "probe"}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if _, exists := got["error"]; exists {
		t.Fatalf("expected \"error\" to be omitted, got %v", got["error"])
	}
}

func TestStatus_MarshalJSON_RendersErrAsString(t *testing.T) {
	status := schema.Status{UUID: uuid.New(), Name: "probe", Err: errors.New("boom")}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got["error"] != "boom" {
		t.Fatalf("error = %v, want %q", got["error"], "boom")
	}
}

// TestStatus_JSONRoundTrip guards against the bug where MarshalJSON rendered
// Err as a string but there was no matching UnmarshalJSON - encoding/json's
// default unmarshaling can't populate a bare `error` interface field, so any
// consumer decoding a Status with a non-nil Err (e.g. the httpclient reading
// task/httphandler's SSE stream) got a decode error, not a Status.
func TestStatus_JSONRoundTrip(t *testing.T) {
	want := schema.Status{
		UUID:      uuid.New(),
		Name:      "probe",
		Progress:  schema.Progress{Current: 1, Total: 2},
		Started:   time.Now().Add(-time.Second).Truncate(time.Second),
		Finished:  time.Now().Truncate(time.Second),
		Cancelled: true,
		Err:       errors.New("context canceled"),
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}

	var got schema.Status
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.UUID != want.UUID {
		t.Errorf("UUID = %v, want %v", got.UUID, want.UUID)
	}
	if got.Name != want.Name {
		t.Errorf("Name = %v, want %v", got.Name, want.Name)
	}
	if got.Progress != want.Progress {
		t.Errorf("Progress = %+v, want %+v", got.Progress, want.Progress)
	}
	if !got.Started.Equal(want.Started) {
		t.Errorf("Started = %v, want %v", got.Started, want.Started)
	}
	if !got.Finished.Equal(want.Finished) {
		t.Errorf("Finished = %v, want %v", got.Finished, want.Finished)
	}
	if got.Cancelled != want.Cancelled {
		t.Errorf("Cancelled = %v, want %v", got.Cancelled, want.Cancelled)
	}
	if got.Err == nil || got.Err.Error() != want.Err.Error() {
		t.Errorf("Err = %v, want message %q", got.Err, want.Err.Error())
	}
	if got.State() != schema.StateCancelled {
		t.Errorf("State() = %v, want %v", got.State(), schema.StateCancelled)
	}
}

func TestStatus_UnmarshalJSON_NoErrorField(t *testing.T) {
	var got schema.Status
	if err := json.Unmarshal([]byte(`{"uuid":"`+uuid.New().String()+`","name":"probe"}`), &got); err != nil {
		t.Fatal(err)
	}
	if got.Err != nil {
		t.Fatalf("Err = %v, want nil", got.Err)
	}
}
